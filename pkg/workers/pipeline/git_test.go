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
	"bytes"
	"context"
	"errors"
	"io/ioutil"
	"net/http"
	"os"
	"path/filepath"
	"strings"
	"testing"

	belaur "github.com/bhojpur/belaur"
	"github.com/bhojpur/belaur/pkg/services"
	"github.com/google/go-github/github"
	"github.com/hashicorp/go-hclog"
)

func TestGitCloneRepo(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestGitCloneRepo")
	repo := &belaur.GitRepo{
		URL:       "https://github.com/bhojpur/pipeline-test",
		LocalDest: tmp,
	}
	err := gitCloneRepo(repo)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUpdateAllPipelinesRepositoryNotFound(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestUpdateAllPipelinesRepositoryNotFound")
	belaur.Cfg = new(belaur.Config)
	belaur.Cfg.HomePath = tmp
	// Initialize shared logger
	b := new(bytes.Buffer)
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: b,
		Name:   "Belaur",
	})

	p := new(belaur.Pipeline)
	p.Repo = &belaur.GitRepo{LocalDest: tmp}
	GlobalActivePipelines = NewActivePipelines()
	GlobalActivePipelines.Append(*p)
	pipelineService := NewBelaurPipelineService(Dependencies{
		Scheduler: &mockScheduleService{},
	})
	pipelineService.UpdateAllCurrentPipelines()
	if !strings.Contains(b.String(), "repository does not exist") {
		t.Fatal("error message not found in logs: ", b.String())
	}
}

func TestUpdateAllPipelinesAlreadyUpToDate(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestUpdateAllPipelinesAlreadyUpToDate")
	belaur.Cfg = new(belaur.Config)
	belaur.Cfg.HomePath = tmp
	// Initialize shared logger
	b := new(bytes.Buffer)
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: b,
		Name:   "Belaur",
	})
	repo := &belaur.GitRepo{
		URL:            "https://github.com/bhojpur/pipeline-test",
		LocalDest:      tmp,
		SelectedBranch: "refs/heads/master",
	}
	// always ensure that tmp folder is cleaned up
	defer os.RemoveAll(tmp)
	err := gitCloneRepo(repo)
	if err != nil {
		t.Fatal(err)
	}

	p := new(belaur.Pipeline)
	p.Name = "main"
	p.Repo = &belaur.GitRepo{}
	p.Repo.SelectedBranch = "refs/heads/master"
	p.Repo.LocalDest = tmp
	GlobalActivePipelines = NewActivePipelines()
	GlobalActivePipelines.Append(*p)
	pipelineService := NewBelaurPipelineService(Dependencies{
		Scheduler: &mockScheduleService{},
	})
	pipelineService.UpdateAllCurrentPipelines()
	if !strings.Contains(b.String(), "already up-to-date") {
		t.Fatal("log output did not contain error message that the repo is up-to-date.: ", b.String())
	}
}

func TestCloneRepoWithSSHAuth(t *testing.T) {
	samplePrivateKey := `
-----BEGIN RSA PRIVATE KEY-----
MD8CAQACCQDB9DczYvFuZQIDAQABAgkAtqAKvH9QoQECBQDjAl9BAgUA2rkqJQIE
Xbs5AQIEIzWnmQIFAOEml+E=
-----END RSA PRIVATE KEY-----
`
	tmp, _ := ioutil.TempDir("", "TestCloneRepoWithSSHAuth")
	belaur.Cfg = new(belaur.Config)
	belaur.Cfg.HomePath = tmp
	// Initialize shared logger
	b := new(bytes.Buffer)
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: b,
		Name:   "Belaur",
	})
	repo := &belaur.GitRepo{
		URL:            "github.com:bhojpur/pipeline-test",
		LocalDest:      tmp,
		SelectedBranch: "refs/heads/master",
		PrivateKey: belaur.PrivateKey{
			Key:      samplePrivateKey,
			Username: "git",
			Password: "",
		},
	}
	hostConfig := "notgithub.comom,192.30.252.130 ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAq2A7hRGmdnm9tUDbO9IDSwBK6TbQa+PXYPCPy6rbTrTtw7PHkccKrpp0yVhp5HdEIcKr6pLlVDBfOLX9QUsyCOV0wzfjIJNlGEYsdlLJizHhbn2mUjvSAHQqZETYP81eFzLQNnPHt4EVVUh7VfDESU84KezmD5QlWpXLmvU31/yMf+Se8xhHTvKSCZIFImWwoG6mbUoWf9nzpIoaSjB+weqqUUmpaaasXVal72J+UX2B+2RPW3RcT0eOzQgqlJL3RKrTJvdsjE3JEAvGq3lGHSZXy28G3skua2SmVi/w4yCE6gbODqnTWlg7+wC604ydGXA8VJiS5ap43JXiUFFAaQ=="
	knownHostsLocation := filepath.Join(tmp, ".known_hosts")
	err := ioutil.WriteFile(knownHostsLocation, []byte(hostConfig), belaur.ExecutablePermission)
	if err != nil {
		t.Fatal(err)
	}

	_ = os.Setenv("SSH_KNOWN_HOSTS", knownHostsLocation)

	// always ensure that tmp folder is cleaned up
	defer os.RemoveAll(tmp)
	_ = gitCloneRepo(repo)
	want := "knownhosts: key is unknown"
	if !strings.Contains(b.String(), want) {
		t.Fatalf("wanted buf to contain: '%s', got: '%s'", want, b.String())
	}
}

func TestUpdateRepoWithSSHAuth(t *testing.T) {
	samplePrivateKey := `
-----BEGIN RSA PRIVATE KEY-----
MD8CAQACCQDB9DczYvFuZQIDAQABAgkAtqAKvH9QoQECBQDjAl9BAgUA2rkqJQIE
Xbs5AQIEIzWnmQIFAOEml+E=
-----END RSA PRIVATE KEY-----
`
	tmp, _ := ioutil.TempDir("", "TestUpdateRepoWithSSHAuth")
	belaur.Cfg = new(belaur.Config)
	belaur.Cfg.HomePath = tmp
	// Initialize shared logger
	b := new(bytes.Buffer)
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: b,
		Name:   "Belaur",
	})
	repo := &belaur.GitRepo{
		URL:            "github.com:bhojpur/pipeline-test",
		LocalDest:      tmp,
		SelectedBranch: "refs/heads/master",
		PrivateKey: belaur.PrivateKey{
			Key:      samplePrivateKey,
			Username: "git",
			Password: "",
		},
	}
	hostConfig := "github.com,140.82.121.4 ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAq2A7hRGmdnm9tUDbO9IDSwBK6TbQa+PXYPCPy6rbTrTtw7PHkccKrpp0yVhp5HdEIcKr6pLlVDBfOLX9QUsyCOV0wzfjIJNlGEYsdlLJizHhbn2mUjvSAHQqZETYP81eFzLQNnPHt4EVVUh7VfDESU84KezmD5QlWpXLmvU31/yMf+Se8xhHTvKSCZIFImWwoG6mbUoWf9nzpIoaSjB+weqqUUmpaaasXVal72J+UX2B+2RPW3RcT0eOzQgqlJL3RKrTJvdsjE3JEAvGq3lGHSZXy28G3skua2SmVi/w4yCE6gbODqnTWlg7+wC604ydGXA8VJiS5ap43JXiUFFAaQ=="
	knownHostsLocation := filepath.Join(tmp, ".known_hosts")
	err := ioutil.WriteFile(knownHostsLocation, []byte(hostConfig), belaur.ExecutablePermission)
	if err != nil {
		t.Fatal(err)
	}

	_ = os.Setenv("SSH_KNOWN_HOSTS", knownHostsLocation)

	// always ensure that tmp folder is cleaned up
	defer os.RemoveAll(tmp)
	_ = gitCloneRepo(repo)

	p := new(belaur.Pipeline)
	p.Name = "main"
	p.Repo = &belaur.GitRepo{}
	p.Repo.SelectedBranch = "refs/heads/master"
	p.Repo.LocalDest = tmp
	p.Type = belaur.PTypeGolang
	GlobalActivePipelines = NewActivePipelines()
	GlobalActivePipelines.Append(*p)
	hostConfig = "invalid.com,140.82.121.4 ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAq2A7hRGmdnm9tUDbO9IDSwBK6TbQa+PXYPCPy6rbTrTtw7PHkccKrpp0yVhp5HdEIcKr6pLlVDBfOLX9QUsyCOV0wzfjIJNlGEYsdlLJizHhbn2mUjvSAHQqZETYP81eFzLQNnPHt4EVVUh7VfDESU84KezmD5QlWpXLmvU31/yMf+Se8xhHTvKSCZIFImWwoG6mbUoWf9nzpIoaSjB+weqqUUmpaaasXVal72J+UX2B+2RPW3RcT0eOzQgqlJL3RKrTJvdsjE3JEAvGq3lGHSZXy28G3skua2SmVi/w4yCE6gbODqnTWlg7+wC604ydGXA8VJiS5ap43JXiUFFAaQ=="
	err = ioutil.WriteFile(knownHostsLocation, []byte(hostConfig), belaur.ExecutablePermission)
	if err != nil {
		t.Fatal(err)
	}
	pipelineService := NewBelaurPipelineService(Dependencies{
		Scheduler: &mockScheduleService{},
	})
	pipelineService.UpdateAllCurrentPipelines()
	want := "knownhosts: key is unknown"
	if !strings.Contains(b.String(), want) {
		t.Fatalf("wanted buf to contain: '%s', got: '%s'", want, b.String())
	}
}

func TestUpdateAllPipelinesAlreadyUpToDateWithMoreThanOnePipeline(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestUpdateAllPipelinesAlreadyUpToDateWithMoreThanOnePipeline")
	belaur.Cfg = new(belaur.Config)
	belaur.Cfg.HomePath = tmp
	// Initialize shared logger
	b := new(bytes.Buffer)
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: b,
		Name:   "Belaur",
	})
	repo := &belaur.GitRepo{
		URL:            "https://github.com/bhojpur/pipeline-test",
		LocalDest:      tmp,
		SelectedBranch: "refs/heads/master",
	}
	// always ensure that tmp folder is cleaned up
	defer os.RemoveAll(tmp)
	err := gitCloneRepo(repo)
	if err != nil {
		t.Fatal(err)
	}

	p1 := new(belaur.Pipeline)
	p1.Name = "main"
	p1.Repo = &belaur.GitRepo{}
	p1.Repo.SelectedBranch = "refs/heads/master"
	p1.Repo.LocalDest = tmp
	p2 := new(belaur.Pipeline)
	p2.Name = "main"
	p2.Repo = &belaur.GitRepo{}
	p2.Repo.SelectedBranch = "refs/heads/master"
	p2.Repo.LocalDest = tmp
	GlobalActivePipelines = NewActivePipelines()
	defer func() { GlobalActivePipelines = nil }()
	GlobalActivePipelines.Append(*p1)
	GlobalActivePipelines.Append(*p2)
	pipelineService := NewBelaurPipelineService(Dependencies{
		Scheduler: &mockScheduleService{},
	})
	pipelineService.UpdateAllCurrentPipelines()
	if !strings.Contains(b.String(), "already up-to-date") {
		t.Fatal("log output did not contain error message that the repo is up-to-date.: ", b.String())
	}
}

func TestGetAuthInfoWithUsernameAndPassword(t *testing.T) {
	repoWithUsernameAndPassword := &belaur.GitRepo{
		URL:       "https://github.com/bhojpur/pipeline-test",
		LocalDest: "tmp",
		Username:  "username",
		Password:  "password",
	}

	auth, _ := getAuthInfo(repoWithUsernameAndPassword, nil)
	if auth == nil {
		t.Fatal("auth should not be nil when username and password is provided")
	}
}

func TestGetAuthInfoWithPrivateKey(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestGetAuthInfoWithPrivateKey")
	hostConfig := "github.com,140.82.121.4 ssh-rsa AAAAB3NzaC1yc2EAAAABIwAAAQEAq2A7hRGmdnm9tUDbO9IDSwBK6TbQa+PXYPCPy6rbTrTtw7PHkccKrpp0yVhp5HdEIcKr6pLlVDBfOLX9QUsyCOV0wzfjIJNlGEYsdlLJizHhbn2mUjvSAHQqZETYP81eFzLQNnPHt4EVVUh7VfDESU84KezmD5QlWpXLmvU31/yMf+Se8xhHTvKSCZIFImWwoG6mbUoWf9nzpIoaSjB+weqqUUmpaaasXVal72J+UX2B+2RPW3RcT0eOzQgqlJL3RKrTJvdsjE3JEAvGq3lGHSZXy28G3skua2SmVi/w4yCE6gbODqnTWlg7+wC604ydGXA8VJiS5ap43JXiUFFAaQ=="
	knownHostsLocation := filepath.Join(tmp, ".known_hosts")
	err := ioutil.WriteFile(knownHostsLocation, []byte(hostConfig), belaur.ExecutablePermission)
	if err != nil {
		t.Fatal(err)
	}
	_ = os.Setenv("SSH_KNOWN_HOSTS", knownHostsLocation)
	samplePrivateKey := `
-----BEGIN RSA PRIVATE KEY-----
MD8CAQACCQDB9DczYvFuZQIDAQABAgkAtqAKvH9QoQECBQDjAl9BAgUA2rkqJQIE
Xbs5AQIEIzWnmQIFAOEml+E=
-----END RSA PRIVATE KEY-----
`
	repoWithValidPrivateKey := &belaur.GitRepo{
		URL:       "https://github.com/bhojpur/pipeline-test",
		LocalDest: "tmp",
		PrivateKey: belaur.PrivateKey{
			Key:      samplePrivateKey,
			Username: "username",
			Password: "password",
		},
		SelectedBranch: "refs/heads/master",
	}
	_, err = getAuthInfo(repoWithValidPrivateKey, nil)
	if err != nil {
		t.Fatal(err)
	}

	repoWithInvalidPrivateKey := &belaur.GitRepo{
		URL:       "https://github.com/bhojpur/pipeline-test",
		LocalDest: "tmp",
		PrivateKey: belaur.PrivateKey{
			Key:      "random_key",
			Username: "username",
			Password: "password",
		},
		SelectedBranch: "refs/heads/master",
	}
	auth, _ := getAuthInfo(repoWithInvalidPrivateKey, nil)
	if auth != nil {
		t.Fatal("auth should be nil for invalid private key")
	}
}

func TestGetAuthInfoEmpty(t *testing.T) {
	repoWithoutAuthInfo := &belaur.GitRepo{
		URL:       "https://github.com/bhojpur/pipeline-test",
		LocalDest: "tmp",
	}
	auth, _ := getAuthInfo(repoWithoutAuthInfo, nil)
	if auth != nil {
		t.Fatal("auth should be nil when no authentication info is provided")
	}
}

type MockGitVaultStorer struct {
	Error error
}

var gitStore []byte

func (mvs *MockGitVaultStorer) Init() error {
	gitStore = make([]byte, 0)
	return mvs.Error
}

func (mvs *MockGitVaultStorer) Read() ([]byte, error) {
	return gitStore, mvs.Error
}

func (mvs *MockGitVaultStorer) Write(data []byte) error {
	gitStore = data
	return mvs.Error
}

type MockGithubRepositoryService struct {
	Hook     *github.Hook
	Response *github.Response
	Error    error
	Owner    string
	Repo     string
}

func (mgc *MockGithubRepositoryService) CreateHook(ctx context.Context, owner, repo string, hook *github.Hook) (*github.Hook, *github.Response, error) {
	if owner != mgc.Owner {
		return nil, nil, errors.New("owner did not equal expected owner: was: " + owner)
	}
	if repo != mgc.Repo {
		return nil, nil, errors.New("repo did not equal expected repo: was: " + repo)
	}
	return mgc.Hook, mgc.Response, mgc.Error
}

func TestCreateGithubWebhook(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestCreateGithubWebhook")
	belaur.Cfg = &belaur.Config{}
	belaur.Cfg.VaultPath = tmp
	belaur.Cfg.CAPath = tmp
	buf := new(bytes.Buffer)
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Belaur",
	})

	m := new(MockGitVaultStorer)
	v, _ := services.VaultService(m)
	v.Add("GITHUB_WEBHOOK_SECRET_1", []byte("superawesomesecretgithubpassword"))
	defer func() {
		services.MockVaultService(nil)
	}()

	t.Run("successfully create webhook", func(t *testing.T) {
		repo := belaur.GitRepo{
			URL:            "https://github.com/bhojpur/belaur",
			SelectedBranch: "refs/heads/master",
		}

		mock := new(MockGithubRepositoryService)
		mock.Hook = &github.Hook{
			Name: github.String("test hook"),
			URL:  github.String("https://api.github.com/repos/bhojpur/belaur/hooks/44321286/test"),
		}
		mock.Response = &github.Response{
			Response: &http.Response{
				Status: "Ok",
			},
		}
		mock.Owner = "bhojpur"
		mock.Repo = "belaur"

		err := createGithubWebhook("asdf", &repo, "1", mock)
		if err != nil {
			t.Fatal("did not expect error to occur. was: ", err)
		}
		body, _ := ioutil.ReadAll(buf)
		expectedStatusMessage := []byte("hook created: : \"test hook\"=Ok")
		expectedHookURL := []byte("https://api.github.com/repos/bhojpur/belaur/hooks/44321286/test")
		if !bytes.Contains(body, expectedStatusMessage) {
			t.Fatalf("expected status message not found in logs. want:'%s', got: '%s'", expectedStatusMessage, body)
		}
		if !bytes.Contains(body, expectedHookURL) {
			t.Fatalf("expected hook url not found in logs. want:'%s', got: '%s'", expectedHookURL, body)
		}
	})

	t.Run("error while creating webhook", func(t *testing.T) {
		repo := belaur.GitRepo{
			URL:            "https://github.com/bhojpur/belaur",
			SelectedBranch: "refs/heads/master",
		}

		mock := new(MockGithubRepositoryService)
		mock.Error = errors.New("error from create webhook")
		mock.Hook = &github.Hook{
			Name: github.String("test hook"),
			URL:  github.String("https://api.github.com/repos/bhojpur/belaur/hooks/44321286/test"),
		}
		mock.Response = &github.Response{
			Response: &http.Response{
				Status: "Ok",
			},
		}
		mock.Owner = "bhojpur"
		mock.Repo = "belaur"

		err := createGithubWebhook("asdf", &repo, "1", mock)
		if err == nil {
			t.Fatal("CreateWebhook should have failed.")
		}
	})

	t.Run("successfully create webhook when password is not defined in advance", func(t *testing.T) {
		v.Remove("GITHUB_WEBHOOK_SECRET_1")
		_ = v.SaveSecrets()
		repo := belaur.GitRepo{
			URL:            "https://github.com/bhojpur/belaur",
			SelectedBranch: "refs/heads/master",
		}

		mock := new(MockGithubRepositoryService)
		mock.Hook = &github.Hook{
			Name: github.String("test hook"),
			URL:  github.String("https://api.github.com/repos/bhojpur/belaur/hooks/44321286/test"),
		}
		mock.Response = &github.Response{
			Response: &http.Response{
				Status: "Ok",
			},
		}
		mock.Owner = "bhojpur"
		mock.Repo = "belaur"

		err := createGithubWebhook("asdf", &repo, "1", mock)
		if err != nil {
			t.Fatal("did not expect error to occur. was: ", err)
		}
		body, _ := ioutil.ReadAll(buf)
		expectedStatusMessage := []byte("hook created: : \"test hook\"=Ok")
		expectedHookURL := []byte("https://api.github.com/repos/bhojpur/belaur/hooks/44321286/test")
		if !bytes.Contains(body, expectedStatusMessage) {
			t.Fatalf("expected status message not found in logs. want:'%s', got: '%s'", expectedStatusMessage, body)
		}
		if !bytes.Contains(body, expectedHookURL) {
			t.Fatalf("expected hook url not found in logs. want:'%s', got: '%s'", expectedHookURL, body)
		}
	})

	t.Run("if a secret is already defined it will not be overwritten", func(t *testing.T) {
		v.Add("GITHUB_WEBHOOK_SECRET_1", []byte("superawesomesecretgithubpassword"))
		_ = v.SaveSecrets()
		repo := belaur.GitRepo{
			URL:            "https://github.com/bhojpur/belaur",
			SelectedBranch: "refs/heads/master",
		}

		mock := new(MockGithubRepositoryService)
		mock.Hook = &github.Hook{
			Name: github.String("test hook"),
			URL:  github.String("https://api.github.com/repos/bhojpur/belaur/hooks/44321286/test"),
		}
		mock.Response = &github.Response{
			Response: &http.Response{
				Status: "Ok",
			},
		}
		mock.Owner = "bhojpur"
		mock.Repo = "belaur"

		err := createGithubWebhook("asdf", &repo, "1", mock)
		if err != nil {
			t.Fatal("did not expect error to occur. was: ", err)
		}
		body, _ := ioutil.ReadAll(buf)
		expectedStatusMessage := []byte("hook created: : \"test hook\"=Ok")
		expectedHookURL := []byte("https://api.github.com/repos/bhojpur/belaur/hooks/44321286/test")
		if !bytes.Contains(body, expectedStatusMessage) {
			t.Fatalf("expected status message not found in logs. want:'%s', got: '%s'", expectedStatusMessage, body)
		}
		if !bytes.Contains(body, expectedHookURL) {
			t.Fatalf("expected hook url not found in logs. want:'%s', got: '%s'", expectedHookURL, body)
		}
		secret, _ := v.Get("GITHUB_WEBHOOK_SECRET_1")
		if !bytes.Equal(secret, []byte("superawesomesecretgithubpassword")) {
			t.Fatalf("secret did not match. want: '%s' got: '%s'", "superawesomesecretgithubpassword", string(secret))
		}
	})
}

func TestMultipleGithubWebHookURLTypes(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestMultipleGithubWebHookURLTypes")
	belaur.Cfg = &belaur.Config{}
	belaur.Cfg.VaultPath = tmp
	belaur.Cfg.CAPath = tmp
	buf := new(bytes.Buffer)
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Belaur",
	})
	m := new(MockGitVaultStorer)
	v, _ := services.VaultService(m)
	v.Add("GITHUB_WEBHOOK_SECRET_1", []byte("superawesomesecretgithubpassword"))
	defer func() {
		services.MockVaultService(nil)
	}()

	t.Run("https url", func(t *testing.T) {
		repo := belaur.GitRepo{
			URL:            "https://github.com/bhojpur/belaur",
			SelectedBranch: "refs/heads/master",
		}

		mock := new(MockGithubRepositoryService)
		mock.Hook = &github.Hook{
			Name: github.String("test hook"),
			URL:  github.String("https://api.github.com/repos/bhojpur/belaur/hooks/44321286/test"),
		}
		mock.Response = &github.Response{
			Response: &http.Response{
				Status: "Ok",
			},
		}
		mock.Owner = "bhojpur"
		mock.Repo = "belaur"

		err := createGithubWebhook("asdf", &repo, "1", mock)
		if err != nil {
			t.Fatal("did not expect error to occur. was: ", err)
		}
	})

	t.Run("ssh url", func(t *testing.T) {
		repo := belaur.GitRepo{
			URL:            "git@github.com:bhojpur/belaur.git",
			SelectedBranch: "refs/heads/master",
		}

		mock := new(MockGithubRepositoryService)
		mock.Hook = &github.Hook{
			Name: github.String("test hook"),
			URL:  github.String("https://api.github.com/repos/bhojpur/belaur/hooks/44321286/test"),
		}
		mock.Response = &github.Response{
			Response: &http.Response{
				Status: "Ok",
			},
		}
		mock.Owner = "bhojpur"
		mock.Repo = "belaur"

		err := createGithubWebhook("asdf", &repo, "1", mock)
		if err != nil {
			t.Fatal("did not expect error to occur. was: ", err)
		}
	})

	t.Run("simple http with git extension", func(t *testing.T) {
		repo := belaur.GitRepo{
			URL:            "https://github.com/bhojpur/belaur.git",
			SelectedBranch: "refs/heads/master",
		}

		mock := new(MockGithubRepositoryService)
		mock.Hook = &github.Hook{
			Name: github.String("test hook"),
			URL:  github.String("https://api.github.com/repos/bhojpur/belaur/hooks/44321286/test"),
		}
		mock.Response = &github.Response{
			Response: &http.Response{
				Status: "Ok",
			},
		}
		mock.Owner = "bhojpur"
		mock.Repo = "belaur"

		err := createGithubWebhook("asdf", &repo, "1", mock)
		if err != nil {
			t.Fatal("did not expect error to occur. was: ", err)
		}
	})

	t.Run("failed extracting repo owner", func(t *testing.T) {
		repo := belaur.GitRepo{
			URL:            "https://invalid-giturl.com",
			SelectedBranch: "refs/heads/master",
		}

		mock := new(MockGithubRepositoryService)
		mock.Hook = &github.Hook{
			Name: github.String("test hook"),
			URL:  github.String("https://api.github.com/repos/bhojpur/belaur/hooks/44321286/test"),
		}
		mock.Response = &github.Response{
			Response: &http.Response{
				Status: "Ok",
			},
		}
		mock.Owner = "bhojpur"
		mock.Repo = "belaur"

		err := createGithubWebhook("asdf", &repo, "1", mock)
		if err == nil {
			t.Fatal("expected error. none found")
		}
	})
}

func TestGitLSRemote(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestCloneRepoWithSSHAuth")
	belaur.Cfg = new(belaur.Config)
	belaur.Cfg.HomePath = tmp
	// Initialize shared logger
	b := new(bytes.Buffer)
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: b,
		Name:   "Belaur",
	})
	repo := &belaur.GitRepo{
		URL:            "https://github.com/bhojpur/pipeline-test",
		LocalDest:      tmp,
		SelectedBranch: "refs/heads/master",
	}

	// always ensure that tmp folder is cleaned up
	defer os.RemoveAll(tmp)
	err := GitLSRemote(repo)
	if err != nil {
		t.Fatal(err)
	}
}
