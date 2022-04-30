package agent

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
	"crypto/tls"
	"crypto/x509"
	"encoding/base64"
	"errors"
	"fmt"
	"io"
	"io/ioutil"
	"os"
	"path/filepath"
	"strconv"
	"strings"
	"syscall"
	"time"

	belaur "github.com/bhojpur/belaur"
	pb "github.com/bhojpur/belaur/pkg/api/v1/worker"
	"github.com/bhojpur/belaur/pkg/helper/filehelper"
	"github.com/bhojpur/belaur/pkg/helper/pipelinehelper"
	"github.com/bhojpur/belaur/pkg/security"
	"github.com/bhojpur/belaur/pkg/store"
	"github.com/bhojpur/belaur/pkg/workers/agent/api"
	gp "github.com/bhojpur/belaur/pkg/workers/pipeline"
	"github.com/bhojpur/belaur/pkg/workers/scheduler/service"
	"google.golang.org/grpc"
	_ "google.golang.org/grpc/balancer/grpclb" // needed
	"google.golang.org/grpc/credentials"
	"google.golang.org/grpc/metadata"
)

const (
	// schedulerTickerSeconds defines the interval in seconds for the scheduler.
	schedulerTickerSeconds = 3

	// updateTickerSeconds defines the interval in seconds to send updates.
	updateTickerSeconds = 2

	// chunkSize is the size of binary chunks transferred to workers.
	chunkSize = 64 * 1024 // 64 KiB

	// idMDKey is the key used for the gRPC metadata map.
	idMDKey = "uniqueid"

	// defaultHostname is the default hostname set for the mTLS certificate.
	defaultHostname = "bhojpur.net"
)

// Agent represents an instance of an agent
type Agent struct {
	// client represents the interface for the worker client
	client pb.WorkerClient

	// self represents the current agent instance information
	self *pb.WorkerInstance

	// certFile represents the local path to the agent cert
	certFile string

	// keyFile represents the local path to the agent key
	keyFile string

	// caCertFile represents the local path to the agent ca cert
	caCertFile string

	// Instance of scheduler
	scheduler service.BelaurScheduler

	// Pipeline Servicer
	pipelineService gp.Servicer

	// Instance of store
	store store.BelaurStore

	// Signal channel for this agent
	exitChan chan os.Signal
}

// InitAgent initiates the agent instance
func InitAgent(exitChan chan os.Signal, scheduler service.BelaurScheduler, pipelineService gp.Servicer, store store.BelaurStore, certPath string) *Agent {
	ag := &Agent{
		exitChan:        exitChan,
		scheduler:       scheduler,
		store:           store,
		pipelineService: pipelineService,
	}

	// Set path to local certificates
	ag.certFile = filepath.Join(certPath, "cert.pem")
	ag.keyFile = filepath.Join(certPath, "key.pem")
	ag.caCertFile = filepath.Join(certPath, "caCert.pem")

	// return instance
	return ag
}

// StartAgent starts the agent and returns a clean up function.
func (a *Agent) StartAgent() (func(), error) {
	// Setup connection information
	clientTLS, err := a.setupConnectionInfo()
	if err != nil {
		return nil, err
	}

	// Setup gRPC connection
	dialOption := grpc.WithTransportCredentials(clientTLS)
	conn, err := grpc.Dial(belaur.Cfg.WorkerGRPCHostURL, dialOption)
	if err != nil {
		return nil, fmt.Errorf("failed to connect to remote host: %s", err.Error())
	}

	// Get worker interface
	a.client = pb.NewWorkerClient(conn)

	// Start periodic go routine which schedules the worker work
	workTicker := time.NewTicker(schedulerTickerSeconds * time.Second)
	quitScheduler := make(chan struct{})
	go func() {
		for {
			select {
			case <-workTicker.C:
				// execute schedule function
				a.scheduleWork()
			case <-quitScheduler:
				workTicker.Stop()
				return
			}
		}
	}()

	// Start periodic go routine which sends back information to the Bhojpur Belaur primary instance
	updateTicker := time.NewTicker(updateTickerSeconds * time.Second)
	quitUpdate := make(chan struct{})
	go func() {
		for {
			select {
			case <-updateTicker.C:
				// run update function
				a.updateWork()
			case <-quitUpdate:
				updateTicker.Stop()
				return
			}
		}
	}()

	// Return clean up function
	return func() {
		// Close gRPC connection
		if err := conn.Close(); err != nil {
			belaur.Cfg.Logger.Error("failed to close gRPC connection", "error", err)
		}

		// Safely stop scheduler
		close(quitScheduler)
		close(quitUpdate)
	}, nil
}

// setupConnectionInfo setups the connection info object by parsing existing
// mTLS certificates or registering the worker at the Bhojpur Belaur primary instance
// and receiving new mTLS certs.
func (a *Agent) setupConnectionInfo() (credentials.TransportCredentials, error) {
	// Evaluate and summarize all worker tags
	var tags []string
	if belaur.Cfg.WorkerTags != "" {
		trimmedTags := strings.ReplaceAll(belaur.Cfg.WorkerTags, " ", "")
		tags = strings.Split(trimmedTags, ",")
	}
	tags = findLocalBinaries(tags)

	// Check if this worker has been already registered at this Bhojpur Belaur primary instance
	var regResp *api.RegisterResponse
	clientTLS, err := a.generateClientTLSCreds()
	if err != nil {
		// If there is an error, no matter if no certificates exist or
		// we cannot load them, we try the registration process to register
		// the worker again.
		regResp, err = api.RegisterWorker(belaur.Cfg.WorkerHostURL, belaur.Cfg.WorkerSecret, belaur.Cfg.WorkerName, tags)
		if err != nil {
			return nil, fmt.Errorf("failed to register worker: %s", err.Error())
		}

		// The registration process was successful.
		belaur.Cfg.Logger.Debug("Worker has been successfully registered at the Bhojpur Belaur primary instance")

		// Decode received certificates
		cert, err := base64.StdEncoding.DecodeString(regResp.Cert)
		if err != nil {
			return nil, fmt.Errorf("cannot decode certificate: %s", err.Error())
		}
		key, err := base64.StdEncoding.DecodeString(regResp.Key)
		if err != nil {
			return nil, fmt.Errorf("cannot decode key: %s", err.Error())
		}
		caCert, err := base64.StdEncoding.DecodeString(regResp.CACert)
		if err != nil {
			return nil, fmt.Errorf("cannot decode ca cert: %s", err.Error())
		}

		// Store received certificates locally
		if err = ioutil.WriteFile(a.certFile, cert, 0600); err != nil {
			return nil, fmt.Errorf("cannot write cert to disk: %s", err.Error())
		}
		if err = ioutil.WriteFile(a.keyFile, key, 0600); err != nil {
			return nil, fmt.Errorf("cannot write key to disk: %s", err.Error())
		}
		if err = ioutil.WriteFile(a.caCertFile, caCert, 0600); err != nil {
			return nil, fmt.Errorf("cannot write ca cert to disk: %s", err.Error())
		}

		// Update the client TLS object
		clientTLS, err = a.generateClientTLSCreds()
		if err != nil {
			return nil, fmt.Errorf("failed to generate TLS credentials: %s", err.Error())
		}
	}

	// Setup worker object
	worker := &belaur.Worker{}

	// Worker has been registered
	if regResp != nil {
		worker.UniqueID = regResp.UniqueID
	} else {
		// Load existing worker id from store
		w, err := a.store.WorkerGetAll()
		if err != nil {
			return nil, fmt.Errorf("failed to load worker id from store: %s", err.Error())
		}

		// Only one worker obj should exist
		if len(w) != 1 {
			return nil, fmt.Errorf("failed to load worker obj from store. Expected one object but got %d", len(w))
		}

		// Set unique id from store
		worker.UniqueID = w[0].UniqueID
	}

	// Set tags
	worker.Tags = tags

	// Setup information object about the current agent
	a.self = &pb.WorkerInstance{
		UniqueId: worker.UniqueID,
		Tags:     worker.Tags,
	}

	// Prevent odd/old data is still in our store.
	if err = a.store.WorkerDeleteAll(); err != nil {
		return nil, fmt.Errorf("failed to clean up worker bucket in store: %s", err.Error())
	}

	// Store updated worker object
	if err = a.store.WorkerPut(worker); err != nil {
		return nil, fmt.Errorf("failed to store worker obj in store: %s", err.Error())
	}

	return clientTLS, nil
}

// scheduleWork is a periodic Go routine which continuously pulls work
// from the Bhojpur Belaur primary instance. In case the pipeline is not available
// on this machine, the pipeline will be downloaded from the Bhojpur Belaur primary instance.
func (a *Agent) scheduleWork() {
	// Print info output
	belaur.Cfg.Logger.Trace("try to pull work from Bhojpur Belaur primary instance...")

	// Set available worker slots. Primary instance decides if worker needs work.
	a.self.WorkerSlots = a.scheduler.GetFreeWorkers()

	// Setup context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), (12*schedulerTickerSeconds)*time.Second)
	ctx = metadata.AppendToOutgoingContext(ctx, idMDKey, a.self.UniqueId)
	defer cancel()

	// Get actual work from remote Bhojpur Belaur instance
	stream, err := a.client.GetWork(ctx, a.self)
	if err != nil {
		belaur.Cfg.Logger.Error("failed to retrieve work from remote instance", "error", err.Error())
		return
	}

	// Read until the stream was closed
	workCounter := 0
	for {
		pipelineRunPB, err := stream.Recv()

		// Stream was closed
		if err == io.EOF {
			break
		}
		if err != nil {
			belaur.Cfg.Logger.Error("failed to stream work from remote instance", "error", err.Error())

			// In case the worker has been deregistered at the primary instance, we need to stop
			// the agent here.
			if strings.Contains(err.Error(), "worker is not registered") {
				// Since the worker has been deregistered, we should make sure that we
				// delete the existing certificates for security reasons.
				if err := os.Remove(a.certFile); err != nil {
					belaur.Cfg.Logger.Error("failed to remove cert file", "error", err)
				}
				if err := os.Remove(a.keyFile); err != nil {
					belaur.Cfg.Logger.Error("failed to remove key file", "error", err)
				}
				if err := os.Remove(a.caCertFile); err != nil {
					belaur.Cfg.Logger.Error("failed to remove ca cert file", "error", err)
				}

				// Send quit signal
				a.exitChan <- syscall.SIGTERM
			}
			return
		}

		belaur.Cfg.Logger.Info("received work from Bhojpur Belaur primary instance...")
		workCounter++

		// Convert protobuf pipeline run to internal struct
		pipelineRun := &belaur.PipelineRun{
			UniqueID:     pipelineRunPB.UniqueId,
			ID:           int(pipelineRunPB.Id),
			Status:       belaur.PipelineRunStatus(pipelineRunPB.Status),
			PipelineID:   int(pipelineRunPB.PipelineId),
			ScheduleDate: time.Unix(pipelineRunPB.ScheduleDate, 0),
			PipelineType: belaur.PipelineType(pipelineRunPB.PipelineType),
			Docker:       pipelineRunPB.Docker,
		}

		// Convert jobs
		jobsMap := make(map[uint32]*belaur.Job)
		for _, job := range pipelineRunPB.Jobs {
			j := &belaur.Job{
				ID:          job.UniqueId,
				Title:       job.Title,
				Status:      belaur.JobStatus(job.Status),
				Description: job.Description,
			}
			jobsMap[j.ID] = j
			pipelineRun.Jobs = append(pipelineRun.Jobs, j)

			// Arguments
			j.Args = make([]*belaur.Argument, 0, len(job.Args))
			for _, arg := range job.Args {
				a := &belaur.Argument{
					Description: arg.Description,
					Type:        arg.Type,
					Key:         arg.Key,
					Value:       arg.Value,
				}
				j.Args = append(j.Args, a)
			}
		}

		// Convert dependencies
		for _, pbJob := range pipelineRunPB.Jobs {
			// Get job
			j := jobsMap[pbJob.UniqueId]

			// Iterate all dependencies
			j.DependsOn = make([]*belaur.Job, 0, len(pbJob.DependsOn))
			for _, depJob := range pbJob.DependsOn {
				// Get dependency
				depJ := jobsMap[depJob.UniqueId]

				// Set dependency
				j.DependsOn = append(j.DependsOn, depJ)
			}
		}

		// Get pipeline binary name and SHA256SUM
		pipelineName := pipelinehelper.AppendTypeToName(pipelineRunPB.PipelineName, belaur.PipelineType(pipelineRunPB.PipelineType))
		pipelineSHA256SUM := pipelineRunPB.ShaSum

		// Setup reschedule of pipeline in case something goes wrong
		reschedulePipeline := func() {
			pipelineRunPB.Status = string(belaur.RunReschedule)
			if _, err := a.client.UpdateWork(ctx, pipelineRunPB); err != nil {
				belaur.Cfg.Logger.Error("failed to reschedule work at primary instance", "error", err)
			}
		}

		// Check if the binary is already stored locally
		pipelineFullPath := filepath.Join(belaur.Cfg.PipelinePath, pipelineName)
		if _, err := os.Stat(pipelineFullPath); err != nil {
			// Download binary from remote Bhojpur Belaur instance
			if err = a.streamBinary(pipelineRunPB, pipelineFullPath); err != nil {
				belaur.Cfg.Logger.Error("failed to download pipeline binary from remote instance", "error", err.Error(), "pipelinerun", pipelineRunPB)
				reschedulePipeline()
				return
			}
		}

		// Validate SHA256 sum to make sure the integrity is provided
		sha256Sum, err := filehelper.GetSHA256Sum(pipelineFullPath)
		if err != nil {
			belaur.Cfg.Logger.Error("failed to determine SHA256Sum of pipeline file", "error", err.Error(), "pipelinerun", pipelineRunPB)
			reschedulePipeline()
			return
		}

		if !bytes.Equal(sha256Sum, pipelineSHA256SUM) {
			if !a.compareSHAs(pipelineRun.PipelineID, sha256Sum, pipelineSHA256SUM) {
				belaur.Cfg.Logger.Debug("sha mismatch... attempting to re-download the binary")
				// A possible scenario is that the pipeline has been updated and the old binary still exists here.
				// Let us try to delete the binary and re-download the pipeline.
				if err := os.Remove(pipelineFullPath); err != nil {
					belaur.Cfg.Logger.Error("failed to remove inconsistent pipeline binary", "error", err.Error(), "pipelinerun", pipelineRunPB)
					reschedulePipeline()
					return
				}
				if err := a.streamBinary(pipelineRunPB, pipelineFullPath); err != nil {
					belaur.Cfg.Logger.Error("failed to download pipeline binary from remote instance", "error", err.Error(), "pipelinerun", pipelineRunPB)
					reschedulePipeline()
					return
				}

				// Validate SHA256 sum again to make sure the integrity is provided
				sha256Sum, err := filehelper.GetSHA256Sum(pipelineFullPath)
				if err != nil {
					belaur.Cfg.Logger.Error("failed to determine SHA256Sum of pipeline file", "error", err.Error(), "pipelinerun", pipelineRunPB)
					reschedulePipeline()
					return
				}
				if !bytes.Equal(sha256Sum, pipelineSHA256SUM) {
					belaur.Cfg.Logger.Error("pipeline binary SHA256Sum mismatch", "pipelinerun", pipelineRunPB)
					reschedulePipeline()
					return
				}
			}
		}

		// Check if the pipeline has been already stored
		var pipeline *belaur.Pipeline
		pipeline, err = a.store.PipelineGet(pipelineRun.PipelineID)
		if err != nil {
			belaur.Cfg.Logger.Error("failed to load pipeline from store", "error", err.Error(), "pipelinerun", pipelineRunPB)
			reschedulePipeline()
			return
		}
		if pipeline == nil {
			// Create a new pipeline object
			pipelineType := belaur.PipelineType(pipelineRunPB.PipelineType)
			pipeline = &belaur.Pipeline{
				ID:       pipelineRun.PipelineID,
				Name:     pipelineRunPB.PipelineName,
				Type:     pipelineType,
				ExecPath: pipelineFullPath,
				Jobs:     pipelineRun.Jobs,
			}
		}

		// Doesn't matter if we created a new pipeline object or load it from store,
		// we always set the correct SHA256Sum to make sure this is always the newest.
		pipeline.SHA256Sum = pipelineSHA256SUM

		// Let us try to start the plugin and receive all implemented jobs
		if err = a.scheduler.SetPipelineJobs(pipeline); err != nil {
			if !strings.Contains(err.Error(), "exec format error") {
				belaur.Cfg.Logger.Error("cannot get pipeline jobs", "error", err.Error(), "pipelinerun", pipelineRunPB)
				reschedulePipeline()
				return
			}
			belaur.Cfg.Logger.Info("pipeline in a different format than worker; attempting to rebuild...")
			// Try rebuilding the pipeline...
			if err := os.Remove(pipelineFullPath); err != nil {
				belaur.Cfg.Logger.Error("failed to remove pipeline binary", "error", err.Error(), "pipelinerun", pipelineRunPB)
				reschedulePipeline()
				return
			}
			err = a.rebuildWorkerBinary(ctx, pipeline)
			if err != nil {
				belaur.Cfg.Logger.Error("failed to rebuild pipeline for worker", "error", err.Error(), "pipelinerun", pipelineRunPB)
				reschedulePipeline()
				return
			}

			workerSHA256Sum, err := filehelper.GetSHA256Sum(pipelineFullPath)
			if err != nil {
				belaur.Cfg.Logger.Error("failed to determine SHA256Sum of pipeline file", "error", err.Error(), "pipelinerun", pipelineRunPB)
				reschedulePipeline()
				return
			}

			shaPair := belaur.SHAPair{
				Original:   pipelineSHA256SUM,
				Worker:     workerSHA256Sum,
				PipelineID: pipelineRun.PipelineID,
			}

			err = a.store.UpsertSHAPair(shaPair)
			if err != nil {
				belaur.Cfg.Logger.Error("failed to upsert new sha pair", "error", err.Error(), "pipelinerun", pipelineRunPB)
				reschedulePipeline()
				return
			}

			// Try setting the pipeline jobs again.
			if err = a.scheduler.SetPipelineJobs(pipeline); err != nil {
				belaur.Cfg.Logger.Error("cannot get pipeline jobs", "error", err.Error(), "pipelinerun", pipelineRunPB)
				reschedulePipeline()
				return
			}
		}
		pipelineRun.Jobs = pipeline.Jobs
		// Store pipeline
		if err = a.store.PipelinePut(pipeline); err != nil {
			belaur.Cfg.Logger.Error("failed to store pipeline in store", "error", err.Error(), "pipelinerun", pipelineRunPB)
			reschedulePipeline()
			return
		}

		// The scheduler picks only runs up which are in state "NotScheduled".
		// Since the scheduler from the Bhojpur Belaur primary instance set the state already to "scheduled",
		// we have to reset the state here so that the scheduler will pick it up.
		pipelineRun.Status = belaur.RunNotScheduled

		// Store finally the pipeline run
		if err = a.store.PipelinePutRun(pipelineRun); err != nil {
			belaur.Cfg.Logger.Error("failed to store pipeline run in store", "error", err.Error(), "pipelinerun", pipelineRunPB)
			reschedulePipeline()
			return
		}
	}

	// Check if we received work at all
	if workCounter == 0 {
		belaur.Cfg.Logger.Trace("got no work from Bhojpur Belaur primary instance. Will try it again after a while...")
	}
}

// compareSHAs compares shas of the binaries with possibly stored sha pairs. First it compares the original if they match
// second it compares the local sha with the new one that the worker possibly rebuilt. If there is no entry,
// we return false, because we don't know anything about the sha.
func (a *Agent) compareSHAs(id int, sha256Sum, pipelineSHA256SUM []byte) bool {
	ok, shaPair, err := a.store.GetSHAPair(id)
	if err != nil {
		belaur.Cfg.Logger.Error("failed to get sha pair from memdb", "error", err.Error())
		return false
	}

	if !ok {
		belaur.Cfg.Logger.Debug("no record found for pipeline. skipping this check")
		return false
	}
	belaur.Cfg.Logger.Debug("record found for pipeline... comparing.")
	return bytes.Equal(sha256Sum, shaPair.Worker) &&
		bytes.Equal(pipelineSHA256SUM, shaPair.Original)
}

func (a *Agent) rebuildWorkerBinary(ctx context.Context, pipeline *belaur.Pipeline) error {
	pCreate := &belaur.CreatePipeline{}
	pCreate.ID = security.GenerateRandomUUIDV5()
	pCreate.Pipeline = *pipeline

	repo, err := a.client.GetGitRepo(ctx, &pb.PipelineID{Id: int64(pipeline.ID)})
	if err != nil {
		return err
	}

	// Unfortunately, since pb.GitRepo has extra gRPC fields on it
	// we can't use belaur.GitRepo(repo) here to convert immediately.
	gitRepo := belaur.GitRepo{}
	gitRepo.Username = repo.Username
	gitRepo.Password = repo.Password

	pk := belaur.PrivateKey{}
	if repo.PrivateKey != nil {
		pk.Password = repo.PrivateKey.Password
		pk.Username = repo.PrivateKey.Username
		pk.Key = repo.PrivateKey.Key
	}

	gitRepo.PrivateKey = pk
	gitRepo.URL = repo.Url
	gitRepo.SelectedBranch = repo.SelectedBranch
	pCreate.Pipeline.Repo = &gitRepo

	a.pipelineService.CreatePipeline(pCreate)
	if pCreate.StatusType == belaur.CreatePipelineFailed {
		return fmt.Errorf("error while creating pipeline: %s", pCreate.Output)
	}

	pipeline = &pCreate.Pipeline
	if err = a.scheduler.SetPipelineJobs(pipeline); err != nil {
		return err
	}

	return nil
}

// streamBinary streams the binary in chunks from the remote instance to the given path.
func (a *Agent) streamBinary(pipelineRunPB *pb.PipelineRun, pipelinePath string) error {
	// Create a context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	ctx = metadata.AppendToOutgoingContext(ctx, idMDKey, a.self.UniqueId)
	defer cancel()

	// Initiate streaming
	stream, err := a.client.StreamBinary(ctx, pipelineRunPB)
	if err != nil {
		belaur.Cfg.Logger.Error("failed to stream pipeline binary from remote instance", "error", err.Error(), "pipelinerun", pipelineRunPB)
		return err
	}

	// Open output file
	pipelineFile, err := os.Create(pipelinePath)
	if err != nil {
		belaur.Cfg.Logger.Error("failed to create new pipeline binary file during remote binary stream", "error", err.Error(), "pipelinerun", pipelineRunPB)
		return err
	}
	defer pipelineFile.Close()

	// Read whole stream
	for {
		streamChunk, err := stream.Recv()

		// Check if stream was closed remotely
		if err == io.EOF {
			break
		}
		if err != nil {
			belaur.Cfg.Logger.Error("failed to stream full pipeline binary from remote instance", "error", err.Error(), "pipelinerun", pipelineRunPB)
			return err
		}

		// Write chunk to file
		if _, err := pipelineFile.Write(streamChunk.Chunk); err != nil {
			belaur.Cfg.Logger.Error("failed to write chunk to local disk during stream binary", "error", err.Error(), "pipelinerun", pipelineRunPB)
			return err
		}
	}

	// Set pipeline executable rights
	return os.Chmod(pipelinePath, belaur.ExecutablePermission)
}

// updateWork is periodically called and it is used to
// send new information about a pipeline run to the Bhojpur Belaur primary instance.
func (a *Agent) updateWork() {
	// Read all pipeline runs from the store. The number of pipeline runs
	// should be relatively low since we delete pipeline runs after successful
	// execution.
	runs, err := a.store.PipelineGetAllRuns()
	if err != nil {
		belaur.Cfg.Logger.Error("failed to load pipeline runs from store", "error", err.Error())
		return
	}

	// Create context with timeout
	ctx, cancel := context.WithTimeout(context.Background(), (updateTickerSeconds*3)*time.Second)
	ctx = metadata.AppendToOutgoingContext(ctx, idMDKey, a.self.UniqueId)
	defer cancel()

	// Send all pipeline runs to the remote primary instance
	for _, run := range runs {
		// Transform to protobuf struct
		runPB := &pb.PipelineRun{
			Id:           int64(run.ID),
			UniqueId:     run.UniqueID,
			Status:       string(run.Status),
			PipelineId:   int64(run.PipelineID),
			ScheduleDate: run.ScheduleDate.Unix(),
			StartDate:    run.StartDate.Unix(),
			FinishDate:   run.FinishDate.Unix(),
			Docker:       run.Docker,
		}

		// Transform pipeline run jobs
		jobsMap := make(map[uint32]*pb.Job)
		for _, job := range run.Jobs {
			j := &pb.Job{
				UniqueId:    job.ID,
				Title:       job.Title,
				Status:      string(job.Status),
				Description: job.Description,
			}
			runPB.Jobs = append(runPB.Jobs, j)

			// Fill helper map for job dependency search
			jobsMap[j.UniqueId] = j

			// Convert arguments
			j.Args = make([]*pb.Argument, 0, len(job.Args))
			for _, arg := range job.Args {
				a := &pb.Argument{
					Description: arg.Description,
					Type:        arg.Type,
					Key:         arg.Key,
					Value:       arg.Value,
				}
				j.Args = append(j.Args, a)
			}
		}

		// Convert dependencies
		for _, job := range run.Jobs {
			// Get job
			j := jobsMap[job.ID]

			// Iterate all dependencies
			j.DependsOn = make([]*pb.Job, 0, len(job.DependsOn))
			for _, depJob := range job.DependsOn {
				// Get dependency
				depJ := jobsMap[depJob.ID]

				// Set dependency
				j.DependsOn = append(j.DependsOn, depJ)
			}
		}

		// Ship pipeline logs if exists
		if err := a.shipPipelineLogs(ctx, &run); err != nil {
			return
		}

		// Send to remote instance
		if _, err := a.client.UpdateWork(ctx, runPB); err != nil {
			belaur.Cfg.Logger.Error("failed to send update information to remote instance", "error", err.Error())
			return
		}

		// Remove pipeline run from store when the state is finalized
		if run.Status == belaur.RunFailed || run.Status == belaur.RunSuccess || run.Status == belaur.RunCancelled || run.Status == belaur.RunReschedule {
			if err = a.store.PipelineRunDelete(run.UniqueID); err != nil {
				belaur.Cfg.Logger.Error("failed to remove pipeline run from store", "error", err.Error(), "pipelinerun", run)
			}
		}
	}
}

// shipPipelineLogs ships pipeline from the given pipeline run to the primary instance.
// It will only return an error when an error occurred during transmission, not when
// no logs for a pipeline are not existent.
func (a *Agent) shipPipelineLogs(ctx context.Context, run *belaur.PipelineRun) error {
	// Check if log file exists for pipeline run.
	// If the file does not exist, we simply skip the shipping.
	logFilePath := filepath.Join(belaur.Cfg.WorkspacePath, strconv.Itoa(run.PipelineID), strconv.Itoa(run.ID), belaur.LogsFolderName, belaur.LogsFileName)
	if _, err := os.Stat(logFilePath); err != nil {
		return nil
	}

	// Open file handle
	file, err := os.Open(logFilePath)
	if err != nil {
		belaur.Cfg.Logger.Warn("failed to open pipeline run log file via shipPipelineLogs", "error", err.Error(), "pipelinerun", run)
		return err
	}
	defer file.Close()

	// Open streaming session to primary instance
	stream, err := a.client.StreamLogs(ctx)
	if err != nil {
		belaur.Cfg.Logger.Warn("failed to open stream session to primary instance to ship logs via shipPipelineLogs", "error", err.Error(), "pipelinerun", run)
		return err
	}

	chunk := &pb.LogChunk{
		PipelineId: int64(run.PipelineID),
		RunId:      int64(run.ID),
	}
	buffer := make([]byte, chunkSize)
	for {
		bytesread, err := file.Read(buffer)

		// Check for errors
		if err != nil {
			if err != io.EOF {
				belaur.Cfg.Logger.Warn("error occurred during pipeline run log disk read", "error", err.Error(), "pipelinerun", run)
				return err
			}
			break
		}

		// Set bytes
		chunk.Chunk = buffer[:bytesread]

		// Stream it to primary instance
		if err = stream.Send(chunk); err != nil {
			belaur.Cfg.Logger.Error("failed to stream log chunk to primary instance", "error", err.Error(), "pipelinerun", run)
			return err
		}
	}
	if err = stream.CloseSend(); err != nil {
		belaur.Cfg.Logger.Warn("failed to safely close gRPC connection via updatework", "error", err.Error())
		return err
	}
	return nil
}

// generateClientTLSCreds checks if certificates exist in the home directory.
// It will load the certificates and generates TLS creds for mTLS connection.
func (a *Agent) generateClientTLSCreds() (credentials.TransportCredentials, error) {
	// Check if all certs exist
	if _, err := os.Stat(a.certFile); os.IsNotExist(err) {
		return nil, err
	}
	if _, err := os.Stat(a.keyFile); os.IsNotExist(err) {
		return nil, err
	}
	if _, err := os.Stat(a.caCertFile); os.IsNotExist(err) {
		return nil, err
	}

	// Load client key pair
	certs, err := tls.LoadX509KeyPair(a.certFile, a.keyFile)
	if err != nil {
		return nil, err
	}

	certPool := x509.NewCertPool()
	caCert, err := ioutil.ReadFile(a.caCertFile)
	if err != nil {
		return nil, err
	}

	// Add certificate to cert pool
	ok := certPool.AppendCertsFromPEM(caCert)
	if !ok {
		return nil, errors.New("cannot append ca cert to cert pool")
	}

	return credentials.NewTLS(&tls.Config{
		ServerName:   defaultHostname,
		Certificates: []tls.Certificate{certs},
		RootCAs:      certPool,
	}), nil
}
