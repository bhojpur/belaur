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
	"github.com/labstack/echo/v4"

	"github.com/bhojpur/belaur/pkg/store"
	"github.com/bhojpur/belaur/pkg/workers/pipeline"
	"github.com/bhojpur/belaur/pkg/workers/scheduler/service"
)

// Dependencies define providers and services which this service needs.
type Dependencies struct {
	Scheduler       service.BelaurScheduler
	PipelineService pipeline.Servicer
	SettingsStore   store.SettingsStore
}

// PipelineProvider is a provider for all pipeline related operations.
type PipelineProvider struct {
	deps Dependencies
}

// PipelineProviderer defines functionality which this provider provides.
// These are used by the handler service.
type PipelineProviderer interface {
	PipelineGitLSRemote(c echo.Context) error
	CreatePipeline(c echo.Context) error
	CreatePipelineGetAll(c echo.Context) error
	PipelineNameAvailable(c echo.Context) error
	PipelineGet(c echo.Context) error
	PipelineGetAll(c echo.Context) error
	PipelineUpdate(c echo.Context) error
	PipelinePull(c echo.Context) error
	PipelineDelete(c echo.Context) error
	PipelineTrigger(c echo.Context) error
	PipelineResetToken(c echo.Context) error
	PipelineTriggerAuth(c echo.Context) error
	PipelineStart(c echo.Context) error
	PipelineGetAllWithLatestRun(c echo.Context) error
	PipelineCheckPeriodicSchedules(c echo.Context) error
	PipelineStop(c echo.Context) error
	PipelineRunGet(c echo.Context) error
	PipelineGetAllRuns(c echo.Context) error
	PipelineGetLatestRun(c echo.Context) error
	GetJobLogs(c echo.Context) error
	GitWebHook(c echo.Context) error
	SettingsPollOn(c echo.Context) error
	SettingsPollOff(c echo.Context) error
	SettingsPollGet(c echo.Context) error
}

// NewPipelineProvider creates a new provider with the needed dependencies.
func NewPipelineProvider(deps Dependencies) *PipelineProvider {
	return &PipelineProvider{deps: deps}
}
