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
	"errors"
	"fmt"
	"os"
	"path/filepath"
	"strconv"
	"sync"
	"sync/atomic"
	"time"

	belaur "github.com/bhojpur/belaur"
	"github.com/bhojpur/belaur/pkg/helper/stringhelper"
	"github.com/bhojpur/belaur/pkg/plugin"
	"github.com/bhojpur/belaur/pkg/security"
	"github.com/bhojpur/belaur/pkg/store"
	"github.com/bhojpur/belaur/pkg/store/memdb"
	"github.com/bhojpur/belaur/pkg/workers/docker"
	"github.com/gofrs/uuid"
)

const (
	// Maximum buffer limit for scheduler
	schedulerBufferLimit = 50

	// schedulerIntervalSeconds defines the interval the scheduler will look
	// for new work to schedule. Definition in seconds.
	schedulerIntervalSeconds = 3

	// errCircularDep is thrown when a circular dependency has been detected.
	errCircularDep = "circular dependency detected between %s and %s"

	// argTypeVault is the argument type vault.
	argTypeVault = "vault"

	// logFlushInterval defines the interval where logs will be flushed to disk.
	logFlushInterval = 1
)

var (
	// errCreateCMDForPipeline is thrown when we couldn't create a command to start
	// a plugin.
	errCreateCMDForPipeline = errors.New("could not create execute command for plugin")

	// Java executable name
	javaExecName = "java"

	// Python executable name
	pythonExecName = "python"

	// Ruby executable name
	rubyExecName = "ruby"

	// Ruby gem binary name
	rubyGemName = "gem"

	// NodeJS binary name
	nodeJSExecName = "node"
)

// BelaurScheduler is a job scheduler for Bhojpur Belaur pipeline runs.
type BelaurScheduler interface {
	Init()
	SchedulePipeline(p *belaur.Pipeline, startedBy string, args []*belaur.Argument) (*belaur.PipelineRun, error)
	SetPipelineJobs(p *belaur.Pipeline) error
	StopPipelineRun(p *belaur.Pipeline, runID int) error
	GetFreeWorkers() int32
	CountScheduledRuns() int
}

var _ BelaurScheduler = (*Scheduler)(nil)

// Scheduler represents the schuler object
type Scheduler struct {
	// buffered channel which is used as queue
	scheduledRuns chan belaur.PipelineRun

	// storeService is an instance of store.
	storeService store.BelaurStore

	// memDBService is an instance of the memDB.
	memDBService memdb.BelaurMemDB

	// pluginSystem is the used plugin system.
	pluginSystem plugin.Plugin

	// ca is the instance of the CA used to handle certs.
	ca security.CAAPI

	// vault is the instance of the vault.
	vault security.BelaurVault

	// Atomic Counter that represents the current free workers
	freeWorkers *int32

	// Lock for scheduling
	schedulePipelineLock sync.RWMutex
	// Lock for scheduling
	schedulerLock sync.RWMutex

	// killedPipelineRun is used to signal the scheduler to abort a pipeline run.
	// This has the size one for delayed guarantee of signal delivery.
	killedPipelineRun chan *belaur.PipelineRun
}

// Dependencies defines the dependencies of the scheduler service.
type Dependencies struct {
	Store store.BelaurStore
	DB    memdb.BelaurMemDB
	PS    plugin.Plugin
	CA    security.CAAPI
	Vault security.BelaurVault
}

// NewScheduler creates a new Scheduler service.
func NewScheduler(deps Dependencies) (*Scheduler, error) {
	// Create new scheduler
	s := &Scheduler{
		scheduledRuns:     make(chan belaur.PipelineRun, schedulerBufferLimit),
		storeService:      deps.Store,
		memDBService:      deps.DB,
		pluginSystem:      deps.PS,
		ca:                deps.CA,
		vault:             deps.Vault,
		freeWorkers:       new(int32),
		killedPipelineRun: make(chan *belaur.PipelineRun, 1),
	}
	return s, nil
}

// Init initializes the scheduler.
func (s *Scheduler) Init() {
	// Setup worker
	for i := 0; i < belaur.Cfg.Worker; i++ {
		go s.work()
	}

	// Create a periodic job that fills the scheduler with new pipelines.
	schedulerJob := time.NewTicker(schedulerIntervalSeconds * time.Second)
	go func() {
		for {
			select {
			case <-schedulerJob.C:
				// Do the scheduling
				s.schedule()
			}
		}
	}()
}

// work takes work from the scheduled run buffer channel and starts
// the preparation and execution of the pipeline. Then repeats.
func (s *Scheduler) work() {
	// This worker never stops working.
	for {
		// We haven't picked up work yet so mark this worker as free
		atomic.AddInt32(s.freeWorkers, 1)

		// Take one scheduled run, block if there are no scheduled pipelines
		r := <-s.scheduledRuns

		// We picked up work and are from now on busy
		atomic.AddInt32(s.freeWorkers, -1)

		// Prepare execution and start it
		s.prepareAndExec(r)
	}
}

// prepareAndExec does the preparation and starts the execution.
func (s *Scheduler) prepareAndExec(r belaur.PipelineRun) {
	// Mark the scheduled run as running
	r.Status = belaur.RunRunning
	r.StartDate = time.Now()

	// Update entry in store
	err := s.storeService.PipelinePutRun(&r)
	if err != nil {
		belaur.Cfg.Logger.Debug("could not put pipeline run into store during executing work", "error", err.Error())
		return
	}

	// Get related pipeline from pipeline run
	pipeline, _ := s.storeService.PipelineGet(r.PipelineID)

	// Check if this pipeline has jobs declared
	if len(r.Jobs) == 0 {
		// Finish pipeline run
		s.finishPipelineRun(&r, belaur.RunSuccess)
		return
	}

	// Check if circular dependency exists
	for _, job := range r.Jobs {
		if _, err := s.checkCircularDep(job, []*belaur.Job{}, []*belaur.Job{}); err != nil {
			belaur.Cfg.Logger.Info("circular dependency detected", "pipeline", pipeline)
			belaur.Cfg.Logger.Info("information", "info", err.Error())

			// Update store
			s.finishPipelineRun(&r, belaur.RunFailed)
			return
		}
	}

	// Create logs folder for this run
	path := filepath.Join(belaur.Cfg.WorkspacePath, strconv.Itoa(r.PipelineID), strconv.Itoa(r.ID), belaur.LogsFolderName)
	err = os.MkdirAll(path, 0700)
	if err != nil {
		belaur.Cfg.Logger.Error("cannot create pipeline run folder", "error", err.Error(), "path", path)
	}

	// Create the start command for the pipeline
	c := createPipelineCmd(pipeline)
	if c == nil {
		belaur.Cfg.Logger.Debug("cannot create pipeline start command", "error", errCreateCMDForPipeline.Error())
		s.finishPipelineRun(&r, belaur.RunFailed)
		return
	}

	// Create new plugin instance
	pS := s.pluginSystem.NewPlugin(s.ca)

	// Init the plugin
	path = filepath.Join(path, belaur.LogsFileName)
	if err := pS.Init(c, &path); err != nil {
		belaur.Cfg.Logger.Debug("cannot initialize the plugin", "error", err.Error(), "pipeline", pipeline)
		s.finishPipelineRun(&r, belaur.RunFailed)
		return
	}

	// Validate the plugin(pipeline)
	if err := pS.Validate(); err != nil {
		belaur.Cfg.Logger.Debug("cannot validate pipeline", "error", err.Error(), "pipeline", pipeline)
		s.finishPipelineRun(&r, belaur.RunFailed)
		return
	}
	defer pS.Close()

	// Schedule jobs and execute them.
	// Also update the run in the store.
	s.executeScheduledJobs(r, pS)
}

// schedule looks in the store for new work and schedules it.
func (s *Scheduler) schedule() {
	s.schedulerLock.Lock()
	defer s.schedulerLock.Unlock()

	// Do we have space left in our buffer?
	if s.CountScheduledRuns() >= schedulerBufferLimit {
		// No space left. Exit.
		return
	}

	// Get scheduled pipelines but limit the returning number of elements.
	scheduled, err := s.storeService.PipelineGetScheduled(schedulerBufferLimit)
	if err != nil {
		belaur.Cfg.Logger.Debug("cannot get scheduled pipelines", "error", err.Error())
		return
	}

	// Iterate scheduled runs
	for id := range scheduled {
		// Small helper function to update the pipeline run status in the store
		storeUpdate := func(run *belaur.PipelineRun, status belaur.PipelineRunStatus) {
			// Update entry in store
			run.Status = status
			if err := s.storeService.PipelinePutRun(run); err != nil {
				belaur.Cfg.Logger.Debug("could not put pipeline run into store via schedule", "error", err.Error(), "run", run)
			}
		}

		// If we are a server instance, we will by default give the worker the advantage.
		// Only in case all workers are busy we will schedule work on the server.
		workers := s.memDBService.GetAllWorker()
		if belaur.Cfg.Mode == belaur.ModeServer && len(workers) > 0 {
			// Check if we have a suitable worker at all
			invalidWorkers := 0
			for _, w := range workers {
				switch {
				case w.Slots == 0:
					invalidWorkers++
				case w.Status != belaur.WorkerActive:
					invalidWorkers++
				case stringhelper.IsContainedInSlice(w.Tags, "dockerworker", true):
					invalidWorkers++
				}
			}

			// Insert pipeline run into memdb where all workers get their work from
			if len(workers) > invalidWorkers {
				if err := s.memDBService.InsertPipelineRun(scheduled[id]); err != nil {
					belaur.Cfg.Logger.Error("failed to insert pipeline run into memdb via schedule", "error", err.Error())
					continue
				}
				storeUpdate(scheduled[id], belaur.RunScheduled)
				continue
			}
		}

		// Check if this primary is not allowed to run work
		if belaur.Cfg.PreventPrimaryWork {
			continue
		}

		// Check if this pipeline run is a docker run
		if scheduled[id].Docker || belaur.Cfg.AutoDockerMode {
			// Retrieve the global worker registration secret
			if err = s.vault.LoadSecrets(); err != nil {
				belaur.Cfg.Logger.Error("failed to load secrets from vault", "error", err)
				continue
			}
			workerSecretBytes, err := s.vault.Get(belaur.WorkerRegisterKey)
			if err != nil {
				belaur.Cfg.Logger.Error("failed to get global worker registration secret from vault", "error", err)
				continue
			}
			workerSecret := string(workerSecretBytes[:])

			// Start docker worker for this pipeline run
			worker := docker.NewDockerWorker(belaur.Cfg.DockerHostURL, scheduled[id].UniqueID)
			if err := worker.SetupDockerWorker(belaur.Cfg.DockerRunImage, workerSecret); err != nil {
				belaur.Cfg.Logger.Error("failed to setup docker worker for pipeline run", "error", err)
				continue
			}

			// Cache docker worker
			if err := s.memDBService.InsertDockerWorker(worker); err != nil {
				belaur.Cfg.Logger.Error("failed to cache docker worker in memdb", "error", err)
				if err := worker.KillDockerWorker(); err != nil {
					belaur.Cfg.Logger.Error("failed to kill docker worker", "error", err)
				}
				continue
			}

			// Prevent the docker worker to start another docker worker container
			scheduled[id].Docker = false

			// If it is a docker run, pipeline will be executed by a worker inside a container
			scheduled[id].DockerWorkerID = worker.WorkerID
			scheduled[id].PipelineTags = append(scheduled[id].PipelineTags, []string{worker.WorkerID, "dockerworker"}...)
			if err := s.memDBService.InsertPipelineRun(scheduled[id]); err != nil {
				belaur.Cfg.Logger.Error("failed to insert pipeline run into memdb via schedule", "error", err.Error())
				continue
			}

			// Reset the docker status manipulation
			scheduled[id].Docker = true
			storeUpdate(scheduled[id], belaur.RunScheduled)
			continue
		}

		// push scheduled run into our channel
		s.scheduledRuns <- *scheduled[id]

		// Run is now scheduled
		storeUpdate(scheduled[id], belaur.RunScheduled)
	}
}

// StopPipelineRun will prematurely cancel a pipeline run by killing all of its
// jobs and running processes immediately.
func (s *Scheduler) StopPipelineRun(p *belaur.Pipeline, runID int) error {
	pr, err := s.storeService.PipelineGetRunByPipelineIDAndID(p.ID, runID)
	if err != nil {
		return err
	}
	if pr.Status == belaur.RunFailed || pr.Status == belaur.RunCancelled || pr.Status == belaur.RunSuccess {
		return errors.New("pipeline is not in a cancellable state")
	}

	pr.Status = belaur.RunCancelled
	err = s.storeService.PipelinePutRun(pr)
	if err != nil {
		return err
	}
	s.killedPipelineRun <- pr

	return nil
}

// SchedulePipeline schedules a pipeline. We create a new schedule object
// and save it in our store. The scheduler will later pick this up and will continue the work.
func (s *Scheduler) SchedulePipeline(p *belaur.Pipeline, startedReason string, args []*belaur.Argument) (*belaur.PipelineRun, error) {

	// Introduce a semaphore locking here because this function can be called
	// in parallel if multiple users happen to trigger a pipeline run at the same time.
	// (or someone is just simply eager and presses (Start Pipeline) in quick successions).
	// This means that one of the calls will take slightly longer (a couple of nanoseconds)
	// while the other finishes to save the pipelinerun.
	// This is to ensure that the highest ID for the next pipeline is calculated properly.
	s.schedulePipelineLock.Lock()
	defer s.schedulePipelineLock.Unlock()

	// Get highest public id used for this pipeline
	highestID, err := s.storeService.PipelineGetRunHighestID(p)
	if err != nil {
		belaur.Cfg.Logger.Error("cannot find highest pipeline run id", "error", err.Error())
		return nil, err
	}

	// increment by one
	highestID++

	// Get jobs
	jobs, err := s.getPipelineJobs(p)
	if err != nil {
		belaur.Cfg.Logger.Error("cannot get pipeline jobs during schedule", "error", err.Error(), "pipeline", p)
		return nil, err
	}

	// Load secret from vault and set it
	err = s.vault.LoadSecrets()
	if err != nil {
		belaur.Cfg.Logger.Error("cannot load secrets from vault during schedule pipeline", "error", err.Error())
		return nil, err
	}
	// We have to go through all jobs to find the related arguments.
	// We will only pass related arguments to the specific job.
	for jobID, job := range jobs {
		if job.Args != nil {
			for argID, arg := range job.Args {
				// check if it's of type vault
				if arg.Type == argTypeVault {
					// Get & Set argument
					s, err := s.vault.Get(arg.Key)
					if err != nil {
						belaur.Cfg.Logger.Error("cannot find secret with given key in vault", "key", arg.Key, "pipeline", p)
						return nil, err
					}
					jobs[jobID].Args[argID].Value = string(s)
				} else {
					// Find related argument in given arguments
					for _, givenArg := range args {
						if arg.Key == givenArg.Key {
							jobs[jobID].Args[argID] = givenArg
						}
					}
				}
			}
		}
	}

	// Create new not scheduled pipeline run
	v4, err := uuid.NewV4()
	if err != nil {
		return nil, err
	}
	run := belaur.PipelineRun{
		UniqueID:     uuid.Must(v4, nil).String(),
		ID:           highestID,
		PipelineID:   p.ID,
		ScheduleDate: time.Now(),
		Jobs:         jobs,
		Status:       belaur.RunNotScheduled,
		PipelineType: p.Type,
		PipelineTags: p.Tags,
		Docker:       p.Docker,
		StartReason:  startedReason,
	}

	// Put run into store
	return &run, s.storeService.PipelinePutRun(&run)
}

// executeJob executes a job and informs via triggerSave that the job can be saved to the store.
// This method is blocking.
func executeJob(j belaur.Job, pS plugin.Plugin, triggerSave chan belaur.Job) {
	// Set Job to running and trigger save
	j.Status = belaur.JobRunning
	triggerSave <- j

	// Execute job
	if err := pS.Execute(&j); err != nil {
		belaur.Cfg.Logger.Debug("error during job execution", "error", err.Error(), "job", j)
	}

	// Trigger another save to store the result of the execute
	triggerSave <- j
}

// checkCircularDep checks for circular dependencies.
// An error is thrown when one is found.
func (s *Scheduler) checkCircularDep(j *belaur.Job, resolved []*belaur.Job, unresolved []*belaur.Job) ([]*belaur.Job, error) {
	unresolved = append(unresolved, j)

DependsonLoop:
	for _, job := range j.DependsOn {
		// Check if job is already in resolved list
		for _, resolvedJob := range resolved {
			if resolvedJob.ID == job.ID {
				continue DependsonLoop
			}
		}

		// Check if job is already in unresolved list
		for _, unresolvedJob := range unresolved {
			if unresolvedJob.ID == job.ID {
				// Circular dependency detected
				// Return the conflicting dependencies
				return nil, fmt.Errorf(errCircularDep, unresolvedJob.Title, j.Title)
			}
		}

		// Resolve job
		var err error
		resolved, err = s.checkCircularDep(job, resolved, unresolved)
		if err != nil {
			return nil, err
		}
	}

	return append(resolved, j), nil
}

// resolveDependencies resolves the dependencies of the given workload
// and sends all resolved workloads to our executeScheduler queue.
// After a workload has been send to the executeScheduler, the method will
// block and wait until the workload is done.
// This method is designed to be called recursive and blocking.
func (s *Scheduler) resolveDependencies(j *belaur.Job, mw *managedWorkloads, executeScheduler chan *belaur.Job, done chan bool) {
	for _, depJob := range j.DependsOn {
		// Check if this workload is already resolved
		var resolved bool
		for workload := range mw.Iter() {
			if workload.done && workload.job.ID == depJob.ID {
				resolved = true
			}
		}

		// Job has been resolved
		if resolved {
			continue
		}

		// Resolve job
		s.resolveDependencies(depJob, mw, executeScheduler, done)
	}

	// Queue used to signal that the work should be finished soon.
	// We do not block here because this is just a pre-validation step.
	select {
	case _, ok := <-done:
		if !ok {
			return
		}
	default:
	}

	// If we are here, then the job is resolved.
	// We have to check if the job still needs to be run
	// or if another goroutine has already started the execution.
	relatedWL := mw.GetByID(j.ID)
	if !relatedWL.started {
		// Job has not been executed yet.
		// Send workload to execute scheduler.
		executeScheduler <- j

		// Wait until finished
		<-relatedWL.finishedSig
	} else if !relatedWL.done {
		// Job has been started but not finished.
		// Let us wait till finished.
		<-relatedWL.finishedSig
	}
}

// executeScheduledJobs is a small wrapper around executeScheduler which
// is responsible for finalizing the pipeline run.
func (s *Scheduler) executeScheduledJobs(r belaur.PipelineRun, pS plugin.Plugin) {
	// Start the main execute process and wait until finished.
	s.executeScheduler(&r, pS)

	// Run finished. Set pipeline status.
	var runFail bool
	for _, job := range r.Jobs {
		if job.Status != belaur.JobSuccess && job.FailPipeline {
			runFail = true
		}
	}

	if runFail && r.Status != belaur.RunCancelled {
		s.finishPipelineRun(&r, belaur.RunFailed)
	} else if r.Status == belaur.RunCancelled {
		s.finishPipelineRun(&r, belaur.RunCancelled)
	} else {
		s.finishPipelineRun(&r, belaur.RunSuccess)
	}
}

// executeScheduler is our main function which coordinates the
// whole execution process and dependency resolve algorithm.
func (s *Scheduler) executeScheduler(r *belaur.PipelineRun, pS plugin.Plugin) {
	// Create a queue which is used to execute the resolved workloads.
	executeScheduler := make(chan *belaur.Job)

	// Done queue to signal go routines to exit.
	// This is usually used when a job failed and the whole pipeline
	// should be cancelled.
	done := make(chan bool)

	// Iterate all jobs from this run
	mw := newManagedWorkloads()
	for id := range r.Jobs {
		// Create new workload object
		mw.Append(workload{
			job:         r.Jobs[id],
			finishedSig: make(chan bool),
		})

		// Start resolving go routine for this job
		go s.resolveDependencies(r.Jobs[id], mw, executeScheduler, done)
	}

	// Create a new ticker (scheduled go routine) which periodically
	// flushes the logs buffer.
	ticker := time.NewTicker(logFlushInterval * time.Second)
	pipelineFinished := make(chan bool)
	go func() {
		defer ticker.Stop()
		for {
			select {
			case <-ticker.C:
				_ = pS.FlushLogs()
			case _, ok := <-pipelineFinished:
				if !ok {
					return
				}
			}
		}
	}()

	// Separate channel to save updates about the status of job executions.
	triggerSave := make(chan belaur.Job)

	// Let's loop until we are done
	var finalize bool
	finished := make(chan bool, 1)
	for {
		select {
		case pr, ok := <-s.killedPipelineRun:
			if ok {
				if pr.ID == r.ID {
					for _, job := range r.Jobs {
						if job.Status == belaur.JobRunning || job.Status == belaur.JobWaitingExec {
							job.Status = belaur.JobFailed
							job.FailPipeline = true
						}
					}
					r.Status = belaur.RunCancelled
					_ = s.storeService.PipelinePutRun(r)
					close(done)
					close(executeScheduler)
					finished <- true
					finalize = true
					return
				}
			}
		case <-finished:
			close(pipelineFinished)
			return
		case j, ok := <-triggerSave:
			if !ok {
				break
			}

			// Filter out the job and update the real job
			for id, job := range r.Jobs {
				if job.ID == j.ID {
					r.Jobs[id].Status = j.Status
					r.Jobs[id].FailPipeline = j.FailPipeline
					break
				}
			}

			// Store status update
			_ = s.storeService.PipelinePutRun(r)

			// Send signal to resolver that this job is finished.
			if j.Status == belaur.JobSuccess || j.Status == belaur.JobFailed {
				// Job is done
				wl := mw.GetByID(j.ID)
				wl.done = true
				mw.Replace(*wl)

				// Let's check if we are done and if all jobs ran successful.
				var allWLDone = true
				for wl := range mw.Iter() {
					if !wl.done {
						allWLDone = false
					}
				}

				if allWLDone && !finalize {
					close(done)
					close(executeScheduler)
					close(triggerSave)
					finished <- true
					finalize = true
				}

				// Close go-routine which was waiting for this job.
				close(wl.finishedSig)
			}

			// Dependent of the status output, decide what should happen next.
			if !finalize && j.Status == belaur.JobFailed {
				// we are entering the finalize phase
				finalize = true

				// Send done signal to all resolvers
				close(done)

				// read all jobs which are waiting to be executed to free the channel
				var channelClean = false
				for !channelClean {
					select {
					case <-executeScheduler:
						// just read from the channel
					default:
						channelClean = true
					}
				}

				// Close executeScheduler. No new jobs should be scheduled.
				close(executeScheduler)

				// A job failed. Finish all currently running jobs.
				go func() {
					// We might have still running jobs. Wait until all jobs are finished.
					for {
						var notFinishedWorkloadChan chan bool
						for singleWL := range mw.Iter() {
							if singleWL.started && !singleWL.done {
								notFinishedWorkloadChan = singleWL.finishedSig
							}
						}

						if notFinishedWorkloadChan == nil {
							break
						}

						// wait until finished
						<-notFinishedWorkloadChan
					}

					finished <- true
					close(triggerSave)
				}()
			}
		case j, ok := <-executeScheduler:
			if !ok {
				break
			}

			// Get related workload
			wl := mw.GetByID(j.ID)

			// Check if this workload has been already started by another routine.
			if !wl.started {
				// Update
				wl.started = true
				mw.Replace(*wl)

				// Start execution
				go executeJob(*j, pS, triggerSave)
			}
		}
	}
}

// getPipelineJobs uses the plugin system to get all jobs from the given pipeline.
func (s *Scheduler) getPipelineJobs(p *belaur.Pipeline) ([]*belaur.Job, error) {
	// Create the start command for the pipeline
	c := createPipelineCmd(p)
	if c == nil {
		belaur.Cfg.Logger.Debug("cannot set pipeline jobs", "error", errCreateCMDForPipeline.Error(), "pipeline", p)
		return nil, errCreateCMDForPipeline
	}

	// Create new Plugin instance
	pS := s.pluginSystem.NewPlugin(s.ca)

	// Init the go-plugin
	if err := pS.Init(c, nil); err != nil {
		belaur.Cfg.Logger.Debug("cannot initialize the pipeline", "error", err.Error(), "pipeline", p)
		return nil, err
	}

	// Validate the plugin(pipeline)
	if err := pS.Validate(); err != nil {
		belaur.Cfg.Logger.Debug("cannot validate pipeline", "error", err.Error(), "pipeline", p)
		return nil, err
	}
	defer pS.Close()

	return pS.GetJobs()
}

// SetPipelineJobs uses the plugin system to get all jobs from the given pipeline.
// This function is blocking and might take some time.
func (s *Scheduler) SetPipelineJobs(p *belaur.Pipeline) error {
	// Get jobs
	jobs, err := s.getPipelineJobs(p)
	if err != nil {
		return err
	}
	p.Jobs = jobs

	return nil
}

// GetFreeWorkers returns the number of free workers.
func (s *Scheduler) GetFreeWorkers() int32 {
	return atomic.LoadInt32(s.freeWorkers)
}

// CountScheduledRuns returns the number of scheduled runs.
func (s *Scheduler) CountScheduledRuns() int {
	return len(s.scheduledRuns)
}

// finishPipelineRun finishes the pipeline run and stores the results.
func (s *Scheduler) finishPipelineRun(r *belaur.PipelineRun, status belaur.PipelineRunStatus) {
	// Set pipeline run status
	r.Status = status

	// Finish date
	r.FinishDate = time.Now()

	// Store it
	err := s.storeService.PipelinePutRun(r)
	if err != nil {
		belaur.Cfg.Logger.Error("cannot store finished pipeline", "error", err.Error())
	}
}
