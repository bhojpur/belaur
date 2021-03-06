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
	"errors"
	"fmt"
	"strconv"
	"strings"
	"unicode"

	belaur "github.com/bhojpur/belaur"
	"github.com/bhojpur/belaur/pkg/security"
	"github.com/bhojpur/belaur/pkg/services"
)

const (
	// Percent of pipeline creation progress after git clone
	pipelineCloneStatus = 25

	// Percent of pipeline creation progress after compile process done
	pipelineCompileStatus = 50

	// Percent of pipeline creation progress after validation
	pipelineValidateStatus = 75

	// Completed percent progress
	pipelineCompleteStatus = 100

	// Split char to separate path from pipeline and name
	pipelinePathSplitChar = "/"
)

var (
	// errPathLength is a validation error during pipeline name input
	errPathLength = errors.New("name of pipeline is empty or one of the path elements length exceeds 50 characters")

	// errPipelineNameInUse is thrown when a pipelines name is already in use
	errPipelineNameInUse = errors.New("pipeline name is already in use")

	// errPipelineNameInvalid is thrown when the pipeline name contains invalid characters
	errPipelineNameInvalid = errors.New("must match [A-z][0-9][-][_][ ]")
)

// CreatePipeline is the main function which executes step by step the creation
// of a plugin.
// After each step, the status is written to store and can be retrieved via API.
func (s *BelaurPipelineService) CreatePipeline(p *belaur.CreatePipeline) {
	gitToken := p.GitHubToken
	p.GitHubToken = ""
	storeService, _ := services.StorageService()
	// Define build process for the given type
	bP := newBuildPipeline(p.Pipeline.Type)
	if bP == nil {
		// Pipeline type is not supported
		p.StatusType = belaur.CreatePipelineFailed
		p.Output = fmt.Sprintf("create pipeline failed. Pipeline type is not supported %s is not supported", p.Pipeline.Type)
		_ = storeService.CreatePipelinePut(p)
		return
	}

	// Setup environment before cloning repo and command
	err := bP.PrepareEnvironment(p)
	if err != nil {
		p.StatusType = belaur.CreatePipelineFailed
		p.Output = fmt.Sprintf("cannot prepare build: %s", err.Error())
		_ = storeService.CreatePipelinePut(p)
		return
	}

	// Clone git repo
	err = gitCloneRepo(p.Pipeline.Repo)
	if err != nil {
		p.StatusType = belaur.CreatePipelineFailed
		p.Output = fmt.Sprintf("cannot prepare build: %s", err.Error())
		_ = storeService.CreatePipelinePut(p)
		return
	}

	// Update status of our pipeline build
	p.Status = pipelineCloneStatus
	err = storeService.CreatePipelinePut(p)
	if err != nil {
		p.StatusType = belaur.CreatePipelineFailed
		belaur.Cfg.Logger.Error("cannot put create pipeline into store", "error", err.Error())
		return
	}

	// Run compile process
	err = bP.ExecuteBuild(p)
	if err != nil {
		p.StatusType = belaur.CreatePipelineFailed
		_ = storeService.CreatePipelinePut(p)
		return
	}

	// Update status of our pipeline build
	p.Status = pipelineCompileStatus
	err = storeService.CreatePipelinePut(p)
	if err != nil {
		p.StatusType = belaur.CreatePipelineFailed
		belaur.Cfg.Logger.Error("cannot put create pipeline into store", "error", err.Error())
		return
	}

	// Run update if needed
	err = updatePipeline(&p.Pipeline)
	if err != nil {
		p.StatusType = belaur.CreatePipelineFailed
		p.Output = fmt.Sprintf("cannot update pipeline: %s", err.Error())
		_ = storeService.CreatePipelinePut(p)
		return
	}

	// Try to get pipeline jobs to check if this pipeline is valid.
	if err = s.deps.Scheduler.SetPipelineJobs(&p.Pipeline); err != nil {
		p.StatusType = belaur.CreatePipelineFailed
		p.Output = fmt.Sprintf("cannot validate pipeline: %s", err.Error())
		_ = storeService.CreatePipelinePut(p)
		return
	}

	// Update status of our pipeline build
	p.Status = pipelineValidateStatus
	err = storeService.CreatePipelinePut(p)
	if err != nil {
		belaur.Cfg.Logger.Error("cannot put create pipeline into store", "error", err.Error())
		return
	}

	p.Pipeline.TriggerToken = security.GenerateRandomUUIDV5()

	// Save the generated pipeline data
	err = bP.SavePipeline(&p.Pipeline)
	if err != nil {
		p.StatusType = belaur.CreatePipelineFailed
		p.Output = fmt.Sprintf("failed to save the created pipeline: %s", err.Error())
		_ = storeService.CreatePipelinePut(p)
		return
	}

	// Copy compiled binary to plugins folder which is the final step
	err = bP.CopyBinary(p)
	if err != nil {
		p.StatusType = belaur.CreatePipelineFailed
		p.Output = fmt.Sprintf("cannot copy compiled binary: %s", err.Error())
		_ = storeService.CreatePipelinePut(p)
		return
	}

	// Set create pipeline status to complete
	p.Status = pipelineCompleteStatus
	p.StatusType = belaur.CreatePipelineSuccess
	err = storeService.CreatePipelinePut(p)
	if err != nil {
		p.StatusType = belaur.CreatePipelineFailed
		belaur.Cfg.Logger.Error("cannot put create pipeline into store", "error", err.Error())
		return
	}

	if !belaur.Cfg.Poll && len(gitToken) > 0 {
		// if there is a githubtoken provided, that means that a webhook was requested to be added.
		id := strconv.Itoa(p.Pipeline.ID)
		err = createGithubWebhook(gitToken, p.Pipeline.Repo, id, nil)
		if err != nil {
			belaur.Cfg.Logger.Error("error while creating webhook for repository", "error", err.Error())
			return
		}
	}
}

// ValidatePipelineName validates a given pipeline name and
// returns the correct error back.
func ValidatePipelineName(pName string) error {

	valid := func(r rune) bool {
		return unicode.IsDigit(r) || unicode.IsLetter(r) || unicode.IsSpace(r) || r == '-' || r == '_'
	}
	// Note, this is faster than regex.
	for _, c := range pName {
		if !valid(c) {
			return errPipelineNameInvalid
		}
	}

	// The name could contain a path. Split it up.
	path := strings.Split(pName, pipelinePathSplitChar)

	// Iterate all objects.
	for _, s := range path {
		// Length should be correct.
		if len(s) < 1 || len(s) > 50 {
			return errPathLength
		}

		// Check if pipeline name is already in use.
		for _, activePipeline := range GlobalActivePipelines.GetAll() {
			if strings.EqualFold(s, activePipeline.Name) {
				return errPipelineNameInUse
			}
		}
	}
	return nil
}
