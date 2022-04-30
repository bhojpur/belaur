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
	"io/ioutil"
	"os"
	"testing"
	"time"

	belaur "github.com/bhojpur/belaur"
	"github.com/bhojpur/belaur/pkg/services"
	"github.com/bhojpur/belaur/pkg/store/memdb"
	"github.com/bhojpur/belaur/pkg/workers/scheduler/service"
	"github.com/hashicorp/go-hclog"
)

type mockScheduleService struct {
	service.BelaurScheduler
	err error
}

func (ms *mockScheduleService) SetPipelineJobs(p *belaur.Pipeline) error {
	return ms.err
}

type mockMemDBService struct {
	worker    *belaur.Worker
	setWorker *belaur.Worker
	memdb.BelaurMemDB
}

func (mm *mockMemDBService) GetAllWorker() []*belaur.Worker {
	return []*belaur.Worker{mm.setWorker}
}
func (mm *mockMemDBService) UpsertWorker(w *belaur.Worker, persist bool) error {
	mm.worker = w
	return nil
}

//
func TestCheckActivePipelines(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestCheckActivePipelines")
	dataDir := tmp

	belaur.Cfg = &belaur.Config{
		Logger:       hclog.NewNullLogger(),
		DataPath:     dataDir,
		HomePath:     dataDir,
		PipelinePath: dataDir,
	}
	belaur.Cfg.Bolt.Mode = 0600

	// Initialize store
	dataStore, err := services.StorageService()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { services.MockStorageService(nil) }()
	// Initialize global active pipelines
	ap := NewActivePipelines()
	GlobalActivePipelines = ap

	pipeline1 := belaur.Pipeline{
		ID:      1,
		Name:    "testpipe",
		Type:    belaur.PTypeGolang,
		Created: time.Now(),
	}

	// Create fake binary
	src := GetExecPath(pipeline1)
	f, _ := os.Create(src)
	defer f.Close()
	defer os.Remove(src)

	// Manually run check
	pipelineService := NewBelaurPipelineService(Dependencies{
		Scheduler: &mockScheduleService{},
	})
	pipelineService.CheckActivePipelines()

	// Check if pipeline was added to store
	_, err = dataStore.PipelineGet(pipeline1.ID)
	if err != nil {
		t.Error("cannot find pipeline in store")
	}
}

func TestTurningThePollerOn(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestTurningThePollerOn")
	dataDir := tmp

	belaur.Cfg = &belaur.Config{
		Logger:       hclog.NewNullLogger(),
		DataPath:     dataDir,
		HomePath:     dataDir,
		PipelinePath: dataDir,
		Poll:         true,
	}
	pipelineService := NewBelaurPipelineService(Dependencies{
		Scheduler: &mockScheduleService{},
	})
	defer pipelineService.StopPoller()
	err := pipelineService.StartPoller()
	if err != nil {
		t.Fatal("error was not expected. got: ", err)
	}
}

func TestTurningThePollerOnWhilePollingIsDisabled(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestTurningThePollerOnWhilePollingIsDisabled")
	dataDir := tmp

	belaur.Cfg = &belaur.Config{
		Logger:       hclog.NewNullLogger(),
		DataPath:     dataDir,
		HomePath:     dataDir,
		PipelinePath: dataDir,
		Poll:         false,
	}
	pipelineService := NewBelaurPipelineService(Dependencies{
		Scheduler: &mockScheduleService{},
	})
	err := pipelineService.StartPoller()
	if err != nil {
		t.Fatal("error was not expected. got: ", err)
	}
	if isPollerRunning != false {
		t.Fatal("expected isPollerRunning to be false. was: ", isPollerRunning)
	}
}

func TestTurningThePollerOnWhilePollingIsEnabled(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestTurningThePollerOnWhilePollingIsEnabled")
	dataDir := tmp

	belaur.Cfg = &belaur.Config{
		Logger:       hclog.NewNullLogger(),
		DataPath:     dataDir,
		HomePath:     dataDir,
		PipelinePath: dataDir,
		Poll:         true,
	}
	pipelineService := NewBelaurPipelineService(Dependencies{
		Scheduler: &mockScheduleService{},
	})
	defer pipelineService.StopPoller()
	err := pipelineService.StartPoller()
	if err != nil {
		t.Fatal("error was not expected. got: ", err)
	}
	if isPollerRunning != true {
		t.Fatal("expected isPollerRunning to be true. was: ", isPollerRunning)
	}
}

func TestTurningThePollerOff(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestTurningThePollerOff")
	dataDir := tmp

	belaur.Cfg = &belaur.Config{
		Logger:       hclog.NewNullLogger(),
		DataPath:     dataDir,
		HomePath:     dataDir,
		PipelinePath: dataDir,
		Poll:         true,
	}
	pipelineService := NewBelaurPipelineService(Dependencies{
		Scheduler: &mockScheduleService{},
	})
	err := pipelineService.StartPoller()
	if err != nil {
		t.Fatal("error was not expected. got: ", err)
	}
	if isPollerRunning != true {
		t.Fatal("expected isPollerRunning to be true. was: ", isPollerRunning)
	}

	err = pipelineService.StopPoller()
	if err != nil {
		t.Fatal("error was not expected. got: ", err)
	}
	if isPollerRunning != false {
		t.Fatal("expected isPollerRunning to be false. was: ", isPollerRunning)
	}
}

func TestTogglePoller(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestTogglePoller")
	dataDir := tmp

	belaur.Cfg = &belaur.Config{
		Logger:       hclog.NewNullLogger(),
		DataPath:     dataDir,
		HomePath:     dataDir,
		PipelinePath: dataDir,
		Poll:         true,
	}
	pipelineService := NewBelaurPipelineService(Dependencies{
		Scheduler: &mockScheduleService{},
	})
	err := pipelineService.StartPoller()
	if err != nil {
		t.Fatal("error was not expected. got: ", err)
	}
	err = pipelineService.StartPoller()
	if err == nil {
		t.Fatal("starting the poller again should have failed")
	}
	err = pipelineService.StopPoller()
	if err != nil {
		t.Fatal("stopping the poller while it's running should not have failed. got: ", err)
	}
	err = pipelineService.StopPoller()
	if err == nil {
		t.Fatal("stopping the poller again while it's stopped should have failed.")
	}
}

func TestUpdateWorker(t *testing.T) {
	tmp, err := ioutil.TempDir("", "TestUpdateWorker")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)

	belaur.Cfg = &belaur.Config{
		Logger:       hclog.NewNullLogger(),
		DataPath:     tmp,
		HomePath:     tmp,
		PipelinePath: tmp,
	}

	db := &mockMemDBService{}
	services.MockMemDBService(db)
	defer func() { services.MockMemDBService(nil) }()
	db.setWorker = &belaur.Worker{
		Status:      belaur.WorkerActive,
		LastContact: time.Now().Add(-6 * time.Minute),
	}

	// Run update worker
	updateWorker()

	// Validate
	if db.worker == nil {
		t.Fatal("worker should not be nil")
	}
	if db.worker.Status != belaur.WorkerInactive {
		t.Fatalf("expected '%s' but got '%s'", string(belaur.WorkerInactive), string(db.worker.Status))
	}
	db.worker = nil

	// Set new test data
	db.setWorker = &belaur.Worker{
		Status:      belaur.WorkerInactive,
		LastContact: time.Now(),
	}

	// Run update worker
	updateWorker()

	// Validate
	if db.worker == nil {
		t.Fatal("worker should not be nil")
	}
	if db.worker.Status != belaur.WorkerActive {
		t.Fatalf("expected '%s' but got '%s'", string(belaur.WorkerActive), string(db.worker.Status))
	}
}
