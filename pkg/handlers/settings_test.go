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
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/hashicorp/go-hclog"
	"github.com/labstack/echo/v4"
	"github.com/stretchr/testify/assert"

	belaur "github.com/bhojpur/belaur"
	"github.com/bhojpur/belaur/pkg/providers/pipelines"
	"github.com/bhojpur/belaur/pkg/workers/pipeline"
	"github.com/bhojpur/belaur/pkg/workers/scheduler/service"
)

type status struct {
	Status bool
}

type mockScheduleService struct {
	service.BelaurScheduler
	pipelineRun *belaur.PipelineRun
	err         error
}

func (ms *mockScheduleService) SchedulePipeline(p *belaur.Pipeline, startReason string, args []*belaur.Argument) (*belaur.PipelineRun, error) {
	return ms.pipelineRun, ms.err
}

type mockSettingStoreService struct {
	get func() (*belaur.StoreConfig, error)
	put func(*belaur.StoreConfig) error
}

func (m mockSettingStoreService) SettingsGet() (*belaur.StoreConfig, error) {
	return m.get()
}

func (m mockSettingStoreService) SettingsPut(c *belaur.StoreConfig) error {
	return m.put(c)
}

func TestSetPollerToggle(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestSetPollerON")
	dataDir := tmp

	belaur.Cfg = &belaur.Config{
		Logger:       hclog.NewNullLogger(),
		DataPath:     dataDir,
		HomePath:     dataDir,
		PipelinePath: dataDir,
		Poll:         false,
	}

	pipelineService := pipeline.NewBelaurPipelineService(pipeline.Dependencies{
		Scheduler: &mockScheduleService{},
	})

	get := func() (*belaur.StoreConfig, error) {
		return nil, nil
	}
	put := func(*belaur.StoreConfig) error {
		return nil
	}
	m := mockSettingStoreService{get: get, put: put}

	pp := pipelines.NewPipelineProvider(pipelines.Dependencies{
		Scheduler:       &mockScheduleService{},
		PipelineService: pipelineService,
		SettingsStore:   m,
	})
	// // Initialize echo
	e := echo.New()

	t.Run("switching it on twice should fail", func(t2 *testing.T) {
		req := httptest.NewRequest(echo.POST, "/", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/" + belaur.APIVersion + "/setttings/poll/on")

		_ = pp.SettingsPollOn(c)
		retStatus := http.StatusOK
		if rec.Code != retStatus {
			t.Fatalf("expected response code %v got %v", retStatus, rec.Code)
		}

		req2 := httptest.NewRequest(echo.POST, "/", nil)
		req2.Header.Set("Content-Type", "application/json")
		rec2 := httptest.NewRecorder()
		c2 := e.NewContext(req2, rec2)
		c2.SetPath("/api/" + belaur.APIVersion + "/setttings/poll/on")

		_ = pp.SettingsPollOn(c2)
		secondRetStatus := http.StatusBadRequest
		if rec2.Code != secondRetStatus {
			t.Fatalf("expected response code %v got %v", secondRetStatus, rec2.Code)
		}
	})
	t.Run("switching it on while the setting is on should fail", func(t *testing.T) {
		belaur.Cfg = &belaur.Config{
			Logger:       hclog.NewNullLogger(),
			DataPath:     dataDir,
			HomePath:     dataDir,
			PipelinePath: dataDir,
			Poll:         true,
		}
		req := httptest.NewRequest(echo.POST, "/", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/" + belaur.APIVersion + "/setttings/poll/on")

		_ = pp.SettingsPollOn(c)
		retStatus := http.StatusBadRequest
		if rec.Code != retStatus {
			t.Fatalf("expected response code %v got %v", retStatus, rec.Code)
		}
	})
	t.Run("switching it off while the setting is on should pass", func(t *testing.T) {
		belaur.Cfg = &belaur.Config{
			Logger:       hclog.NewNullLogger(),
			DataPath:     dataDir,
			HomePath:     dataDir,
			PipelinePath: dataDir,
			Poll:         true,
		}
		req := httptest.NewRequest(echo.POST, "/", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/" + belaur.APIVersion + "/setttings/poll/off")

		_ = pp.SettingsPollOff(c)
		retStatus := http.StatusOK
		if rec.Code != retStatus {
			t.Fatalf("expected response code %v got %v", retStatus, rec.Code)
		}
		if belaur.Cfg.Poll != false {
			t.Fatalf("poll value should have been set to false. was: %v", belaur.Cfg.Poll)
		}
	})
	t.Run("switching it off while the setting is off should fail", func(t *testing.T) {
		belaur.Cfg = &belaur.Config{
			Logger:       hclog.NewNullLogger(),
			DataPath:     dataDir,
			HomePath:     dataDir,
			PipelinePath: dataDir,
			Poll:         false,
		}
		req := httptest.NewRequest(echo.POST, "/", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/" + belaur.APIVersion + "/setttings/poll/off")

		_ = pp.SettingsPollOff(c)
		retStatus := http.StatusBadRequest
		if rec.Code != retStatus {
			t.Fatalf("expected response code %v got %v", retStatus, rec.Code)
		}
	})
	t.Run("getting the value should return the correct setting", func(t *testing.T) {
		belaur.Cfg = &belaur.Config{
			Logger:       hclog.NewNullLogger(),
			DataPath:     dataDir,
			HomePath:     dataDir,
			PipelinePath: dataDir,
			Poll:         true,
		}
		req := httptest.NewRequest(echo.GET, "/", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/" + belaur.APIVersion + "/setttings/poll")

		_ = pp.SettingsPollGet(c)
		retStatus := http.StatusOK
		if rec.Code != retStatus {
			t.Fatalf("expected response code %v got %v", retStatus, rec.Code)
		}
		var s status
		_ = json.NewDecoder(rec.Body).Decode(&s)
		if s.Status != true {
			t.Fatalf("expected returned status to be true. was: %v", s.Status)
		}
	})
}

func TestGettingSettingFromDBTakesPrecedence(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestGettingSettingFromDBTakesPrecedence")
	dataDir := tmp

	belaur.Cfg = &belaur.Config{
		Logger:       hclog.NewNullLogger(),
		DataPath:     dataDir,
		HomePath:     dataDir,
		PipelinePath: dataDir,
		Poll:         false,
	}

	pipelineService := pipeline.NewBelaurPipelineService(pipeline.Dependencies{
		Scheduler: &mockScheduleService{},
	})

	get := func() (*belaur.StoreConfig, error) {
		return &belaur.StoreConfig{
			Poll: true,
		}, nil
	}
	put := func(*belaur.StoreConfig) error {
		return nil
	}
	m := mockSettingStoreService{get: get, put: put}

	pp := pipelines.NewPipelineProvider(pipelines.Dependencies{
		Scheduler:       &mockScheduleService{},
		PipelineService: pipelineService,
		SettingsStore:   m,
	})

	e := echo.New()

	req := httptest.NewRequest(echo.GET, "/", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/" + belaur.APIVersion + "/setttings/poll/")

	_ = pp.SettingsPollGet(c)
	retStatus := http.StatusOK
	if rec.Code != retStatus {
		t.Fatalf("expected response code %v got %v", retStatus, rec.Code)
	}
	var s status
	_ = json.NewDecoder(rec.Body).Decode(&s)
	if s.Status != true {
		t.Fatalf("expected returned status to be true from storage. was: %v", s.Status)
	}
}

func TestSettingPollerOnAlsoSavesSettingsInDB(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestSettingPollerOnAlsoSavesSettingsInDB")
	dataDir := tmp

	belaur.Cfg = &belaur.Config{
		Logger:       hclog.NewNullLogger(),
		DataPath:     dataDir,
		HomePath:     dataDir,
		PipelinePath: dataDir,
		Poll:         false,
	}

	pipelineService := pipeline.NewBelaurPipelineService(pipeline.Dependencies{
		Scheduler: &mockScheduleService{},
	})

	get := func() (*belaur.StoreConfig, error) {
		return &belaur.StoreConfig{
			Poll: true,
		}, nil
	}
	putCalled := false
	put := func(*belaur.StoreConfig) error {
		putCalled = true
		return nil
	}
	m := mockSettingStoreService{get: get, put: put}

	pp := pipelines.NewPipelineProvider(pipelines.Dependencies{
		Scheduler:       &mockScheduleService{},
		PipelineService: pipelineService,
		SettingsStore:   m,
	})

	e := echo.New()

	req := httptest.NewRequest(echo.POST, "/", nil)
	req.Header.Set("Content-Type", "application/json")
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetPath("/api/" + belaur.APIVersion + "/setttings/poll/on")

	_ = pp.SettingsPollOn(c)
	retStatus := http.StatusOK
	if rec.Code != retStatus {
		t.Fatalf("expected response code %v got %v", retStatus, rec.Code)
	}

	if putCalled != true {
		t.Fatal("SettingPut should have been called. Was not.")
	}
	putCalled = false
	_ = pp.SettingsPollOff(c)
	if putCalled != true {
		t.Fatal("SettingPut should have been called. Was not.")
	}
}

func Test_SettingsHandler_RBACGet(t *testing.T) {
	belaur.Cfg = &belaur.Config{
		Logger: hclog.NewNullLogger(),
	}

	e := echo.New()
	m := &mockSettingStoreService{}
	settingsHandler := newSettingsHandler(m)

	m.get = func() (*belaur.StoreConfig, error) {
		return &belaur.StoreConfig{}, nil
	}

	t.Run("error from store returns 500", func(t *testing.T) {
		req := httptest.NewRequest(echo.GET, "/", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/" + belaur.APIVersion + "/setttings/rbac")

		m.get = func() (*belaur.StoreConfig, error) {
			return &belaur.StoreConfig{}, errors.New("store error")
		}

		_ = settingsHandler.rbacGet(c)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Equal(t, "Something went wrong while retrieving settings information.", rec.Body.String())
	})

	t.Run("valid settings from store returns correct value", func(t *testing.T) {
		m.get = func() (*belaur.StoreConfig, error) {
			return &belaur.StoreConfig{
				RBACEnabled: true,
			}, nil
		}

		req := httptest.NewRequest(echo.GET, "/", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/" + belaur.APIVersion + "/setttings/rbac")

		_ = settingsHandler.rbacGet(c)

		assert.Equal(t, rec.Code, http.StatusOK)
		assert.Equal(t, rec.Body.String(), "{\"enabled\":true}\n")
	})
}

func Test_SettingsHandler_RBACPut(t *testing.T) {
	belaur.Cfg = &belaur.Config{
		Logger: hclog.NewNullLogger(),
	}

	e := echo.New()
	m := &mockSettingStoreService{}
	settingsHandler := newSettingsHandler(m)

	m.get = func() (*belaur.StoreConfig, error) {
		return &belaur.StoreConfig{}, nil
	}

	t.Run("store error returns 500", func(t *testing.T) {
		req := httptest.NewRequest(echo.GET, "/", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/" + belaur.APIVersion + "/setttings/rbac")

		m.put = func(config *belaur.StoreConfig) error {
			return errors.New("store error")
		}

		_ = settingsHandler.rbacPut(c)

		assert.Equal(t, http.StatusInternalServerError, rec.Code)
		assert.Equal(t, "An error occurred while saving the settings.", rec.Body.String())
	})

	t.Run("store success returns 200", func(t *testing.T) {
		m.get = func() (*belaur.StoreConfig, error) {
			return &belaur.StoreConfig{
				RBACEnabled: true,
			}, nil
		}

		req := httptest.NewRequest(echo.GET, "/", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/" + belaur.APIVersion + "/setttings/rbac")

		m.put = func(config *belaur.StoreConfig) error {
			return nil
		}

		_ = settingsHandler.rbacPut(c)

		assert.Equal(t, rec.Code, http.StatusOK)
		assert.Equal(t, rec.Body.String(), "Settings have been updated.")
	})
}
