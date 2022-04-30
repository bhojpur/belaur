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
	"bytes"
	"encoding/json"
	"errors"
	"io/ioutil"
	"net/http"
	"net/http/httptest"
	"os"
	"path/filepath"
	"strings"
	"testing"
	"time"

	"github.com/hashicorp/go-hclog"
	"github.com/labstack/echo/v4"

	belaur "github.com/bhojpur/belaur"
	"github.com/bhojpur/belaur/pkg/services"
	gStore "github.com/bhojpur/belaur/pkg/store"
	"github.com/bhojpur/belaur/pkg/workers/pipeline"
	"github.com/bhojpur/belaur/pkg/workers/scheduler/service"
)

type mockScheduleService struct {
	service.BelaurScheduler
	pipelineRun *belaur.PipelineRun
	err         error
}

func (ms *mockScheduleService) SchedulePipeline(p *belaur.Pipeline, startReason string, args []*belaur.Argument) (*belaur.PipelineRun, error) {
	return ms.pipelineRun, ms.err
}

func TestPipelineGitLSRemote(t *testing.T) {
	dataDir, _ := ioutil.TempDir("", "TestPipelineGitLSRemote")

	defer func() {
		belaur.Cfg = nil
	}()

	belaur.Cfg = &belaur.Config{
		Logger:   hclog.NewNullLogger(),
		DataPath: dataDir,
	}

	pipelineService := pipeline.NewBelaurPipelineService(pipeline.Dependencies{
		Scheduler: &mockScheduleService{},
	})

	pp := NewPipelineProvider(Dependencies{
		Scheduler:       &mockScheduleService{},
		PipelineService: pipelineService,
	})

	e := echo.New()

	t.Run("fails with invalid data", func(t *testing.T) {
		req := httptest.NewRequest(echo.POST, "/api/"+belaur.APIVersion+"/pipeline/gitlsremote", bytes.NewBuffer([]byte(`{"invalid"}`)))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = pp.PipelineGitLSRemote(c)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected response code %v got %v", http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("fails with invalid access", func(t *testing.T) {
		repoURL := "https://example.com"
		body := map[string]string{
			"url":      repoURL,
			"user":     "admin",
			"password": "admin",
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(echo.POST, "/api/"+belaur.APIVersion+"/pipeline/gitlsremote", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = pp.PipelineGitLSRemote(c)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected response code %v got %v", http.StatusOK, rec.Code)
		}
	})

	t.Run("otherwise succeed", func(t *testing.T) {
		repoURL := "https://github.com/bhojpur/pipeline-test"
		body := map[string]string{
			"url": repoURL,
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(echo.POST, "/api/"+belaur.APIVersion+"/pipeline/gitlsremote", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = pp.PipelineGitLSRemote(c)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected response code %v got %v", http.StatusOK, rec.Code)
		}
	})
}

func TestPipelineGitLSRemoteWithKeys(t *testing.T) {
	samplePrivateKey := `
-----BEGIN RSA PRIVATE KEY-----
MD8CAQACCQDB9DczYvFuZQIDAQABAgkAtqAKvH9QoQECBQDjAl9BAgUA2rkqJQIE
Xbs5AQIEIzWnmQIFAOEml+E=
-----END RSA PRIVATE KEY-----
`
	dataDir, _ := ioutil.TempDir("", "TestPipelineGitLSRemoteWithKeys")

	defer func() {
		belaur.Cfg = nil
	}()

	belaur.Cfg = &belaur.Config{
		Logger:   hclog.NewNullLogger(),
		DataPath: dataDir,
	}

	pipelineService := pipeline.NewBelaurPipelineService(pipeline.Dependencies{
		Scheduler: &mockScheduleService{},
	})

	pp := NewPipelineProvider(Dependencies{
		Scheduler:       &mockScheduleService{},
		PipelineService: pipelineService,
	})

	e := echo.New()

	t.Run("invalid hostconfig for github in known_hosts file", func(t *testing.T) {
		buf := new(bytes.Buffer)
		belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
			Level:  hclog.Trace,
			Output: buf,
			Name:   "Belaur",
		})
		hostConfig := "github.comom,1.2.3.4 ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAq2A7hRGmdnm9tUDbO9IDSwBK6TbQa+PXYPCPy6rbTrTtw7PHkccKrpp0yVhp5HdEIcKr6pLlVDBfOLX9QUsyCOV0wzfjIJNlGEYsdlLJizHhbn2mUjvSAHQqZETYP81eFzLQNnPHt4EVVUh7VfDESU84KezmD5QlWpXLmvU31/yMf+Se8xhHTvKSCZIFImWwoG6mbUoWf9nzpIoaSjB+weqqUUmpaaasXVal72J+UX2B+2RPW3RcT0eOzQgqlJL3RKrTJvdsjE3JEAvGq3lGHSZXy28G3skua2SmVi/w4yCE6gbODqnTWlg7+wC604ydGXA8VJiS5ap43JXiUFFAaQ=="
		knownHostsLocation := filepath.Join(dataDir, ".known_hosts")
		err := ioutil.WriteFile(knownHostsLocation, []byte(hostConfig), belaur.ExecutablePermission)
		if err != nil {
			t.Fatal(err)
		}
		_ = os.Setenv("SSH_KNOWN_HOSTS", knownHostsLocation)
		repoURL := "github.com:bhojpur/pipeline-test"
		gr := belaur.GitRepo{
			URL: repoURL,
			PrivateKey: belaur.PrivateKey{
				Key:      samplePrivateKey,
				Username: "git",
				Password: "",
			},
		}
		bodyBytes, _ := json.Marshal(gr)
		req := httptest.NewRequest(echo.POST, "/api/"+belaur.APIVersion+"/pipeline/gitlsremote", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = pp.PipelineGitLSRemote(c)

		// This will fail because the above SSH key is invalid. But that is fine,
		// because the initial host file will fail earlier than that.
		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected response code %v got %v", http.StatusForbidden, rec.Code)
		}

		// This is the important bit that needs to be tested.
		want := "knownhosts: key is unknown"
		if !strings.Contains(buf.String(), want) {
			t.Fatalf("wanted buf to contain: '%s', got: '%s'", want, buf.String())
		}
	})
}

func TestPipelineUpdate(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestPipelineUpdate")
	dataDir := tmp

	belaur.Cfg = &belaur.Config{
		Logger:       hclog.NewNullLogger(),
		DataPath:     dataDir,
		HomePath:     dataDir,
		PipelinePath: dataDir,
	}
	belaur.Cfg.Bolt.Mode = 0600

	pipelineService := pipeline.NewBelaurPipelineService(pipeline.Dependencies{
		Scheduler: &mockScheduleService{},
	})

	pp := NewPipelineProvider(Dependencies{
		Scheduler:       &mockScheduleService{},
		PipelineService: pipelineService,
	})

	// Initialize store
	dataStore, err := services.StorageService()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { services.MockStorageService(nil) }()
	// Initialize global active pipelines
	ap := pipeline.NewActivePipelines()
	pipeline.GlobalActivePipelines = ap

	// Initialize echo
	e := echo.New()

	pipeline1 := belaur.Pipeline{
		ID:                1,
		Name:              "Pipeline A",
		Type:              belaur.PTypeGolang,
		Created:           time.Now(),
		PeriodicSchedules: []string{"0 30 * * * *"},
	}

	pipeline2 := belaur.Pipeline{
		ID:      2,
		Name:    "Pipeline B",
		Type:    belaur.PTypeGolang,
		Created: time.Now(),
	}

	// Add to store
	err = dataStore.PipelinePut(&pipeline1)
	if err != nil {
		t.Fatal(err)
	}
	// Add to active pipelines
	ap.Append(pipeline1)
	// Create binary
	src := pipeline.GetExecPath(pipeline1)
	f, _ := os.Create(src)
	defer f.Close()
	defer os.Remove(src)

	t.Run("fails for non-existent pipeline", func(t *testing.T) {
		bodyBytes, _ := json.Marshal(pipeline2)
		req := httptest.NewRequest(echo.PUT, "/", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/" + belaur.APIVersion + "/pipeline/:pipelineid")
		c.SetParamNames("pipelineid")
		c.SetParamValues("2")

		_ = pp.PipelineUpdate(c)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected response code %v got %v", http.StatusNotFound, rec.Code)
		}
	})

	t.Run("works for existing pipeline", func(t *testing.T) {
		bodyBytes, _ := json.Marshal(pipeline1)
		req := httptest.NewRequest(echo.PUT, "/", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/" + belaur.APIVersion + "/pipeline/:pipelineid")
		c.SetParamNames("pipelineid")
		c.SetParamValues("1")

		_ = pp.PipelineUpdate(c)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected response code %v got %v", http.StatusOK, rec.Code)
		}
	})

	t.Run("update periodic schedules success", func(t *testing.T) {
		p := belaur.Pipeline{
			ID:                1,
			Name:              "newname",
			PeriodicSchedules: []string{"0 */1 * * * *"},
		}
		bodyBytes, _ := json.Marshal(p)
		req := httptest.NewRequest(echo.PUT, "/", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/" + belaur.APIVersion + "/pipeline/:pipelineid")
		c.SetParamNames("pipelineid")
		c.SetParamValues("1")

		ms := new(mockScheduleService)
		pRun := new(belaur.PipelineRun)
		pRun.ID = 999
		ms.pipelineRun = pRun

		_ = pp.PipelineUpdate(c)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected response code %v got %v", http.StatusOK, rec.Code)
		}
	})

	t.Run("update periodic schedules failed", func(t *testing.T) {
		p := belaur.Pipeline{
			ID:                1,
			Name:              "newname",
			PeriodicSchedules: []string{"0 */1 * * * * *"},
		}
		bodyBytes, _ := json.Marshal(p)
		req := httptest.NewRequest(echo.PUT, "/", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/" + belaur.APIVersion + "/pipeline/:pipelineid")
		c.SetParamNames("pipelineid")
		c.SetParamValues("1")

		ms := new(mockScheduleService)
		pRun := new(belaur.PipelineRun)
		pRun.ID = 999
		ms.pipelineRun = pRun

		_ = pp.PipelineUpdate(c)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected response code %v got %v", http.StatusBadRequest, rec.Code)
		}
	})
}

func TestPipelineDelete(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestPipelineDelete")
	dataDir := tmp

	belaur.Cfg = &belaur.Config{
		Logger:       hclog.NewNullLogger(),
		HomePath:     dataDir,
		DataPath:     dataDir,
		PipelinePath: dataDir,
	}

	// Initialize store
	dataStore, err := services.StorageService()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { services.MockStorageService(nil) }()

	// Initialize global active pipelines
	ap := pipeline.NewActivePipelines()
	pipeline.GlobalActivePipelines = ap

	pipelineService := pipeline.NewBelaurPipelineService(pipeline.Dependencies{
		Scheduler: &mockScheduleService{},
	})

	pp := NewPipelineProvider(Dependencies{
		Scheduler:       &mockScheduleService{},
		PipelineService: pipelineService,
	})

	// Initialize echo
	e := echo.New()

	p := belaur.Pipeline{
		ID:      1,
		Name:    "Pipeline A",
		Type:    belaur.PTypeGolang,
		Created: time.Now(),
	}

	// Add to store
	err = dataStore.PipelinePut(&p)
	if err != nil {
		t.Fatal(err)
	}
	// Add to active pipelines
	ap.Append(p)
	// Create binary
	src := pipeline.GetExecPath(p)
	f, _ := os.Create(src)
	defer f.Close()
	defer os.Remove(src)

	_ = ioutil.WriteFile(src, []byte("testcontent"), 0666)

	t.Run("fails for non-existent pipeline", func(t *testing.T) {
		req := httptest.NewRequest(echo.DELETE, "/", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/" + belaur.APIVersion + "/pipeline/:pipelineid")
		c.SetParamNames("pipelineid")
		c.SetParamValues("2")

		_ = pp.PipelineDelete(c)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected response code %v got %v", http.StatusNotFound, rec.Code)
		}
	})

	t.Run("works for existing pipeline", func(t *testing.T) {
		req := httptest.NewRequest(echo.DELETE, "/", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/" + belaur.APIVersion + "/pipeline/:pipelineid")
		c.SetParamNames("pipelineid")
		c.SetParamValues("1")

		_ = pp.PipelineDelete(c)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected response code %v got %v", http.StatusNotFound, rec.Code)
		}
	})

}

func TestPipelineStart(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestPipelineStart")
	belaur.Cfg = &belaur.Config{
		Logger:       hclog.NewNullLogger(),
		HomePath:     tmp,
		DataPath:     tmp,
		PipelinePath: tmp,
	}

	// Initialize global active pipelines
	ap := pipeline.NewActivePipelines()
	pipeline.GlobalActivePipelines = ap

	pipelineService := pipeline.NewBelaurPipelineService(pipeline.Dependencies{
		Scheduler: &mockScheduleService{},
	})

	pp := NewPipelineProvider(Dependencies{
		Scheduler:       &mockScheduleService{},
		PipelineService: pipelineService,
	})

	// Initialize echo
	e := echo.New()

	p := belaur.Pipeline{
		ID:      1,
		Name:    "Pipeline A",
		Type:    belaur.PTypeGolang,
		Created: time.Now(),
	}

	// Add to active pipelines
	ap.Append(p)

	t.Run("can start a pipeline", func(t *testing.T) {
		bodyBytes, _ := json.Marshal(map[string]interface{}{
			"docker": false,
		})
		req := httptest.NewRequest(echo.POST, "/api/"+belaur.APIVersion+"/pipeline/:pipelineid/start", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("pipelineid")
		c.SetParamValues("1")

		pRun := new(belaur.PipelineRun)
		pRun.ID = 999
		pp := NewPipelineProvider(Dependencies{
			Scheduler:       &mockScheduleService{pipelineRun: pRun},
			PipelineService: pipelineService,
		})
		_ = pp.PipelineStart(c)

		if rec.Code != http.StatusCreated {
			t.Fatalf("expected response code %v got %v", http.StatusCreated, rec.Code)
		}

		expectedBody := `{"uniqueid":"","id":999,"pipelineid":0,"startdate":"0001-01-01T00:00:00Z","started_reason":"","finishdate":"0001-01-01T00:00:00Z","scheduledate":"0001-01-01T00:00:00Z"}
`
		body, _ := ioutil.ReadAll(rec.Body)
		if string(body) != expectedBody {
			t.Fatalf("body did not equal expected content. expected: %s, got: %s", expectedBody, string(body))
		}
	})

	t.Run("fails when scheduler throws error", func(t *testing.T) {
		bodyBytes, _ := json.Marshal(map[string]interface{}{
			"docker": false,
		})
		req := httptest.NewRequest(echo.POST, "/api/"+belaur.APIVersion+"/pipeline/:pipelineid/start", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("pipelineid")
		c.SetParamValues("1")

		pRun := new(belaur.PipelineRun)
		pRun.ID = 999
		pp := NewPipelineProvider(Dependencies{
			Scheduler:       &mockScheduleService{pipelineRun: pRun, err: errors.New("failed to run pipeline")},
			PipelineService: pipelineService,
		})
		_ = pp.PipelineStart(c)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected response code %v got %v", http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("fails when scheduler doesn't find the pipeline but does not return error", func(t *testing.T) {
		bodyBytes, _ := json.Marshal(map[string]interface{}{
			"docker": false,
		})
		req := httptest.NewRequest(echo.POST, "/api/"+belaur.APIVersion+"/pipeline/:pipelineid/start", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("pipelineid")
		c.SetParamValues("1")

		_ = pp.PipelineStart(c)

		if rec.Code != http.StatusNotFound {
			t.Fatalf("expected response code %v got %v", http.StatusNotFound, rec.Code)
		}
	})
}

type mockPipelineService struct {
	pipeline.Servicer
	err error
}

func (m *mockPipelineService) UpdateRepository(p *belaur.Pipeline) error {
	return m.err
}

func TestPipelinePull(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestPipelinePull")
	belaur.Cfg = &belaur.Config{
		Logger:       hclog.NewNullLogger(),
		HomePath:     tmp,
		DataPath:     tmp,
		PipelinePath: tmp,
	}

	// Initialize global active pipelines
	ap := pipeline.NewActivePipelines()
	pipeline.GlobalActivePipelines = ap

	// Initialize echo
	e := echo.New()

	p := belaur.Pipeline{
		ID:      1,
		Name:    "Pipeline A",
		Type:    belaur.PTypeGolang,
		Created: time.Now(),
		Repo: &belaur.GitRepo{
			URL: "https://github.com/Skarlso/go-example",
		},
	}

	// Add to active pipelines
	ap.Append(p)

	t.Run("can pull a pipeline", func(t *testing.T) {
		bodyBytes, _ := json.Marshal(map[string]interface{}{
			"docker": false,
		})
		req := httptest.NewRequest(echo.POST, "/api/"+belaur.APIVersion+"/pipeline/:pipelineid/pull", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("pipelineid")
		c.SetParamValues("1")
		pp := NewPipelineProvider(Dependencies{
			Scheduler:       &mockScheduleService{},
			PipelineService: &mockPipelineService{},
		})
		err := pp.PipelinePull(c)
		if err != nil {
			t.Fatal(err)
		}

		if rec.Code != http.StatusOK {
			t.Fatalf("expected response code %v got %v", http.StatusOK, rec.Code)
		}
	})
}

type MockVaultStorer struct {
	Error error
}

var storeBts []byte

func (mvs *MockVaultStorer) Init() error {
	storeBts = make([]byte, 0)
	return mvs.Error
}

func (mvs *MockVaultStorer) Read() ([]byte, error) {
	return storeBts, mvs.Error
}

func (mvs *MockVaultStorer) Write(data []byte) error {
	storeBts = data
	return mvs.Error
}

func TestHookReceive(t *testing.T) {
	dataDir, err := ioutil.TempDir("", "TestHookReceive")
	if err != nil {
		t.Fatalf("error creating data dir %v", err.Error())
	}
	pipelineService := pipeline.NewBelaurPipelineService(pipeline.Dependencies{
		Scheduler: &mockScheduleService{},
	})

	defer func() {
		belaur.Cfg = nil
		_ = os.RemoveAll(dataDir)
	}()
	belaur.Cfg = &belaur.Config{
		Logger:    hclog.NewNullLogger(),
		DataPath:  dataDir,
		CAPath:    dataDir,
		VaultPath: dataDir,
		HomePath:  dataDir,
	}
	pp := NewPipelineProvider(Dependencies{
		Scheduler:       &mockScheduleService{},
		PipelineService: pipelineService,
	})
	m := new(MockVaultStorer)
	v, _ := services.VaultService(m)
	v.Add("GITHUB_WEBHOOK_SECRET_1", []byte("superawesomesecretgithubpassword"))
	defer func() {
		services.MockVaultService(nil)
	}()
	e := echo.New()

	// Initialize global active pipelines
	ap := pipeline.NewActivePipelines()
	pipeline.GlobalActivePipelines = ap

	p := belaur.Pipeline{
		ID:      1,
		Name:    "Pipeline A",
		Type:    belaur.PTypeGolang,
		Created: time.Now(),
		Repo: &belaur.GitRepo{
			URL: "https://github.com/Codertocat/Hello-World",
		},
	}

	ap.Append(p)

	t.Run("successfully extracting path information from payload", func(t *testing.T) {
		payload, _ := ioutil.ReadFile(filepath.Join("fixtures", "hook_basic_push_payload.json"))
		req := httptest.NewRequest(echo.POST, "/api/"+belaur.APIVersion+"/pipeline/githook", bytes.NewBuffer(payload))
		req.Header.Set("Content-Type", "application/json")
		// Use https://www.freeformatter.com/hmac-generator.html#ad-output for example
		// to calculate a new sha if the fixture would change.
		req.Header.Set("x-hub-signature", "sha1=940e53f44518a6cf9ba002c29c8ace7799af2b13")
		req.Header.Set("x-github-event", "push")
		req.Header.Set("X-github-delivery", "1234asdf")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = pp.GitWebHook(c)

		// Expected failure because repository does not exist
		if rec.Code != http.StatusInternalServerError {
			t.Fatalf("want response code %v got %v", http.StatusInternalServerError, rec.Code)
		}

		// Checking body to make sure it's the failure we want
		body, _ := ioutil.ReadAll(rec.Body)
		want := "failed to build pipeline:  repository does not exist\n"
		if string(body) != want {
			t.Fatalf("want body: %s, got: %s", want, string(body))
		}
	})

	t.Run("only push events are accepted", func(t *testing.T) {
		payload, _ := ioutil.ReadFile(filepath.Join("fixtures", "hook_basic_push_payload.json"))
		req := httptest.NewRequest(echo.POST, "/api/"+belaur.APIVersion+"/pipeline/githook", bytes.NewBuffer(payload))
		req.Header.Set("Content-Type", "application/json")
		// Use https://www.freeformatter.com/hmac-generator.html#ad-output for example
		// to calculate a new sha if the fixture would change.
		req.Header.Set("x-hub-signature", "sha1=940e53f44518a6cf9ba002c29c8ace7799af2b13")
		req.Header.Set("x-github-event", "pull")
		req.Header.Set("X-github-delivery", "1234asdf")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = pp.GitWebHook(c)

		// Expected failure because repository does not exist
		if rec.Code != http.StatusBadRequest {
			t.Fatalf("want response code %v got %v", http.StatusBadRequest, rec.Code)
		}

		// Checking body to make sure it's the failure we want
		body, _ := ioutil.ReadAll(rec.Body)
		want := "invalid event"
		if string(body) != want {
			t.Fatalf("want body: %s, got: %s", want, string(body))
		}
	})
}

type mockUserStoreService struct {
	gStore.BelaurStore
	user *belaur.User
	err  error
}

func (m mockUserStoreService) UserGet(username string) (*belaur.User, error) {
	return m.user, m.err
}

func TestPipelineRemoteTrigger(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestPipelineRemoteTrigger")
	belaur.Cfg = &belaur.Config{
		Logger:       hclog.NewNullLogger(),
		HomePath:     tmp,
		DataPath:     tmp,
		PipelinePath: tmp,
	}

	// Initialize global active pipelines
	ap := pipeline.NewActivePipelines()
	pipeline.GlobalActivePipelines = ap

	pipelineService := pipeline.NewBelaurPipelineService(pipeline.Dependencies{
		Scheduler: &mockScheduleService{},
	})

	// Initialize echo
	e := echo.New()

	p := belaur.Pipeline{
		ID:           1,
		Name:         "Pipeline A",
		Type:         belaur.PTypeGolang,
		Created:      time.Now(),
		TriggerToken: "triggerToken",
	}

	// Add to active pipelines
	ap.Append(p)

	t.Run("can trigger a pipeline with auto user", func(t *testing.T) {
		user := belaur.User{}
		user.Username = "auto"
		user.TriggerToken = "triggerToken"
		m := mockUserStoreService{user: &user, err: nil}
		services.MockStorageService(&m)
		defer func() {
			services.MockStorageService(nil)
		}()

		req := httptest.NewRequest(echo.POST, "/api/"+belaur.APIVersion+"/pipeline/1/triggerToken/trigger", nil)
		req.Header.Set("Content-Type", "application/json")
		req.SetBasicAuth("auto", "triggerToken")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("pipelineid", "pipelinetoken")
		c.SetParamValues("1", "triggerToken")
		pRun := new(belaur.PipelineRun)
		pRun.ID = 999
		pp := NewPipelineProvider(Dependencies{
			Scheduler:       &mockScheduleService{pipelineRun: pRun},
			PipelineService: pipelineService,
		})

		_ = pp.PipelineTrigger(c)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected response code %v got %v", http.StatusOK, rec.Code)
		}
	})
	t.Run("can't trigger a pipeline with invalid auto user", func(t *testing.T) {
		user := belaur.User{}
		user.Username = "auto"
		user.TriggerToken = "triggerToken"
		m := mockUserStoreService{user: &user, err: nil}
		services.MockStorageService(&m)
		defer func() {
			services.MockStorageService(nil)
		}()

		req := httptest.NewRequest(echo.POST, "/api/"+belaur.APIVersion+"/pipeline/1/triggerToken/trigger", nil)
		req.Header.Set("Content-Type", "application/json")
		req.SetBasicAuth("auto", "invalid")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("pipelineid", "pipelinetoken")
		c.SetParamValues("1", "triggerToken")
		pRun := new(belaur.PipelineRun)
		pRun.ID = 999
		pp := NewPipelineProvider(Dependencies{
			Scheduler:       &mockScheduleService{pipelineRun: pRun},
			PipelineService: pipelineService,
		})

		_ = pp.PipelineTrigger(c)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected response code %v got %v", http.StatusBadRequest, rec.Code)
		}
	})
	t.Run("can't trigger a pipeline with invalid token", func(t *testing.T) {
		user := belaur.User{}
		user.Username = "auto"
		user.TriggerToken = "triggerToken"
		m := mockUserStoreService{user: &user, err: nil}
		services.MockStorageService(&m)
		defer func() {
			services.MockStorageService(nil)
		}()

		req := httptest.NewRequest(echo.POST, "/api/"+belaur.APIVersion+"/pipeline/1/invalid/trigger", nil)
		req.Header.Set("Content-Type", "application/json")
		req.SetBasicAuth("auto", "triggerToken")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("pipelineid", "pipelinetoken")
		c.SetParamValues("1", "invalid")
		pRun := new(belaur.PipelineRun)
		pRun.ID = 999
		pp := NewPipelineProvider(Dependencies{
			Scheduler:       &mockScheduleService{pipelineRun: pRun},
			PipelineService: pipelineService,
		})

		_ = pp.PipelineTrigger(c)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected response code %v got %v", http.StatusForbidden, rec.Code)
		}
	})
	t.Run("can't trigger a pipeline without authentication information", func(t *testing.T) {
		user := belaur.User{}
		user.Username = "auto"
		user.TriggerToken = "triggerToken"
		m := mockUserStoreService{user: &user, err: nil}
		services.MockStorageService(&m)
		defer func() {
			services.MockStorageService(nil)
		}()

		req := httptest.NewRequest(echo.POST, "/api/"+belaur.APIVersion+"/pipeline/1/invalid/trigger", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetParamNames("pipelineid", "pipelinetoken")
		c.SetParamValues("1", "invalid")
		pRun := new(belaur.PipelineRun)
		pRun.ID = 999
		pp := NewPipelineProvider(Dependencies{
			Scheduler:       &mockScheduleService{pipelineRun: pRun},
			PipelineService: pipelineService,
		})

		_ = pp.PipelineTrigger(c)

		if rec.Code != http.StatusForbidden {
			t.Fatalf("expected response code %v got %v", http.StatusForbidden, rec.Code)
		}
	})
}

type mockPipelineResetStorageService struct {
	gStore.BelaurStore
}

var pipelineNewToken string

func (m mockPipelineResetStorageService) PipelinePut(pipeline *belaur.Pipeline) error {
	pipelineNewToken = pipeline.TriggerToken
	return nil
}

func TestPipelineResetToken(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestPipelineResetToken")
	belaur.Cfg = &belaur.Config{
		Logger:       hclog.NewNullLogger(),
		HomePath:     tmp,
		DataPath:     tmp,
		PipelinePath: tmp,
	}

	// Initialize global active pipelines
	ap := pipeline.NewActivePipelines()
	pipeline.GlobalActivePipelines = ap

	pipelineService := pipeline.NewBelaurPipelineService(pipeline.Dependencies{
		Scheduler: &mockScheduleService{},
	})

	// Initialize echo
	e := echo.New()

	p := belaur.Pipeline{
		ID:           1,
		Name:         "Pipeline A",
		Type:         belaur.PTypeGolang,
		Created:      time.Now(),
		TriggerToken: "triggerToken",
	}

	// Add to active pipelines
	ap.Append(p)

	req := httptest.NewRequest(echo.PUT, "/api/"+belaur.APIVersion+"/pipeline/1/reset-trigger-token", nil)
	rec := httptest.NewRecorder()
	c := e.NewContext(req, rec)
	c.SetParamNames("pipelineid")
	c.SetParamValues("1")
	pRun := new(belaur.PipelineRun)
	pRun.ID = 999
	pp := NewPipelineProvider(Dependencies{
		Scheduler:       &mockScheduleService{pipelineRun: pRun},
		PipelineService: pipelineService,
	})

	m := mockPipelineResetStorageService{}
	services.MockStorageService(&m)

	defer func() {
		services.MockStorageService(nil)
	}()

	_ = pp.PipelineResetToken(c)

	if rec.Code != http.StatusOK {
		t.Fatalf("expected response code %v got %v", http.StatusOK, rec.Code)
	}

	if pipelineNewToken == "triggerToken" {
		t.Fatal("expected token to be reset. was not reset.")
	}
}

func TestPipelineCheckPeriodicSchedules(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestPipelineCheckPeriodicSchedules")
	belaur.Cfg = &belaur.Config{
		Logger:       hclog.NewNullLogger(),
		HomePath:     tmp,
		DataPath:     tmp,
		PipelinePath: tmp,
	}

	// Initialize global active pipelines
	ap := pipeline.NewActivePipelines()
	pipeline.GlobalActivePipelines = ap

	pipelineService := pipeline.NewBelaurPipelineService(pipeline.Dependencies{
		Scheduler: &mockScheduleService{},
	})

	pp := NewPipelineProvider(Dependencies{
		Scheduler:       &mockScheduleService{},
		PipelineService: pipelineService,
	})

	// Initialize echo
	e := echo.New()

	t.Run("invalid cron added", func(t *testing.T) {
		body := []string{
			"* * * * * * *",
			"*/1 * * 200 *",
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(echo.POST, "/api/"+belaur.APIVersion+"/pipeline/periodicschedules", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = pp.PipelineCheckPeriodicSchedules(c)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected response code %v got %v", http.StatusBadRequest, rec.Code)
		}
	})

	t.Run("valid cron added", func(t *testing.T) {
		body := []string{
			"0 30 * * * *",
			"0 */5 * * * *",
		}
		bodyBytes, _ := json.Marshal(body)
		req := httptest.NewRequest(echo.POST, "/api/"+belaur.APIVersion+"/pipeline/periodicschedules", bytes.NewBuffer(bodyBytes))
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)

		_ = pp.PipelineCheckPeriodicSchedules(c)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected response code %v got %v", http.StatusOK, rec.Code)
		}
	})
}

func TestPipelineNameValidation(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestPipelineNameValidation")
	dataDir := tmp

	belaur.Cfg = &belaur.Config{
		Logger:       hclog.NewNullLogger(),
		HomePath:     dataDir,
		DataPath:     dataDir,
		PipelinePath: dataDir,
	}

	// Initialize store
	dataStore, err := services.StorageService()
	if err != nil {
		t.Fatal(err)
	}
	defer func() { services.MockStorageService(nil) }()

	// Initialize global active pipelines
	ap := pipeline.NewActivePipelines()
	pipeline.GlobalActivePipelines = ap

	pipelineService := pipeline.NewBelaurPipelineService(pipeline.Dependencies{
		Scheduler: &mockScheduleService{},
	})

	pp := NewPipelineProvider(Dependencies{
		Scheduler:       &mockScheduleService{},
		PipelineService: pipelineService,
	})

	// Initialize echo
	e := echo.New()

	p := belaur.Pipeline{
		ID:      1,
		Name:    "Pipeline A",
		Type:    belaur.PTypeGolang,
		Created: time.Now(),
	}

	// Add to store
	err = dataStore.PipelinePut(&p)
	if err != nil {
		t.Fatal(err)
	}
	// Add to active pipelines
	ap.Append(p)

	t.Run("fails for pipeline name already in use", func(t *testing.T) {
		req := httptest.NewRequest(echo.GET, "/", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/" + belaur.APIVersion + "/pipeline/name")
		q := req.URL.Query()
		q.Add("name", "pipeline a")
		req.URL.RawQuery = q.Encode()

		_ = pp.PipelineNameAvailable(c)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected response code %v got %v", http.StatusBadRequest, rec.Code)
		}
		bodyBytes, err := ioutil.ReadAll(rec.Body)
		if err != nil {
			t.Fatalf("cannot read response body: %s", err.Error())
		}
		nameAlreadyInUseMessage := "pipeline name is already in use"
		if string(bodyBytes[:]) != nameAlreadyInUseMessage {
			t.Fatalf("error message should be '%s' but was '%s'", nameAlreadyInUseMessage, string(bodyBytes[:]))
		}
	})

	t.Run("fails for pipeline name is too long", func(t *testing.T) {
		req := httptest.NewRequest(echo.GET, "/", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/" + belaur.APIVersion + "/pipeline/name")
		q := req.URL.Query()
		q.Add("name", "pipeline a pipeline a pipeline a pipeline a pipeline a pipeline a pipeline a pipeline a pipeline a pipeline a")
		req.URL.RawQuery = q.Encode()

		_ = pp.PipelineNameAvailable(c)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected response code %v got %v", http.StatusBadRequest, rec.Code)
		}
		bodyBytes, err := ioutil.ReadAll(rec.Body)
		if err != nil {
			t.Fatalf("cannot read response body: %s", err.Error())
		}
		nameTooLongMessage := "name of pipeline is empty or one of the path elements length exceeds 50 characters"
		if string(bodyBytes[:]) != nameTooLongMessage {
			t.Fatalf("error message should be '%s' but was '%s'", nameTooLongMessage, string(bodyBytes[:]))
		}
	})

	t.Run("fails for pipeline name with invalid character", func(t *testing.T) {
		req := httptest.NewRequest(echo.GET, "/", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/" + belaur.APIVersion + "/pipeline/name")
		q := req.URL.Query()
		q.Add("name", "this[]isinvalid;")
		req.URL.RawQuery = q.Encode()

		_ = pp.PipelineNameAvailable(c)

		if rec.Code != http.StatusBadRequest {
			t.Fatalf("expected response code %v got %v", http.StatusBadRequest, rec.Code)
		}
		bodyBytes, err := ioutil.ReadAll(rec.Body)
		if err != nil {
			t.Fatalf("cannot read response body: %s", err.Error())
		}
		nameContainsInvalidCharacter := "must match [A-z][0-9][-][_][ ]"
		if string(bodyBytes[:]) != nameContainsInvalidCharacter {
			t.Fatalf("error message should be '%s' but was '%s'", nameContainsInvalidCharacter, string(bodyBytes[:]))
		}
	})

	t.Run("works for pipeline with different name", func(t *testing.T) {
		req := httptest.NewRequest(echo.GET, "/", nil)
		req.Header.Set("Content-Type", "application/json")
		rec := httptest.NewRecorder()
		c := e.NewContext(req, rec)
		c.SetPath("/api/" + belaur.APIVersion + "/pipeline/name")
		q := req.URL.Query()
		q.Add("name", "pipeline b")
		req.URL.RawQuery = q.Encode()

		_ = pp.PipelineNameAvailable(c)

		if rec.Code != http.StatusOK {
			t.Fatalf("expected response code %v got %v", http.StatusOK, rec.Code)
		}
	})
}
