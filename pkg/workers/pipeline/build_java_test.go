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

func TestPrepareEnvironmentJava(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestPrepareEnvironmentJava")
	belaur.Cfg = new(belaur.Config)
	belaur.Cfg.HomePath = tmp
	b := new(BuildPipelineJava)
	p := new(belaur.CreatePipeline)
	err := b.PrepareEnvironment(p)
	if err != nil {
		t.Fatal("error was not expected when preparing environment: ", err)
	}
	var expectedDest = regexp.MustCompile(`^/.*/tmp/java/src/[0-9a-f]{8}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{4}-[0-9a-f]{12}$`)
	if !expectedDest.MatchString(p.Pipeline.Repo.LocalDest) {
		t.Fatalf("expected destination is '%s', but was '%s'", expectedDest, p.Pipeline.Repo.LocalDest)
	}
}

func TestPrepareEnvironmentInvalidPathForMkdirJava(t *testing.T) {
	belaur.Cfg = new(belaur.Config)
	belaur.Cfg.HomePath = "/notexists"
	b := new(BuildPipelineJava)
	p := new(belaur.CreatePipeline)
	err := b.PrepareEnvironment(p)
	if err == nil {
		t.Fatal("error was expected but none occurred")
	}
}

func TestExecuteBuildJava(t *testing.T) {
	execCommandContext = fakeExecCommandContext
	defer func() {
		execCommandContext = exec.CommandContext
	}()
	tmp, _ := ioutil.TempDir("", "TestExecuteBuildJava")
	belaur.Cfg = new(belaur.Config)
	belaur.Cfg.HomePath = tmp
	b := new(BuildPipelineJava)
	p := new(belaur.CreatePipeline)
	// go must be existent, mvn maybe not.
	mavenBinaryName = "go"
	err := b.ExecuteBuild(p)
	if err != nil {
		t.Fatal("error while running executebuild. none was expected")
	}
	expectedBuildArgs := "clean,compile,assembly:single"
	actualArgs := os.Getenv("CMD_ARGS")
	if !strings.Contains(actualArgs, expectedBuildArgs) {
		t.Fatalf("expected args '%s' actual args '%s'", expectedBuildArgs, actualArgs)
	}
}

func TestExecuteBuildContextTimeoutJava(t *testing.T) {
	execCommandContext = fakeExecCommandContext
	buildKillContext = true
	defer func() {
		execCommandContext = exec.CommandContext
		buildKillContext = false
	}()
	tmp, _ := ioutil.TempDir("", "TestExecuteBuildContextTimeoutJava")
	belaur.Cfg = new(belaur.Config)
	belaur.Cfg.HomePath = tmp
	// Initialize shared logger
	buf := new(bytes.Buffer)
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Belaur",
	})
	b := new(BuildPipelineJava)
	p := new(belaur.CreatePipeline)
	// go must be existent, mvn maybe not.
	mavenBinaryName = "go"
	err := b.ExecuteBuild(p)
	if err == nil {
		t.Fatal("no error found while expecting error.")
	}
	if err.Error() != "context deadline exceeded" {
		t.Fatal("context deadline should have been exceeded. was instead: ", err)
	}
}

func TestCopyBinaryJava(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestCopyBinaryJava")
	belaur.Cfg = new(belaur.Config)
	belaur.Cfg.HomePath = tmp
	// Initialize shared logger
	buf := new(bytes.Buffer)
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Belaur",
	})
	b := new(BuildPipelineJava)
	p := new(belaur.CreatePipeline)
	p.Pipeline.Name = "main"
	p.Pipeline.Type = belaur.PTypeJava
	p.Pipeline.Repo = &belaur.GitRepo{LocalDest: tmp}
	_ = os.Mkdir(filepath.Join(tmp, mavenTargetFolder), 0744)
	src := filepath.Join(tmp, mavenTargetFolder, javaFinalJarName)
	dst := pipelinehelper.AppendTypeToName(p.Pipeline.Name, p.Pipeline.Type)
	f, _ := os.Create(src)
	defer os.Remove(filepath.Join(tmp, mavenTargetFolder))
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

func TestCopyBinarySrcDoesNotExistJava(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestCopyBinarySrcDoesNotExistJava")
	belaur.Cfg = new(belaur.Config)
	belaur.Cfg.HomePath = tmp
	// Initialize shared logger
	buf := new(bytes.Buffer)
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Belaur",
	})
	b := new(BuildPipelineJava)
	p := new(belaur.CreatePipeline)
	p.Pipeline.Name = "main"
	p.Pipeline.Type = belaur.PTypeJava
	p.Pipeline.Repo = &belaur.GitRepo{LocalDest: "/noneexistent"}
	err := b.CopyBinary(p)
	if err == nil {
		t.Fatal("error was expected when copying binary but none occurred ")
	}
	if err.Error() != "open /noneexistent/target/plugin-jar-with-dependencies.jar: no such file or directory" {
		t.Fatal("a different error occurred then expected: ", err)
	}
}

type javaMockStorer struct {
	store.BelaurStore
	Error error
}

// PipelinePut is a Mock implementation for pipelines
func (m *javaMockStorer) PipelinePut(p *belaur.Pipeline) error {
	return m.Error
}

func TestSavePipelineJava(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestSavePipelineJava")
	belaur.Cfg = new(belaur.Config)
	belaur.Cfg.HomePath = tmp
	m := new(javaMockStorer)
	services.MockStorageService(m)
	defer os.Remove(tmp)
	defer os.Remove("belaur.db")
	belaur.Cfg.PipelinePath = tmp + "/pipelines/"
	defer os.Remove(belaur.Cfg.PipelinePath)
	p := new(belaur.Pipeline)
	p.Name = "main"
	p.Type = belaur.PTypeJava
	b := new(BuildPipelineJava)
	err := b.SavePipeline(p)
	if err != nil {
		t.Fatal("something went wrong. wasn't supposed to get error: ", err)
	}
	if p.Name != "main" {
		t.Fatal("name of pipeline didn't equal expected 'main'. was instead: ", p.Name)
	}
	if p.Type != belaur.PTypeJava {
		t.Fatal("type of pipeline was not java. instead was: ", p.Type)
	}
}
