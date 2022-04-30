package pipelines

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
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strconv"

	"github.com/labstack/echo/v4"

	belaur "github.com/bhojpur/belaur"
	"github.com/bhojpur/belaur/pkg/services"
	"github.com/bhojpur/belaur/pkg/workers/pipeline"
)

var (
	// errPipelineRunNotFound is thrown when a pipeline run was not found with the given id
	errPipelineRunNotFound = errors.New("pipeline run not found with the given id")
)

// jobLogs represents the json format which is returned
// by GetJobLogs.
type jobLogs struct {
	Log      string `json:"log"`
	Finished bool   `json:"finished"`
}

// PipelineRunGet returns details about a specific pipeline run.
// Required parameters are pipelineid and runid.
// @Summary Get Pipeline run.
// @Description Returns details about a specific pipeline run.
// @Tags pipelinerun
// @Accept plain
// @Produce json
// @Security ApiKeyAuth
// @Param pipelineid query string true "ID of the pipeline"
// @Param runid query string true "ID of the pipeline run"
// @Success 200 {object} belaur.PipelineRun
// @Failure 400 {string} string "Invalid pipeline or pipeline not found."
// @Failure 404 {string} string "Pipeline Run not found."
// @Failure 500 {string} string "Something went wrong while getting pipeline run."
// @Router /pipelinerun/{pipelineid}/{runid} [get]
func (pp *PipelineProvider) PipelineRunGet(c echo.Context) error {
	// Convert string to int because id is int
	storeService, _ := services.StorageService()
	pipelineID, err := strconv.Atoi(c.Param("pipelineid"))
	if err != nil {
		return c.String(http.StatusBadRequest, errInvalidPipelineID.Error())
	}

	// Convert string to int because id is int
	runID, err := strconv.Atoi(c.Param("runid"))
	if err != nil {
		return c.String(http.StatusBadRequest, errPipelineRunNotFound.Error())
	}

	// Find pipeline run in store
	pipelineRun, err := storeService.PipelineGetRunByPipelineIDAndID(pipelineID, runID)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	} else if pipelineRun == nil {
		return c.String(http.StatusNotFound, errPipelineRunNotFound.Error())
	}

	// Return pipeline run
	return c.JSON(http.StatusOK, pipelineRun)
}

// PipelineStop stops a running pipeline.
// @Summary Stop a pipeline run.
// @Description Stops a pipeline run.
// @Tags pipelinerun
// @Accept plain
// @Produce plain
// @Security ApiKeyAuth
// @Param pipelineid query string true "ID of the pipeline"
// @Param runid query string true "ID of the pipeline run"
// @Success 200 {string} string "pipeline successfully stopped"
// @Failure 400 {string} string "Invalid pipeline id or run id"
// @Failure 404 {string} string "Pipeline Run not found."
// @Router /pipelinerun/{pipelineid}/{runid}/stop [post]
func (pp *PipelineProvider) PipelineStop(c echo.Context) error {
	// Get parameters and validate
	pipelineID := c.Param("pipelineid")
	pipelineRunID := c.Param("runid")

	// Transform pipelineid to int
	p, err := strconv.Atoi(pipelineID)
	if err != nil {
		return c.String(http.StatusBadRequest, "invalid pipeline id given")
	}

	// Transform pipelinerunid to int
	r, err := strconv.Atoi(pipelineRunID)
	if err != nil {
		return c.String(http.StatusBadRequest, "invalid pipeline run id given")
	}

	// Look up pipeline for the given id
	var foundPipeline belaur.Pipeline
	for _, pipe := range pipeline.GlobalActivePipelines.GetAll() {
		if pipe.ID == p {
			foundPipeline = pipe
			break
		}
	}

	if foundPipeline.Name != "" {
		err = pp.deps.Scheduler.StopPipelineRun(&foundPipeline, r)
		if err != nil {
			return c.String(http.StatusBadRequest, err.Error())
		}
		return c.String(http.StatusOK, "pipeline successfully stopped")
	}

	// Pipeline not found
	return c.String(http.StatusNotFound, errPipelineNotFound.Error())
}

// PipelineGetAllRuns returns all runs about the given pipeline.
// @Summary Get all pipeline runs.
// @Description Returns all runs about the given pipeline.
// @Tags pipelinerun
// @Accept plain
// @Produce json
// @Security ApiKeyAuth
// @Param pipelineid query string true "ID of the pipeline"
// @Success 200 {array} belaur.PipelineRun "a list of pipeline runes"
// @Failure 400 {string} string "Invalid pipeline id"
// @Failure 500 {string} string "Error retrieving all pipeline runs."
// @Router /pipelinerun/{pipelineid} [get]
func (pp *PipelineProvider) PipelineGetAllRuns(c echo.Context) error {
	// Convert string to int because id is int
	storeService, _ := services.StorageService()
	pipelineID, err := strconv.Atoi(c.Param("pipelineid"))
	if err != nil {
		return c.String(http.StatusBadRequest, errInvalidPipelineID.Error())
	}

	// Get all runs by the given pipeline id
	runs, err := storeService.PipelineGetAllRunsByPipelineID(pipelineID)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, runs)
}

// PipelineGetLatestRun returns the latest run of a pipeline, given by id.
// @Summary Get latest pipeline runs.
// @Description Returns the latest run of a pipeline, given by id.
// @Tags pipelinerun
// @Accept plain
// @Produce json
// @Security ApiKeyAuth
// @Param pipelineid query string true "ID of the pipeline"
// @Success 200 {object} belaur.PipelineRun "the latest pipeline run"
// @Failure 400 {string} string "Invalid pipeline id"
// @Failure 500 {string} string "error getting latest run or cannot read pipeline run log file"
// @Router /pipelinerun/{pipelineid}/latest [get]
func (pp *PipelineProvider) PipelineGetLatestRun(c echo.Context) error {
	// Convert string to int because id is int
	storeService, _ := services.StorageService()
	pipelineID, err := strconv.Atoi(c.Param("pipelineid"))
	if err != nil {
		return c.String(http.StatusBadRequest, errInvalidPipelineID.Error())
	}

	// Get the latest run by the given pipeline id
	run, err := storeService.PipelineGetLatestRun(pipelineID)
	if err != nil {
		return c.String(http.StatusInternalServerError, err.Error())
	}

	return c.JSON(http.StatusOK, run)
}

// GetJobLogs returns logs from a pipeline run.
//
// Required parameters:
// pipelineid - Related pipeline id
// pipelinerunid - Related pipeline run id
// @Summary Get logs for pipeline run.
// @Description Returns logs from a pipeline run.
// @Tags pipelinerun
// @Accept plain
// @Produce json
// @Security ApiKeyAuth
// @Param pipelineid query string true "ID of the pipeline"
// @Param runid query string true "ID of the run"
// @Success 200 {object} jobLogs "logs"
// @Failure 400 {string} string "Invalid pipeline id or run id or pipeline not found"
// @Failure 500 {string} string "cannot read pipeline run log file"
// @Router /pipelinerun/{pipelineid}/{runid}/log [get]
func (pp *PipelineProvider) GetJobLogs(c echo.Context) error {
	// Get parameters and validate
	storeService, _ := services.StorageService()
	pipelineID := c.Param("pipelineid")
	pipelineRunID := c.Param("runid")

	// Transform pipelineid to int
	p, err := strconv.Atoi(pipelineID)
	if err != nil {
		return c.String(http.StatusBadRequest, "invalid pipeline id given")
	}

	// Transform pipelinerunid to int
	r, err := strconv.Atoi(pipelineRunID)
	if err != nil {
		return c.String(http.StatusBadRequest, "invalid pipeline run id given")
	}

	run, err := storeService.PipelineGetRunByPipelineIDAndID(p, r)
	if err != nil {
		return c.String(http.StatusBadRequest, "cannot find pipeline run with given pipeline id and pipeline run id")
	}

	// Create return object
	jL := jobLogs{}

	// Determine if job has been finished
	if run.Status == belaur.RunFailed || run.Status == belaur.RunSuccess || run.Status == belaur.RunCancelled {
		jL.Finished = true
	}

	// Check if log file exists
	logFilePath := filepath.Join(belaur.Cfg.WorkspacePath, pipelineID, pipelineRunID, belaur.LogsFolderName, belaur.LogsFileName)
	if _, err := os.Stat(logFilePath); err == nil {
		content, err := ioutil.ReadFile(logFilePath)
		if err != nil {
			return c.String(http.StatusInternalServerError, "cannot read pipeline run log file")
		}

		// Convert logs
		jL.Log = string(content)
	}

	// Return logs
	return c.JSON(http.StatusOK, jL)
}
