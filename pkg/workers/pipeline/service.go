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
	belaur "github.com/bhojpur/belaur"
	"github.com/bhojpur/belaur/pkg/workers/scheduler/service"
)

// Dependencies defines dependencies which this service needs to operate.
type Dependencies struct {
	Scheduler service.BelaurScheduler
}

// BelaurPipelineService defines a pipeline service provider providing pipeline related functions.
type BelaurPipelineService struct {
	deps Dependencies
}

// Servicer defines a scheduler service.
type Servicer interface {
	CreatePipeline(p *belaur.CreatePipeline)
	InitTicker()
	CheckActivePipelines()
	UpdateRepository(p *belaur.Pipeline) error
	UpdateAllCurrentPipelines()
	StartPoller() error
	StopPoller() error
}

// NewBelaurPipelineService creates a pipeline service with its required dependencies already wired up
func NewBelaurPipelineService(deps Dependencies) *BelaurPipelineService {
	return &BelaurPipelineService{deps: deps}
}
