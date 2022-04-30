package belaurscheduler

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
	"crypto/tls"
	"errors"
	"hash/fnv"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"sync"
	"testing"
	"time"

	belaur "github.com/bhojpur/belaur"
	"github.com/bhojpur/belaur/pkg/plugin"
	"github.com/bhojpur/belaur/pkg/security"
	"github.com/bhojpur/belaur/pkg/store"
	"github.com/bhojpur/belaur/pkg/workers/docker"
	"github.com/gofrs/uuid"
	"github.com/hashicorp/go-hclog"
)

type PluginFake struct{}

func (p *PluginFake) NewPlugin(ca security.CAAPI) plugin.Plugin { return &PluginFake{} }
func (p *PluginFake) Init(cmd *exec.Cmd, logPath *string) error { return nil }
func (p *PluginFake) Validate() error                           { return nil }
func (p *PluginFake) Execute(j *belaur.Job) error {
	j.Status = belaur.JobSuccess
	return nil
}
func (p *PluginFake) GetJobs() ([]*belaur.Job, error) { return prepareJobs(), nil }
func (p *PluginFake) FlushLogs() error                { return nil }
func (p *PluginFake) Close()                          {}

type CAFake struct{}

func (c *CAFake) CreateSignedCertWithValidOpts(hostname string, hoursBeforeValid, hoursAfterValid time.Duration) (string, string, error) {
	return "", "", nil
}
func (c *CAFake) CreateSignedCert() (string, string, error)                       { return "", "", nil }
func (c *CAFake) GenerateTLSConfig(certPath, keyPath string) (*tls.Config, error) { return nil, nil }
func (c *CAFake) CleanupCerts(crt, key string) error                              { return nil }
func (c *CAFake) GetCACertPath() (string, string)                                 { return "", "" }

type VaultFake struct{}

func (v *VaultFake) LoadSecrets() error             { return nil }
func (v *VaultFake) GetAll() []string               { return []string{} }
func (v *VaultFake) SaveSecrets() error             { return nil }
func (v *VaultFake) Add(key string, value []byte)   {}
func (v *VaultFake) Remove(key string)              {}
func (v *VaultFake) Get(key string) ([]byte, error) { return []byte{}, nil }

type MemDBFake struct{}

func (m *MemDBFake) SyncStore() error                                  { return nil }
func (m *MemDBFake) GetAllWorker() []*belaur.Worker                    { return []*belaur.Worker{} }
func (m *MemDBFake) UpsertWorker(w *belaur.Worker, persist bool) error { return nil }
func (m *MemDBFake) GetWorker(id string) (*belaur.Worker, error)       { return &belaur.Worker{}, nil }
func (m *MemDBFake) DeleteWorker(id string, persist bool) error        { return nil }
func (m *MemDBFake) InsertPipelineRun(p *belaur.PipelineRun) error     { return nil }
func (m *MemDBFake) PopPipelineRun(tags []string) (*belaur.PipelineRun, error) {
	return &belaur.PipelineRun{}, nil
}
func (m *MemDBFake) DeletePipelineRun(runID string) error { return nil }
func (m *MemDBFake) UpsertSHAPair(pair belaur.SHAPair) error {
	return nil
}
func (m *MemDBFake) GetSHAPair(pipelineID string) (ok bool, pair belaur.SHAPair, err error) {
	return
}
func (m *MemDBFake) InsertDockerWorker(w *docker.Worker) error { return nil }
func (m *MemDBFake) GetDockerWorker(workerID string) (*docker.Worker, error) {
	return &docker.Worker{}, nil
}
func (m *MemDBFake) DeleteDockerWorker(workerID string) error { return nil }
func (m *MemDBFake) GetAllDockerWorker() ([]*docker.Worker, error) {
	return []*docker.Worker{}, nil
}

func TestInit(t *testing.T) {
	belaur.Cfg = &belaur.Config{}
	storeInstance := store.NewBoltStore()
	tmp, _ := ioutil.TempDir("", "TestInit")
	belaur.Cfg.DataPath = tmp
	belaur.Cfg.WorkspacePath = filepath.Join(tmp, "tmp")
	belaur.Cfg.Bolt.Mode = 0600
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: hclog.DefaultOutput,
		Name:   "Belaur",
	})
	belaur.Cfg.Worker = 2
	if err := storeInstance.Init(tmp); err != nil {
		t.Fatal(err)
	}
	s, err := NewScheduler(Dependencies{storeInstance, &MemDBFake{}, &PluginFake{}, &CAFake{}, &VaultFake{}})
	if err != nil {
		t.Fatal(err)
	}
	s.Init()
}

type PluginFakeFailed struct{}

func (p *PluginFakeFailed) NewPlugin(ca security.CAAPI) plugin.Plugin { return &PluginFakeFailed{} }
func (p *PluginFakeFailed) Init(cmd *exec.Cmd, logPath *string) error { return nil }
func (p *PluginFakeFailed) Validate() error                           { return nil }
func (p *PluginFakeFailed) Execute(j *belaur.Job) error {
	j.Status = belaur.JobFailed
	j.FailPipeline = true
	return errors.New("job failed")
}
func (p *PluginFakeFailed) GetJobs() ([]*belaur.Job, error) { return prepareJobs(), nil }
func (p *PluginFakeFailed) FlushLogs() error                { return nil }
func (p *PluginFakeFailed) Close()                          {}

func TestPrepareAndExecFail(t *testing.T) {
	belaur.Cfg = &belaur.Config{}
	storeInstance := store.NewBoltStore()
	tmp, _ := ioutil.TempDir("", "TestPrepareAndExecFail")
	belaur.Cfg.DataPath = tmp
	belaur.Cfg.WorkspacePath = filepath.Join(tmp, "tmp")
	belaur.Cfg.Bolt.Mode = 0600
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: hclog.DefaultOutput,
		Name:   "Belaur",
	})

	if err := storeInstance.Init(tmp); err != nil {
		t.Fatal(err)
	}
	p, r := prepareTestData()
	_ = storeInstance.PipelinePut(&p)
	s, err := NewScheduler(Dependencies{storeInstance, &MemDBFake{}, &PluginFakeFailed{}, &CAFake{}, &VaultFake{}})
	if err != nil {
		t.Fatal(err)
	}
	s.prepareAndExec(r)

	// get pipeline run from store
	run, err := storeInstance.PipelineGetRunByPipelineIDAndID(p.ID, r.ID)
	if err != nil {
		t.Fatal(err)
	}

	// jobs should be existent
	if len(run.Jobs) == 0 {
		t.Fatal("No jobs in pipeline run found.")
	}

	// Check run status
	if run.Status != belaur.RunFailed {
		t.Fatalf("Run should be of type %s but was %s\n", belaur.RunFailed, run.Status)
	}
}

func TestPrepareAndExecInvalidType(t *testing.T) {
	belaur.Cfg = &belaur.Config{}
	storeInstance := store.NewBoltStore()
	tmp, _ := ioutil.TempDir("", "TestPrepareAndExec")
	belaur.Cfg.DataPath = tmp
	belaur.Cfg.WorkspacePath = filepath.Join(tmp, "tmp")
	belaur.Cfg.Bolt.Mode = 0600
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: hclog.DefaultOutput,
		Name:   "Belaur",
	})

	if err := storeInstance.Init(tmp); err != nil {
		t.Fatal(err)
	}
	p, r := prepareTestData()
	p.Type = belaur.PTypeUnknown
	_ = storeInstance.PipelinePut(&p)
	s, err := NewScheduler(Dependencies{storeInstance, &MemDBFake{}, &PluginFakeFailed{}, &CAFake{}, &VaultFake{}})
	if err != nil {
		t.Fatal(err)
	}
	s.prepareAndExec(r)

	// get pipeline run from store
	run, err := storeInstance.PipelineGetRunByPipelineIDAndID(p.ID, r.ID)
	if err != nil {
		t.Fatal(err)
	}

	// Check run status
	if run.Status != belaur.RunFailed {
		t.Fatalf("Run should be of type %s but was %s\n", belaur.RunFailed, run.Status)
	}
}

func TestPrepareAndExecJavaType(t *testing.T) {
	belaur.Cfg = &belaur.Config{}
	storeInstance := store.NewBoltStore()
	tmp, _ := ioutil.TempDir("", "TestPrepareAndExecJavaType")
	belaur.Cfg.DataPath = tmp
	belaur.Cfg.WorkspacePath = filepath.Join(tmp, "tmp")
	belaur.Cfg.Bolt.Mode = 0600
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: hclog.DefaultOutput,
		Name:   "Belaur",
	})

	if err := storeInstance.Init(tmp); err != nil {
		t.Fatal(err)
	}
	p, r := prepareTestData()
	javaExecName = "go"
	p.Type = belaur.PTypeJava
	_ = storeInstance.PipelinePut(&p)
	s, err := NewScheduler(Dependencies{storeInstance, &MemDBFake{}, &PluginFake{}, &CAFake{}, &VaultFake{}})
	if err != nil {
		t.Fatal(err)
	}
	s.prepareAndExec(r)

	// get pipeline run from store
	run, err := storeInstance.PipelineGetRunByPipelineIDAndID(p.ID, r.ID)
	if err != nil {
		t.Fatal(err)
	}

	// jobs should be existent
	if len(run.Jobs) == 0 {
		t.Fatal("No jobs in pipeline run found.")
	}

	// Iterate jobs
	for _, job := range run.Jobs {
		if job.Status != belaur.JobSuccess {
			t.Fatalf("job status should be success but was %s", string(job.Status))
		} else {
			t.Logf("Job %s has been executed...", job.Title)
		}
	}
}

func TestPrepareAndExecPythonType(t *testing.T) {
	belaur.Cfg = &belaur.Config{}
	storeInstance := store.NewBoltStore()
	tmp, _ := ioutil.TempDir("", "TestPrepareAndExecPythonType")
	belaur.Cfg.DataPath = tmp
	belaur.Cfg.WorkspacePath = filepath.Join(tmp, "tmp")
	belaur.Cfg.Bolt.Mode = 0600
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: hclog.DefaultOutput,
		Name:   "Belaur",
	})

	if err := storeInstance.Init(tmp); err != nil {
		t.Fatal(err)
	}
	p, r := prepareTestData()
	pythonExecName = "go"
	p.Type = belaur.PTypePython
	_ = storeInstance.PipelinePut(&p)
	s, err := NewScheduler(Dependencies{storeInstance, &MemDBFake{}, &PluginFake{}, &CAFake{}, &VaultFake{}})
	if err != nil {
		t.Fatal(err)
	}
	s.prepareAndExec(r)

	// get pipeline run from store
	run, err := storeInstance.PipelineGetRunByPipelineIDAndID(p.ID, r.ID)
	if err != nil {
		t.Fatal(err)
	}

	// jobs should be existent
	if len(run.Jobs) == 0 {
		t.Fatal("No jobs in pipeline run found.")
	}

	// Iterate jobs
	for _, job := range run.Jobs {
		if job.Status != belaur.JobSuccess {
			t.Fatalf("job status should be success but was %s", string(job.Status))
		} else {
			t.Logf("Job %s has been executed...", job.Title)
		}
	}
}

func TestPrepareAndExecCppType(t *testing.T) {
	belaur.Cfg = &belaur.Config{}
	storeInstance := store.NewBoltStore()
	tmp, _ := ioutil.TempDir("", "TestPrepareAndExecCppType")
	belaur.Cfg.DataPath = tmp
	belaur.Cfg.WorkspacePath = filepath.Join(tmp, "tmp")
	belaur.Cfg.Bolt.Mode = 0600
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: hclog.DefaultOutput,
		Name:   "Belaur",
	})

	if err := storeInstance.Init(tmp); err != nil {
		t.Fatal(err)
	}
	p, r := prepareTestData()
	p.Type = belaur.PTypeCpp
	_ = storeInstance.PipelinePut(&p)
	s, err := NewScheduler(Dependencies{storeInstance, &MemDBFake{}, &PluginFake{}, &CAFake{}, &VaultFake{}})
	if err != nil {
		t.Fatal(err)
	}
	s.prepareAndExec(r)

	// get pipeline run from store
	run, err := storeInstance.PipelineGetRunByPipelineIDAndID(p.ID, r.ID)
	if err != nil {
		t.Fatal(err)
	}

	// jobs should be existent
	if len(run.Jobs) == 0 {
		t.Fatal("No jobs in pipeline run found.")
	}

	// Iterate jobs
	for _, job := range run.Jobs {
		if job.Status != belaur.JobSuccess {
			t.Fatalf("job status should be success but was %s", string(job.Status))
		} else {
			t.Logf("Job %s has been executed...", job.Title)
		}
	}
}

func TestPrepareAndExecRubyType(t *testing.T) {
	belaur.Cfg = &belaur.Config{}
	storeInstance := store.NewBoltStore()
	tmp, _ := ioutil.TempDir("", "TestPrepareAndExecRubyType")
	belaur.Cfg.DataPath = tmp
	belaur.Cfg.WorkspacePath = filepath.Join(tmp, "tmp")
	belaur.Cfg.Bolt.Mode = 0600
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: hclog.DefaultOutput,
		Name:   "Belaur",
	})

	if err := storeInstance.Init(tmp); err != nil {
		t.Fatal(err)
	}
	p, r := prepareTestData()
	p.Type = belaur.PTypeRuby
	rubyExecName = "go"
	rubyGemName = "echo"
	findRubyGemCommands = []string{"name: rubytest"}
	_ = storeInstance.PipelinePut(&p)
	s, err := NewScheduler(Dependencies{storeInstance, &MemDBFake{}, &PluginFake{}, &CAFake{}, &VaultFake{}})
	if err != nil {
		t.Fatal(err)
	}
	s.prepareAndExec(r)

	// get pipeline run from store
	run, err := storeInstance.PipelineGetRunByPipelineIDAndID(p.ID, r.ID)
	if err != nil {
		t.Fatal(err)
	}

	// jobs should be existent
	if len(run.Jobs) == 0 {
		t.Fatal("No jobs in pipeline run found.")
	}

	// Iterate jobs
	for _, job := range run.Jobs {
		if job.Status != belaur.JobSuccess {
			t.Fatalf("job status should be success but was %s", string(job.Status))
		} else {
			t.Logf("Job %s has been executed...", job.Title)
		}
	}
}

func TestSchedulePipeline(t *testing.T) {
	belaur.Cfg = &belaur.Config{}
	storeInstance := store.NewBoltStore()
	tmp, _ := ioutil.TempDir("", "TestSchedulePipeline")
	belaur.Cfg.DataPath = tmp
	belaur.Cfg.WorkspacePath = filepath.Join(tmp, "tmp")
	belaur.Cfg.Bolt.Mode = 0600
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: hclog.DefaultOutput,
		Name:   "Belaur",
	})
	belaur.Cfg.Worker = 2
	if err := storeInstance.Init(tmp); err != nil {
		t.Fatal(err)
	}
	p, _ := prepareTestData()
	_ = storeInstance.PipelinePut(&p)
	s, err := NewScheduler(Dependencies{storeInstance, &MemDBFake{}, &PluginFakeFailed{}, &CAFake{}, &VaultFake{}})
	if err != nil {
		t.Fatal(err)
	}
	s.Init()
	_, err = s.SchedulePipeline(&p, belaur.StartReasonManual, prepareArgs())
	if err != nil {
		t.Fatal(err)
	}
}

func TestSchedulePipelineParallel(t *testing.T) {
	belaur.Cfg = &belaur.Config{}
	storeInstance := store.NewBoltStore()
	tmp, _ := ioutil.TempDir("", "TestSchedulePipeline")
	belaur.Cfg.DataPath = tmp
	belaur.Cfg.WorkspacePath = filepath.Join(tmp, "tmp")
	belaur.Cfg.Bolt.Mode = 0600
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: hclog.DefaultOutput,
		Name:   "Belaur",
	})
	belaur.Cfg.Worker = 2
	if err := storeInstance.Init(tmp); err != nil {
		t.Fatal(err)
	}
	p1 := belaur.Pipeline{
		ID:   0,
		Name: "Test Pipeline 1",
		Type: belaur.PTypeGolang,
		Jobs: prepareJobs(),
	}
	p2 := belaur.Pipeline{
		ID:   1,
		Name: "Test Pipeline 2",
		Type: belaur.PTypeGolang,
		Jobs: prepareJobs(),
	}
	_ = storeInstance.PipelinePut(&p1)
	_ = storeInstance.PipelinePut(&p2)
	s, err := NewScheduler(Dependencies{storeInstance, &MemDBFake{}, &PluginFakeFailed{}, &CAFake{}, &VaultFake{}})
	if err != nil {
		t.Fatal(err)
	}
	s.Init()
	var run1 *belaur.PipelineRun
	var run2 *belaur.PipelineRun
	var wg sync.WaitGroup
	wg.Add(2)
	go func() {
		run1, _ = s.SchedulePipeline(&p1, belaur.StartReasonManual, prepareArgs())
		wg.Done()
	}()
	go func() {
		run2, _ = s.SchedulePipeline(&p2, belaur.StartReasonManual, prepareArgs())
		wg.Done()
	}()
	wg.Wait()
	if run1.ID == run2.ID {
		t.Fatal("the two run jobs id should not have equalled. was: ", run1.ID, run2.ID)
	}
}

func TestSchedule(t *testing.T) {
	belaur.Cfg = &belaur.Config{}
	storeInstance := store.NewBoltStore()
	tmp, _ := ioutil.TempDir("", "TestSchedule")
	belaur.Cfg.DataPath = tmp
	belaur.Cfg.WorkspacePath = filepath.Join(tmp, "tmp")
	belaur.Cfg.Bolt.Mode = 0600
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: hclog.DefaultOutput,
		Name:   "Belaur",
	})
	belaur.Cfg.Worker = 2
	if err := storeInstance.Init(tmp); err != nil {
		t.Fatal(err)
	}
	p, _ := prepareTestData()
	_ = storeInstance.PipelinePut(&p)
	s, err := NewScheduler(Dependencies{storeInstance, &MemDBFake{}, &PluginFakeFailed{}, &CAFake{}, &VaultFake{}})
	if err != nil {
		t.Fatal(err)
	}
	_, err = s.SchedulePipeline(&p, belaur.StartReasonManual, prepareArgs())
	if err != nil {
		t.Fatal(err)
	}
	s.schedule()
	r, err := storeInstance.PipelineGetRunByPipelineIDAndID(p.ID, 1)
	if err != nil {
		t.Fatal(err)
	}
	if r.Status != belaur.RunScheduled {
		t.Fatalf("run has status %s but should be %s\n", r.Status, string(belaur.RunScheduled))
	}
}

func TestSetPipelineJobs(t *testing.T) {
	belaur.Cfg = &belaur.Config{}
	storeInstance := store.NewBoltStore()
	tmp, _ := ioutil.TempDir("", "TestSetPipelineJobs")
	belaur.Cfg.DataPath = tmp
	belaur.Cfg.WorkspacePath = filepath.Join(tmp, "tmp")
	belaur.Cfg.Bolt.Mode = 0600
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: hclog.DefaultOutput,
		Name:   "Belaur",
	})
	if err := storeInstance.Init(tmp); err != nil {
		t.Fatal(err)
	}
	p, _ := prepareTestData()
	p.Jobs = nil
	s, err := NewScheduler(Dependencies{storeInstance, &MemDBFake{}, &PluginFakeFailed{}, &CAFake{}, &VaultFake{}})
	if err != nil {
		t.Fatal(err)
	}
	err = s.SetPipelineJobs(&p)
	if err != nil {
		t.Fatal(err)
	}
	if len(p.Jobs) != 4 {
		t.Fatalf("Number of jobs should be 4 but was %d\n", len(p.Jobs))
	}
}

func TestStopPipelineRun(t *testing.T) {
	belaur.Cfg = &belaur.Config{}
	storeInstance := store.NewBoltStore()
	tmp, _ := ioutil.TempDir("", "TestStopPipelineRun")
	belaur.Cfg.DataPath = tmp
	belaur.Cfg.WorkspacePath = filepath.Join(tmp, "tmp")
	belaur.Cfg.Bolt.Mode = 0600
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: hclog.DefaultOutput,
		Name:   "Belaur",
	})
	belaur.Cfg.Worker = 2
	if err := storeInstance.Init(tmp); err != nil {
		t.Fatal(err)
	}
	p, r := prepareTestData()
	_ = storeInstance.PipelinePut(&p)
	s, err := NewScheduler(Dependencies{storeInstance, &MemDBFake{}, &PluginFake{}, &CAFake{}, &VaultFake{}})
	if err != nil {
		t.Fatal(err)
	}

	r.Status = belaur.RunRunning
	_ = storeInstance.PipelinePutRun(&r)

	run, _ := storeInstance.PipelineGetRunByPipelineIDAndID(p.ID, r.ID)
	err = s.StopPipelineRun(&p, run.ID)
	if err != nil {
		t.Fatal(err)
	}
	run, _ = storeInstance.PipelineGetRunByPipelineIDAndID(p.ID, r.ID)
	if run.Status != belaur.RunCancelled {
		t.Fatal("expected pipeline state to be cancelled. got: ", r.Status)
	}
}

func prepareArgs() []*belaur.Argument {
	arg1 := belaur.Argument{
		Description: "First Arg",
		Key:         "firstarg",
		Type:        "textfield",
	}
	arg2 := belaur.Argument{
		Description: "Second Arg",
		Key:         "secondarg",
		Type:        "textarea",
	}
	arg3 := belaur.Argument{
		Description: "Vault Arg",
		Key:         "vaultarg",
		Type:        "vault",
	}
	return []*belaur.Argument{&arg1, &arg2, &arg3}
}

func prepareJobs() []*belaur.Job {
	job1 := belaur.Job{
		ID:        hash("Job1"),
		Title:     "Job1",
		DependsOn: []*belaur.Job{},
		Status:    belaur.JobWaitingExec,
		Args:      prepareArgs(),
	}
	job2 := belaur.Job{
		ID:        hash("Job2"),
		Title:     "Job2",
		DependsOn: []*belaur.Job{&job1},
		Status:    belaur.JobWaitingExec,
	}
	job3 := belaur.Job{
		ID:        hash("Job3"),
		Title:     "Job3",
		DependsOn: []*belaur.Job{&job2},
		Status:    belaur.JobWaitingExec,
	}
	job4 := belaur.Job{
		ID:        hash("Job4"),
		Title:     "Job4",
		DependsOn: []*belaur.Job{&job3},
		Status:    belaur.JobWaitingExec,
	}

	return []*belaur.Job{
		&job1,
		&job2,
		&job3,
		&job4,
	}
}

func prepareTestData() (pipeline belaur.Pipeline, pipelineRun belaur.PipelineRun) {
	pipeline = belaur.Pipeline{
		ID:   1,
		Name: "Test Pipeline",
		Type: belaur.PTypeGolang,
		Jobs: prepareJobs(),
	}
	v4, _ := uuid.NewV4()
	pipelineRun = belaur.PipelineRun{
		ID:         1,
		PipelineID: 1,
		Status:     belaur.RunNotScheduled,
		UniqueID:   uuid.Must(v4, nil).String(),
		Jobs:       pipeline.Jobs,
	}
	return
}

// hash hashes the given string.
func hash(s string) uint32 {
	h := fnv.New32a()
	_, _ = h.Write([]byte(s))
	return h.Sum32()
}
