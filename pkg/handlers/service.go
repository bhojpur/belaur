package handlers

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
	"github.com/bhojpur/belaur/pkg/providers"
	"github.com/bhojpur/belaur/pkg/providers/pipelines"
	"github.com/bhojpur/belaur/pkg/providers/workers"
	"github.com/bhojpur/belaur/pkg/security"
	"github.com/bhojpur/belaur/pkg/security/rbac"
	"github.com/bhojpur/belaur/pkg/store"
	"github.com/bhojpur/belaur/pkg/workers/pipeline"
	"github.com/bhojpur/belaur/pkg/workers/scheduler/service"
)

// Dependencies define dependencies for this service.
type Dependencies struct {
	Scheduler        service.BelaurScheduler
	PipelineService  pipeline.Servicer
	PipelineProvider pipelines.PipelineProviderer
	UserProvider     providers.UserProvider
	RBACProvider     providers.RBACProvider
	WorkerProvider   workers.WorkerProviderer
	Certificate      security.CAAPI
	RBACService      rbac.Service
	Store            store.BelaurStore
}

// BelaurHandler defines handler functions throughout Bhojpur Belaur.
type BelaurHandler struct {
	deps Dependencies
}

// NewBelaurHandler creates a new handler service with the required dependencies.
func NewBelaurHandler(deps Dependencies) *BelaurHandler {
	return &BelaurHandler{deps: deps}
}
