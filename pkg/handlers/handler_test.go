package handlers

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
	"io/ioutil"
	"os"
	"testing"
	"time"

	rbacProvider "github.com/bhojpur/belaur/pkg/providers/rbac"
	userProvider "github.com/bhojpur/belaur/pkg/providers/user"
	"github.com/bhojpur/belaur/pkg/security/rbac"

	belaur "github.com/bhojpur/belaur"
	"github.com/bhojpur/belaur/pkg/providers/pipelines"
	"github.com/bhojpur/belaur/pkg/providers/workers"
	"github.com/bhojpur/belaur/pkg/security"
	"github.com/bhojpur/belaur/pkg/store"
	"github.com/bhojpur/belaur/pkg/workers/pipeline"

	"github.com/hashicorp/go-hclog"
	"github.com/labstack/echo/v4"
)

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

func TestInitHandler(t *testing.T) {
	dataDir, err := ioutil.TempDir("", "TestInitHandler")
	if err != nil {
		t.Fatalf("error creating data dir %v", err.Error())
	}

	defer func() {
		belaur.Cfg = nil
		_ = os.RemoveAll(dataDir)
	}()
	belaur.Cfg = &belaur.Config{
		Logger:    hclog.NewNullLogger(),
		DataPath:  dataDir,
		CAPath:    dataDir,
		VaultPath: dataDir,
		HomePath:  dataDir,
		Mode:      belaur.ModeServer,
		DevMode:   true,
	}
	e := echo.New()

	// Initialize global active pipelines
	ap := pipeline.NewActivePipelines()
	pipeline.GlobalActivePipelines = ap

	p := belaur.Pipeline{
		ID:      1,
		Name:    "Pipeline A",
		Type:    belaur.PTypeGolang,
		Created: time.Now(),
		Repo: &belaur.GitRepo{
			URL: "https://github.com/Codertocat/Hello-World",
		},
	}

	ap.Append(p)
	ms := &mockScheduleService{}
	// Initialize handlers
	pipelineService := pipeline.NewBelaurPipelineService(pipeline.Dependencies{
		Scheduler: ms,
	})
	pp := pipelines.NewPipelineProvider(pipelines.Dependencies{
		Scheduler:       ms,
		PipelineService: pipelineService,
	})
	ca, err := security.InitCA()
	if err != nil {
		belaur.Cfg.Logger.Error("cannot create CA", "error", err.Error())
		return
	}
	wp := workers.NewWorkerProvider(workers.Dependencies{
		Certificate: ca,
		Scheduler:   ms,
	})
	mStore := &mockStorageService{mockPipeline: &p}
	rbacService := rbac.NewNoOpService()
	rbacPrv := rbacProvider.NewProvider(rbacService)
	userPrv := userProvider.NewProvider(mStore, rbacService)
	handlerService := NewBelaurHandler(Dependencies{
		Scheduler:        ms,
		PipelineService:  pipelineService,
		PipelineProvider: pp,
		Certificate:      ca,
		WorkerProvider:   wp,
		Store:            mStore,
		UserProvider:     userPrv,
		RBACProvider:     rbacPrv,
	})
	if err := handlerService.InitHandlers(e); err != nil {
		t.Fatal(err)
	}
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
