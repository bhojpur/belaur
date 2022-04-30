package docker

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
	"context"
	"fmt"
	"io"
	"time"

	belaur "github.com/bhojpur/belaur"
	"github.com/docker/docker/api/types"
	"github.com/docker/docker/api/types/container"
	"github.com/docker/docker/client"

	"github.com/bhojpur/belaur/pkg/security"
)

// Worker represents the data structure of a docker worker.
type Worker struct {
	Host          string `json:"host"`
	WorkerID      string `json:"worker_id"`
	ContainerID   string `json:"container_id"`
	PipelineRunID string `json:"pipeline_run_id"`
}

// NewDockerWorker initiates a new worker instance.
func NewDockerWorker(host string, pipelineRunID string) *Worker {
	return &Worker{Host: host, PipelineRunID: pipelineRunID}
}

// SetupDockerWorker starts a Bhojpur Belaur worker inside a docker container, automatically
// connects it with this Bhojpur Belaur instance and sets a unique tag.
func (w *Worker) SetupDockerWorker(workerImage string, workerSecret string) error {
	// Generate a unique id for this worker
	w.WorkerID = security.GenerateRandomUUIDV5()

	// Setup docker client
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Minute)
	defer cancel()
	cli, err := client.NewClientWithOpts(client.WithHost(w.Host))
	if err != nil {
		belaur.Cfg.Logger.Error("failed to setup docker client", "error", err)
		return err
	}
	cli.NegotiateAPIVersion(ctx)

	// Define small helper function which creates the docker container
	createContainer := func() (container.ContainerCreateCreatedBody, error) {
		return cli.ContainerCreate(ctx, &container.Config{
			Image: workerImage,
			Env: []string{
				"BELAUR_WORKER_HOST_URL=" + belaur.Cfg.DockerWorkerHostURL,
				"BELAUR_MODE=worker",
				"BELAUR_WORKER_GRPC_HOST_URL=" + belaur.Cfg.DockerWorkerGRPCHostURL,
				"BELAUR_WORKER_TAGS=" + fmt.Sprintf("%s,dockerworker", w.WorkerID),
				"BELAUR_WORKER_SECRET=" + workerSecret,
			},
		}, &container.HostConfig{}, nil, nil, "")
	}

	// Create container
	resp, err := createContainer()
	if err != nil {
		belaur.Cfg.Logger.Error("failed to create worker container", "error", err)
		belaur.Cfg.Logger.Info("try to pull docker image before trying it again...")

		// Pull worker image
		out, err := cli.ImagePull(ctx, workerImage, types.ImagePullOptions{})
		if err != nil {
			belaur.Cfg.Logger.Error("failed to pull worker image", "error", err)
			return err
		}
		defer out.Close()

		// Image will be only pulled when we read the stream.
		// We read it here even tho we are not interested into the output.
		belaur.Cfg.Logger.Info("pulling docker image. This may take a while...")
		buffer := make([]byte, 1024)
		for {
			_, err := out.Read(buffer)

			if err != nil {
				if err != io.EOF {
					belaur.Cfg.Logger.Error("failed to pull worker image", "error", err)
					return err
				}
				break
			}
		}
		belaur.Cfg.Logger.Info("finished pulling docker image. Continue with pipeline run...")

		// Try to create the container again
		resp, err = createContainer()
		if err != nil {
			belaur.Cfg.Logger.Error("failed to create worker container after pull", "error", err)
			return err
		}
	}

	// Start container
	if err := cli.ContainerStart(ctx, resp.ID, types.ContainerStartOptions{}); err != nil {
		belaur.Cfg.Logger.Error("failed to start worker container", "error", err)
		return err
	}

	// Store container id for later processing
	w.ContainerID = resp.ID
	return nil
}

// IsDockerWorkerRunning checks if the docker worker is running.
func (w *Worker) IsDockerWorkerRunning() bool {
	// Setup docker client
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.WithHost(w.Host))
	if err != nil {
		belaur.Cfg.Logger.Error("failed to setup docker client", "error", err)
		return false
	}
	cli.NegotiateAPIVersion(ctx)

	// Check if docker worker container is still running
	resp, err := cli.ContainerInspect(ctx, w.ContainerID)
	if err != nil {
		return false
	}

	if resp.State.Running {
		return true
	}
	return false
}

// KillDockerWorker disconnects the existing docker worker and
// kills the docker container.
func (w *Worker) KillDockerWorker() error {
	// Setup docker client
	ctx := context.Background()
	cli, err := client.NewClientWithOpts(client.WithHost(w.Host))
	if err != nil {
		belaur.Cfg.Logger.Error("failed to setup docker client", "error", err)
		return err
	}
	cli.NegotiateAPIVersion(ctx)

	// Kill container
	if err := cli.ContainerRemove(ctx, w.ContainerID, types.ContainerRemoveOptions{
		RemoveVolumes: true,
		Force:         true,
	}); err != nil {
		belaur.Cfg.Logger.Error("failed to remove docker worker", "error", err, "containerid", w.ContainerID)
		return err
	}
	return nil
}
