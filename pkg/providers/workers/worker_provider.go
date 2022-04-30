package workers

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

	"github.com/bhojpur/belaur/pkg/security"
	"github.com/bhojpur/belaur/pkg/workers/scheduler/service"
)

// Dependencies define dependencies which this service needs.
type Dependencies struct {
	Scheduler   service.BelaurScheduler
	Certificate security.CAAPI
}

// WorkerProvider has all the operations for a worker.
type WorkerProvider struct {
	deps Dependencies
}

// WorkerProviderer defines functionality which this provider provides.
type WorkerProviderer interface {
	RegisterWorker(c echo.Context) error
	DeregisterWorker(c echo.Context) error
	GetWorkerRegisterSecret(c echo.Context) error
	GetWorkerStatusOverview(c echo.Context) error
	ResetWorkerRegisterSecret(c echo.Context) error
	GetWorker(c echo.Context) error
}

// NewWorkerProvider creates a provider which provides worker related functionality.
func NewWorkerProvider(deps Dependencies) *WorkerProvider {
	return &WorkerProvider{
		deps: deps,
	}
}
