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
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"regexp"
	"strings"
	"testing"

	belaur "github.com/bhojpur/belaur"
	"github.com/bhojpur/belaur/pkg/helper/pipelinehelper"

	"github.com/bhojpur/belaur/pkg/services"
	"github.com/bhojpur/belaur/pkg/store"
	hclog "github.com/hashicorp/go-hclog"
)

type mockStorer struct {
	store.BelaurStore
	Error error
}

// PipelinePut is a Mock implementation for pipelines
func (m *mockStorer) PipelinePut(p *belaur.Pipeline) error {
	return m.Error
}

func TestPrepareEnvironmentGo(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestPrepareEnvironmentGo")
	belaur.Cfg = new(belaur.Config)
	belaur.Cfg.HomePath = tmp
	b := new(BuildPipelineGolang)
	p := new(belaur.CreatePipeline)
	err := b.PrepareEnvironment(p)
	if err != nil {
		t.Fatal("error was not expected when preparing environment: ", err)
	}
	var expectedDest = regexp.MustCompile(`^/.*/tmp/golang/src/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	if !expectedDest.MatchString(p.Pipeline.Repo.LocalDest) {
		t.Fatalf("expected destination is '%s', but was '%s'", expectedDest, p.Pipeline.Repo.LocalDest)
	}
}

func TestPrepareEnvironmentInvalidPathForMkdirGo(t *testing.T) {
	belaur.Cfg = new(belaur.Config)
	belaur.Cfg.HomePath = "/notexists"
	b := new(BuildPipelineGolang)
	p := new(belaur.CreatePipeline)
	err := b.PrepareEnvironment(p)
	if err == nil {
		t.Fatal("error was expected but none occurred")
	}
}

func TestExecuteBuildGo(t *testing.T) {
	execCommandContext = fakeExecCommandContext
	defer func() {
		execCommandContext = exec.CommandContext
	}()
	tmp, _ := ioutil.TempDir("", "TestExecuteBuildGo")
	belaur.Cfg = new(belaur.Config)
	belaur.Cfg.HomePath = tmp
	b := new(BuildPipelineGolang)
	p := new(belaur.CreatePipeline)
	err := b.ExecuteBuild(p)
	if err != nil {
		t.Fatal("error while running executebuild. none was expected")
	}
	expectedDepArgs := "get,-d,./..."
	expectedBuildArgs := "build,-o,_"
	actualArgs := os.Getenv("CMD_ARGS")
	if !strings.Contains(actualArgs, expectedBuildArgs) && !strings.Contains(actualArgs, expectedDepArgs) {
		t.Fatalf("expected args '%s, %s' actual args '%s'", expectedDepArgs, expectedBuildArgs, actualArgs)
	}
}

func TestExecuteBuildContextTimeoutGo(t *testing.T) {
	execCommandContext = fakeExecCommandContext
	buildKillContext = true
	defer func() {
		execCommandContext = exec.CommandContext
		buildKillContext = false
	}()
	tmp, _ := ioutil.TempDir("", "TestExecuteBuildContextTimeoutGo")
	belaur.Cfg = new(belaur.Config)
	belaur.Cfg.HomePath = tmp
	// Initialize shared logger
	buf := new(bytes.Buffer)
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Belaur",
	})
	b := new(BuildPipelineGolang)
	p := new(belaur.CreatePipeline)
	err := b.ExecuteBuild(p)
	if err == nil {
		t.Fatal("no error found while expecting error.")
	}
	if err.Error() != "context deadline exceeded" {
		t.Fatal("context deadline should have been exceeded. was instead: ", err)
	}
}

func TestExecuteBuildBinaryNotFoundErrorGo(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestExecuteBuildBinaryNotFoundErrorGo")
	belaur.Cfg = new(belaur.Config)
	belaur.Cfg.HomePath = tmp
	// Initialize shared logger
	buf := new(bytes.Buffer)
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Belaur",
	})
	currentPath := os.Getenv("PATH")
	defer func() { _ = os.Setenv("PATH", currentPath) }()
	_ = os.Setenv("PATH", "")
	b := new(BuildPipelineGolang)
	p := new(belaur.CreatePipeline)
	err := b.ExecuteBuild(p)
	if err == nil {
		t.Fatal("no error found while expecting error.")
	}
	if err.Error() != "exec: \"go\": executable file not found in $PATH" {
		t.Fatal("the error wasn't what we expected. instead it was: ", err)
	}
}

func TestCopyBinaryGo(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestCopyBinaryGo")
	belaur.Cfg = new(belaur.Config)
	belaur.Cfg.HomePath = tmp
	// Initialize shared logger
	buf := new(bytes.Buffer)
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Belaur",
	})
	b := new(BuildPipelineGolang)
	p := new(belaur.CreatePipeline)
	p.Pipeline.Name = "main"
	p.Pipeline.Type = "go"
	p.Pipeline.Repo = &belaur.GitRepo{LocalDest: tmp}
	src := filepath.Join(tmp, pipelinehelper.AppendTypeToName(p.Pipeline.Name, p.Pipeline.Type))
	dst := pipelinehelper.AppendTypeToName(p.Pipeline.Name, p.Pipeline.Type)
	f, _ := os.Create(src)
	defer f.Close()
	defer os.Remove(dst)
	_ = ioutil.WriteFile(src, []byte("testcontent"), 0666)
	err := b.CopyBinary(p)
	if err != nil {
		t.Fatal("error was not expected when copying binary: ", err)
	}
	content, err := ioutil.ReadFile(dst)
	if err != nil {
		t.Fatal("error encountered while reading destination file: ", err)
	}
	if string(content) != "testcontent" {
		t.Fatal("file content did not equal src content. was: ", string(content))
	}
}

func TestCopyBinarySrcDoesNotExistGo(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestCopyBinarySrcDoesNotExistGo")
	belaur.Cfg = new(belaur.Config)
	belaur.Cfg.HomePath = tmp
	// Initialize shared logger
	buf := new(bytes.Buffer)
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Belaur",
	})
	b := new(BuildPipelineGolang)
	p := new(belaur.CreatePipeline)
	p.Pipeline.Name = "main"
	p.Pipeline.Type = "go"
	p.Pipeline.Repo = &belaur.GitRepo{LocalDest: "/noneexistent"}
	err := b.CopyBinary(p)
	if err == nil {
		t.Fatal("error was expected when copying binary but none occurred ")
	}
	if err.Error() != "open /noneexistent/main_go: no such file or directory" {
		t.Fatal("a different error occurred then expected: ", err)
	}
}

func TestSavePipelineGo(t *testing.T) {
	defer os.Remove("belaur.db")
	belaur.Cfg = new(belaur.Config)
	belaur.Cfg.HomePath = "/tmp"
	belaur.Cfg.PipelinePath = "/tmp/pipelines/"
	// Initialize shared logger
	p := new(belaur.Pipeline)
	p.Name = "main"
	p.Type = belaur.PTypeGolang
	b := new(BuildPipelineGolang)
	m := new(mockStorer)
	services.MockStorageService(m)
	err := b.SavePipeline(p)
	if err != nil {
		t.Fatal("something went wrong. wasn't supposed to get error: ", err)
	}
	if p.Name != "main" {
		t.Fatal("name of pipeline didn't equal expected 'main'. was instead: ", p.Name)
	}
	if p.Type != belaur.PTypeGolang {
		t.Fatal("type of pipeline was not go. instead was: ", p.Type)
	}
}

func TestSavePipelineSaveErrors(t *testing.T) {
	defer os.Remove("belaur.db")
	belaur.Cfg = new(belaur.Config)
	belaur.Cfg.HomePath = "/tmp"
	belaur.Cfg.PipelinePath = "/tmp/pipelines/"
	// Initialize shared logger
	p := new(belaur.Pipeline)
	p.Name = "main"
	p.Type = belaur.PTypeGolang
	b := new(BuildPipelineGolang)
	m := new(mockStorer)
	m.Error = errors.New("database error")
	services.MockStorageService(m)
	err := b.SavePipeline(p)
	if err == nil {
		t.Fatal("expected error which did not occur")
	}
	if err.Error() != "database error" {
		t.Fatal("error message was not the expected message. was: ", err.Error())
	}
}
