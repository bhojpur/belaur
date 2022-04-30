package rbac

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
	"testing"

	belaur "github.com/bhojpur/belaur"
	plcsvr "github.com/bhojpur/policy/pkg/engine"
	"github.com/hashicorp/go-hclog"
	"github.com/stretchr/testify/assert"
)

type mockEnforcer struct {
	plcsvr.IEnforcer
}

var mappings = APILookup{
	"/api/v1/pipeline/:pipelineid": {
		Methods: map[string]string{
			"GET": "pipelines/get",
		},
		Param: "pipelineid",
	},
}

func (m *mockEnforcer) Enforce(rvals ...interface{}) (bool, error) {
	role := rvals[0].(string)
	if role == "admin" {
		return true, nil
	}
	if role == "failed" {
		return false, nil
	}
	return false, errors.New("error test")
}

func Test_EnforcerService_Enforce_ValidEnforcement(t *testing.T) {
	belaur.Cfg = &belaur.Config{
		Logger: hclog.NewNullLogger(),
	}
	defer func() {
		belaur.Cfg = nil
	}()

	svc := enforcerService{
		enforcer:      &mockEnforcer{},
		rbacAPILookup: mappings,
	}

	err := svc.Enforce("admin", "GET", "/api/v1/pipelines/:pipelineid", map[string]string{"pipelineid": "test"})
	assert.NoError(t, err)
}

func Test_EnforcerService_Enforce_FailedEnforcement(t *testing.T) {
	belaur.Cfg = &belaur.Config{
		Logger: hclog.NewNullLogger(),
	}
	defer func() {
		belaur.Cfg = nil
	}()

	svc := enforcerService{
		enforcer:      &mockEnforcer{},
		rbacAPILookup: mappings,
	}

	err := svc.Enforce("failed", "GET", "/api/v1/pipeline/:pipelineid", map[string]string{"pipelineid": "test"})
	assert.EqualError(t, err, "Permission denied. Must have pipelines/get test")
}

func Test_EnforcerService_Enforce_ErrorEnforcement(t *testing.T) {
	belaur.Cfg = &belaur.Config{
		Logger: hclog.NewNullLogger(),
	}
	defer func() {
		belaur.Cfg = nil
	}()

	svc := enforcerService{
		enforcer:      &mockEnforcer{},
		rbacAPILookup: mappings,
	}

	err := svc.Enforce("error", "GET", "/api/v1/pipeline/:pipelineid", map[string]string{"pipelineid": "test"})
	assert.EqualError(t, err, "error enforcing rbac: error test")
}

func Test_EnforcerService_Enforce_EndpointParamMissing(t *testing.T) {
	belaur.Cfg = &belaur.Config{
		Logger: hclog.NewNullLogger(),
	}
	defer func() {
		belaur.Cfg = nil
	}()

	svc := enforcerService{
		enforcer:      &mockEnforcer{},
		rbacAPILookup: mappings,
	}

	err := svc.Enforce("readonly", "GET", "/api/v1/pipeline/:pipelineid", map[string]string{})
	assert.EqualError(t, err, "error param pipelineid missing")
}
