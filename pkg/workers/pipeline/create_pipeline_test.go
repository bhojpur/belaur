package pipeline

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
	"errors"
	"io/ioutil"
	"log"
	"strings"
	"testing"

	belaur "github.com/bhojpur/belaur"
	"github.com/bhojpur/belaur/pkg/services"
	"github.com/bhojpur/belaur/pkg/store"
	hclog "github.com/hashicorp/go-hclog"
)

type mockCreatePipelineStore struct {
	store.BelaurStore
	Error error
}

func (mcp *mockCreatePipelineStore) CreatePipelinePut(p *belaur.CreatePipeline) error {
	return mcp.Error
}

// PipelinePut is a Mock implementation for pipelines
func (mcp *mockCreatePipelineStore) PipelinePut(p *belaur.Pipeline) error {
	return mcp.Error
}

type mockScheduler struct {
	Error error
}

func (ms *mockScheduler) Init() {}
func (ms *mockScheduler) SchedulePipeline(p *belaur.Pipeline, startedBy string, args []*belaur.Argument) (*belaur.PipelineRun, error) {
	return nil, nil
}
func (ms *mockScheduler) SetPipelineJobs(p *belaur.Pipeline) error            { return ms.Error }
func (ms *mockScheduler) StopPipelineRun(p *belaur.Pipeline, runid int) error { return ms.Error }
func (ms *mockScheduler) GetFreeWorkers() int32                               { return int32(0) }
func (ms *mockScheduler) CountScheduledRuns() int                             { return 0 }

func TestCreatePipelineUnknownType(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestCreatePipelineUnknownType")
	belaur.Cfg = new(belaur.Config)
	belaur.Cfg.HomePath = tmp
	buf := new(bytes.Buffer)
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Belaur",
	})
	mcp := new(mockCreatePipelineStore)
	services.MockStorageService(mcp)
	defer func() { services.MockStorageService(nil) }()
	cp := new(belaur.CreatePipeline)
	cp.Pipeline.Type = belaur.PTypeUnknown
	pipelineService := NewBelaurPipelineService(Dependencies{
		Scheduler: &mockScheduleService{},
	})
	pipelineService.CreatePipeline(cp)
	if cp.Output != "create pipeline failed. Pipeline type is not supported unknown is not supported" {
		t.Fatal("error output was not the expected output. was: ", cp.Output)
	}
	if cp.StatusType != belaur.CreatePipelineFailed {
		t.Fatal("pipeline status is not expected status. was:", cp.StatusType)
	}
}

func TestCreatePipelineMissingGitURL(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestCreatePipelineMissingGitURL")
	belaur.Cfg = new(belaur.Config)
	belaur.Cfg.HomePath = tmp
	buf := new(bytes.Buffer)
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Belaur",
	})
	mcp := new(mockCreatePipelineStore)
	services.MockStorageService(mcp)
	defer func() { services.MockStorageService(nil) }()
	cp := new(belaur.CreatePipeline)
	cp.Pipeline.Type = belaur.PTypeGolang
	pipelineService := NewBelaurPipelineService(Dependencies{
		Scheduler: &mockScheduleService{},
	})
	pipelineService.CreatePipeline(cp)
	if cp.Output != "cannot prepare build: URL field is required" {
		t.Fatal("output was not what was expected. was: ", cp.Output)
	}
}

func TestCreatePipelineFailedToUpdatePipeline(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestCreatePipelineFailedToUpdatePipeline")
	belaur.Cfg = new(belaur.Config)
	belaur.Cfg.HomePath = tmp
	buf := new(bytes.Buffer)
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Belaur",
	})
	mcp := new(mockCreatePipelineStore)
	mcp.Error = errors.New("failed")
	services.MockStorageService(mcp)
	defer func() { services.MockStorageService(nil) }()
	cp := new(belaur.CreatePipeline)
	cp.Pipeline.Type = belaur.PTypeGolang
	cp.Pipeline.Repo = &belaur.GitRepo{URL: "https://github.com/bhojpur/pipeline-test"}
	pipelineService := NewBelaurPipelineService(Dependencies{
		Scheduler: &mockScheduleService{},
	})
	pipelineService.CreatePipeline(cp)
	body, _ := ioutil.ReadAll(buf)
	if !bytes.Contains(body, []byte("cannot put create pipeline into store: error=failed")) {
		t.Fatal("expected log message was not there. was: ", string(body))
	}
}

func TestCreatePipeline(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestCreatePipeline")
	belaur.Cfg = new(belaur.Config)
	belaur.Cfg.HomePath = tmp
	belaur.Cfg.PipelinePath = tmp
	buf := new(bytes.Buffer)
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Belaur",
	})
	mcp := new(mockCreatePipelineStore)
	services.MockStorageService(mcp)
	defer func() { services.MockStorageService(nil) }()
	cp := new(belaur.CreatePipeline)
	cp.Pipeline.Name = "test"
	cp.Pipeline.ID = 1
	cp.Pipeline.Type = belaur.PTypeGolang
	cp.Pipeline.Repo = &belaur.GitRepo{URL: "https://github.com/bhojpur/pipeline-test"}
	pipelineService := NewBelaurPipelineService(Dependencies{
		Scheduler: &mockScheduleService{},
	})
	pipelineService.CreatePipeline(cp)
	if cp.StatusType != belaur.CreatePipelineSuccess {
		log.Println("Output: ", cp.Output)
		t.Fatal("pipeline status was not success. was: ", cp.StatusType)
	}
}

func TestCreatePipelineSetPipelineJobsFail(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestCreatePipelineSetPipelineJobsFail")
	belaur.Cfg = new(belaur.Config)
	belaur.Cfg.HomePath = tmp
	belaur.Cfg.PipelinePath = tmp
	buf := new(bytes.Buffer)
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Belaur",
	})
	mcp := new(mockCreatePipelineStore)
	services.MockStorageService(mcp)
	defer func() { services.MockStorageService(nil) }()
	cp := new(belaur.CreatePipeline)
	cp.Pipeline.Name = "test"
	cp.Pipeline.Type = belaur.PTypeGolang
	cp.Pipeline.Repo = &belaur.GitRepo{URL: "https://github.com/bhojpur/pipeline-test"}
	pipelineService := NewBelaurPipelineService(Dependencies{
		Scheduler: &mockScheduleService{
			err: errors.New("error"),
		},
	})
	pipelineService.CreatePipeline(cp)
	if !strings.Contains(cp.Output, "cannot validate pipeline") {
		t.Fatalf("error thrown should contain 'cannot validate pipeline' but its %s", cp.Output)
	}
}
