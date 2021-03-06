package server

// Copyright (c) 2018 Bhojpur Consulting Private Limited, India. All rights reserved.

// Permission is hereby granted, free of charge, to any person obtaining a copy
// of this software and associated documentation files (the "Software"), to deal
// in the Software without restriction, including without limitation the rights
// to use, copy, modify, merge, publish, distribute, sublicense, and/or sell
// copies of the Software, and to permit persons to whom the Software is
// furnished to do so, subject to the following conditions:

// The above copyright notice and this permission notice shall be included in
// all copies or substantial portions of the Software.

// THE SOFTWARE IS PROVIDED "AS IS", WITHOUT WARRANTY OF ANY KIND, EXPRESS OR
// IMPLIED, INCLUDING BUT NOT LIMITED TO THE WARRANTIES OF MERCHANTABILITY,
// FITNESS FOR A PARTICULAR PURPOSE AND NONINFRINGEMENT. IN NO EVENT SHALL THE
// AUTHORS OR COPYRIGHT HOLDERS BE LIABLE FOR ANY CLAIM, DAMAGES OR OTHER
// LIABILITY, WHETHER IN AN ACTION OF CONTRACT, TORT OR OTHERWISE, ARISING FROM,
// OUT OF OR IN CONNECTION WITH THE SOFTWARE OR THE USE OR OTHER DEALINGS IN
// THE SOFTWARE.

import (
	"bytes"
	"context"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"

	belaur "github.com/bhojpur/belaur"
	pb "github.com/bhojpur/belaur/pkg/api/v1/worker"
	"github.com/bhojpur/belaur/pkg/services"
	"github.com/bhojpur/belaur/pkg/store"
	"github.com/bhojpur/belaur/pkg/store/memdb"
	"github.com/bhojpur/belaur/pkg/workers/pipeline"
	"github.com/golang/protobuf/ptypes/empty"
	hclog "github.com/hashicorp/go-hclog"
	"google.golang.org/grpc"
	"google.golang.org/grpc/metadata"
)

type mockMemDBService struct {
	memdb.BelaurMemDB
}

func (mm *mockMemDBService) GetWorker(id string) (*belaur.Worker, error) {
	return &belaur.Worker{UniqueID: "test-worker"}, nil
}
func (mm *mockMemDBService) UpsertWorker(w *belaur.Worker, persist bool) error { return nil }
func (mm *mockMemDBService) PopPipelineRun(tags []string) (*belaur.PipelineRun, error) {
	return generateTestData(), nil
}
func (mm *mockMemDBService) InsertPipelineRun(p *belaur.PipelineRun) error { return nil }
func (mm *mockMemDBService) DeleteWorker(id string, persist bool) error {
	if id != "my-worker" {
		return fmt.Errorf("expected 'my-worker' but got %s", id)
	}
	return nil
}

type mockStorageService struct {
	store.BelaurStore
	mockPipeline *belaur.Pipeline
}

func (s *mockStorageService) PipelineGetRunByPipelineIDAndID(pipelineid int, runid int) (*belaur.PipelineRun, error) {
	return generateTestData(), nil
}
func (s *mockStorageService) PipelinePutRun(r *belaur.PipelineRun) error { return nil }
func (s *mockStorageService) PipelineGet(id int) (pipeline *belaur.Pipeline, err error) {
	return s.mockPipeline, nil
}
func (s *mockStorageService) PipelineGetRunByID(runID string) (*belaur.PipelineRun, error) {
	return &belaur.PipelineRun{}, nil
}

type mockGetWorkServ struct {
	grpc.ServerStream
}

func (mw mockGetWorkServ) Send(pR *pb.PipelineRun) error {
	switch {
	case pR == nil:
		return fmt.Errorf("given pipeline run is nil")
	case pR.Id != 1:
		return fmt.Errorf("expected 1 but got %d", pR.Id)
	case pR.UniqueId != "first-pipeline-run":
		return fmt.Errorf("expected 'first-pipeline-run' but got %s", pR.UniqueId)
	}
	return nil
}
func (mw mockGetWorkServ) Context() context.Context {
	md := make(map[string]string)
	md["uniqueid"] = "my-unique-id"
	return metadata.NewIncomingContext(context.Background(), metadata.New(md))
}

type mockStreamBinaryServ struct {
	grpc.ServerStream
}

func (ms mockStreamBinaryServ) Send(c *pb.FileChunk) error {
	if !bytes.Equal(c.Chunk, []byte("test data")) {
		return fmt.Errorf("data send is not correct: %s", string(c.Chunk[:]))
	}
	return nil
}
func (ms mockStreamBinaryServ) Context() context.Context {
	md := make(map[string]string)
	md["uniqueid"] = "my-unique-id"
	return metadata.NewIncomingContext(context.Background(), metadata.New(md))
}

type mockStreamLogsServ struct {
	grpc.ServerStream
}

var counter = 0

func (ml mockStreamLogsServ) Recv() (*pb.LogChunk, error) {
	counter++

	if counter < 3 {
		return &pb.LogChunk{
			Chunk:      []byte("test log data"),
			PipelineId: 1,
			RunId:      1,
		}, nil
	}
	return nil, io.EOF
}
func (ml mockStreamLogsServ) Context() context.Context {
	md := make(map[string]string)
	md["uniqueid"] = "my-unique-id"
	return metadata.NewIncomingContext(context.Background(), metadata.New(md))
}
func (ml mockStreamLogsServ) SendAndClose(e *empty.Empty) error {
	return nil
}

func generateTestData() *belaur.PipelineRun {
	return &belaur.PipelineRun{
		UniqueID:   "first-pipeline-run",
		ID:         1,
		PipelineID: 1,
		Jobs: []*belaur.Job{
			{
				ID:     1,
				Title:  "first-job",
				Status: belaur.JobWaitingExec,
			},
		},
	}
}

func TestGetWorkServer(t *testing.T) {
	belaur.Cfg = &belaur.Config{
		Mode: belaur.ModeServer,
	}
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level: hclog.Trace,
		Name:  "Belaur",
	})
	services.MockMemDBService(&mockMemDBService{})
	services.MockStorageService(&mockStorageService{})

	// Init global active pipelines slice
	pipeline.GlobalActivePipelines = pipeline.NewActivePipelines()
	pipeline.GlobalActivePipelines.Append(belaur.Pipeline{ID: 1, SHA256Sum: []byte("testbytes"), Type: belaur.PTypeGolang, ExecPath: "execpath"})

	// Mock gRPC server
	mw := mockGetWorkServ{}

	// Run GetWork
	ws := WorkServer{}
	if err := ws.GetWork(&pb.WorkerInstance{UniqueId: "test", WorkerSlots: 1}, mw); err != nil {
		t.Fatal(err)
	}
}

func TestGetWorkWorker(t *testing.T) {
	belaur.Cfg = &belaur.Config{
		Mode: belaur.ModeWorker,
	}
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level: hclog.Trace,
		Name:  "Belaur",
	})
	services.MockMemDBService(&mockMemDBService{})
	services.MockStorageService(&mockStorageService{})

	// Init global active pipelines slice
	pipeline.GlobalActivePipelines = pipeline.NewActivePipelines()
	pipeline.GlobalActivePipelines.Append(belaur.Pipeline{ID: 1, SHA256Sum: []byte("testbytes"), Type: belaur.PTypeGolang, ExecPath: "execpath"})

	// Mock gRPC server
	mw := mockGetWorkServ{}

	// Run GetWork
	ws := WorkServer{}
	if err := ws.GetWork(&pb.WorkerInstance{UniqueId: "test", WorkerSlots: 1}, mw); err == nil {
		t.Fatal("expected error")
	}
}

func TestUpdateWork(t *testing.T) {
	belaur.Cfg = &belaur.Config{}
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level: hclog.Trace,
		Name:  "Belaur",
	})
	services.MockMemDBService(&mockMemDBService{})
	services.MockStorageService(&mockStorageService{})

	// Mock gRPC server
	mw := mockGetWorkServ{}

	t.Run("run-reschedule-success", func(t *testing.T) {
		pbRun := &pb.PipelineRun{
			UniqueId: "first-pipeline-run",
			Id:       1,
			Status:   string(belaur.RunReschedule),
			Docker:   true,
		}

		// Run UpdateWork
		ws := WorkServer{}
		if _, err := ws.UpdateWork(mw.Context(), pbRun); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("run-notscheduled-success", func(t *testing.T) {
		pbRun := &pb.PipelineRun{
			UniqueId: "first-pipeline-run",
			Id:       1,
			Status:   string(belaur.RunNotScheduled),
			Jobs: []*pb.Job{
				{
					UniqueId: 1,
					Title:    "first-job",
					Args: []*pb.Argument{
						{
							Key:         "key",
							Type:        "type",
							Description: "desc",
						},
					},
				},
			},
		}

		// Run UpdateWork
		ws := WorkServer{}
		if _, err := ws.UpdateWork(mw.Context(), pbRun); err != nil {
			t.Fatal(err)
		}
	})

	t.Run("run-success-success", func(t *testing.T) {
		pbRun := &pb.PipelineRun{
			UniqueId: "first-pipeline-run",
			Id:       1,
			Status:   string(belaur.RunSuccess),
			Jobs: []*pb.Job{
				{
					UniqueId: 1,
					Title:    "first-job",
					Args: []*pb.Argument{
						{
							Key:         "key",
							Type:        "type",
							Description: "desc",
						},
					},
				},
			},
		}

		// Run UpdateWork
		ws := WorkServer{}
		if _, err := ws.UpdateWork(mw.Context(), pbRun); err != nil {
			t.Fatal(err)
		}
	})
}

func TestStreamBinary(t *testing.T) {
	tmp, err := ioutil.TempDir("", "TestStreamBinary")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)
	belaur.Cfg = &belaur.Config{
		Mode: belaur.ModeServer,
	}
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level: hclog.Trace,
		Name:  "Belaur",
	})
	services.MockMemDBService(&mockMemDBService{})

	// Create test pipeline file
	testPipeline := filepath.Join(tmp, "my-pipeline_golang")
	if err := ioutil.WriteFile(testPipeline, []byte("test data"), 0777); err != nil {
		t.Fatal(err)
	}

	// Init global active pipelines slice
	pipeline.GlobalActivePipelines = pipeline.NewActivePipelines()
	pipeline.GlobalActivePipelines.Append(belaur.Pipeline{ID: 1, SHA256Sum: []byte("testbytes"), Type: belaur.PTypeGolang, ExecPath: testPipeline})

	// Mock gRPC server
	mw := mockStreamBinaryServ{}

	// Run StreamBinary
	ws := WorkServer{}
	if err := ws.StreamBinary(&pb.PipelineRun{Id: 1, PipelineId: 1}, mw); err != nil {
		t.Fatal(err)
	}
}

func TestStreamLogs(t *testing.T) {
	tmp, err := ioutil.TempDir("", "TestStreamLogs")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)
	belaur.Cfg = &belaur.Config{WorkspacePath: tmp}
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level: hclog.Trace,
		Name:  "Belaur",
	})
	services.MockMemDBService(&mockMemDBService{})

	// Mock gRPC server
	mw := mockStreamLogsServ{}

	// Run StreamLogs
	ws := WorkServer{}
	if err := ws.StreamLogs(mw); err != nil {
		t.Fatal(err)
	}

	// Validate output file
	logFolderPath := filepath.Join(belaur.Cfg.WorkspacePath, "1", "1", belaur.LogsFolderName)
	logFilePath := filepath.Join(logFolderPath, belaur.LogsFileName)

	content, err := ioutil.ReadFile(logFilePath)
	if err != nil {
		t.Fatal(err)
	}
	if !bytes.Equal(content, []byte("test log datatest log data")) {
		t.Fatalf("expected 'test log datatest log data' but got '%s'", string(content[:]))
	}
}

func TestDeregister(t *testing.T) {
	belaur.Cfg = &belaur.Config{}
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level: hclog.Trace,
		Name:  "Belaur",
	})
	services.MockMemDBService(&mockMemDBService{})
	services.MockStorageService(&mockStorageService{})

	// Mock gRPC server
	mw := mockGetWorkServ{}

	// Run deregister
	ws := WorkServer{}
	if _, err := ws.Deregister(mw.Context(), &pb.WorkerInstance{UniqueId: "my-worker"}); err != nil {
		t.Fatal(err)
	}
}

func TestGetGitRepository(t *testing.T) {
	belaur.Cfg = &belaur.Config{}
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level: hclog.Trace,
		Name:  "Belaur",
	})
	ms := mockStorageService{mockPipeline: &belaur.Pipeline{
		ID:   1,
		Name: "testPipeline",
		Repo: &belaur.GitRepo{
			URL: "https://github.com/bhojpur/go-example",
		},
	}}
	services.MockMemDBService(&mockMemDBService{})
	services.MockStorageService(&ms)

	// Mock gRPC server
	mw := mockGetWorkServ{}

	// Run deregister
	ws := WorkServer{}
	repo, err := ws.GetGitRepo(mw.Context(), &pb.PipelineID{Id: 1})
	if err != nil {
		t.Fatal(err)
	}
	expectedRepoURL := "https://github.com/bhojpur/go-example"
	if repo.Url != expectedRepoURL {
		t.Fatalf("expected git repo url: %s, got: %s\n", expectedRepoURL, repo.Url)
	}
}

func TestGetGitRepositoryRepoNotFound(t *testing.T) {
	belaur.Cfg = &belaur.Config{}
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level: hclog.Trace,
		Name:  "Belaur",
	})
	services.MockMemDBService(&mockMemDBService{})
	services.MockStorageService(&mockStorageService{})

	// Mock gRPC server
	mw := mockGetWorkServ{}

	// Run deregister
	ws := WorkServer{}
	_, err := ws.GetGitRepo(mw.Context(), &pb.PipelineID{Id: 9999})
	if err == nil {
		t.Fatal("should have gotten an error because pipeline doesn't exist")
	}

	expectedError := fmt.Sprintf("pipeline for id %d not found", 9999)
	if err.Error() != expectedError {
		t.Fatalf("expected error message: %s, got: %s", expectedError, err.Error())
	}
}
