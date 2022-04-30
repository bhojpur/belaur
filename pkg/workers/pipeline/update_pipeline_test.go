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
	"path/filepath"
	"testing"
	"time"

	belaur "github.com/bhojpur/belaur"
	hclog "github.com/hashicorp/go-hclog"
)

func TestUpdatePipelinePython(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestUpdatePipelinePython")
	belaur.Cfg = new(belaur.Config)
	belaur.Cfg.HomePath = tmp
	buf := new(bytes.Buffer)
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Belaur",
	})

	p1 := belaur.Pipeline{
		Name:    "PipelinA",
		Type:    belaur.PTypePython,
		Created: time.Now(),
	}

	// Create fake virtualenv folder with temp file
	virtualEnvPath := filepath.Join(belaur.Cfg.HomePath, belaur.TmpFolder, belaur.TmpPythonFolder, p1.Name)
	err := os.MkdirAll(virtualEnvPath, 0700)
	if err != nil {
		t.Fatal(err)
	}
	src := filepath.Join(tmp, "PipelineA_python")
	p1.ExecPath = src
	defer os.RemoveAll(tmp)
	_ = ioutil.WriteFile(src, []byte("testcontent"), 0666)

	// fake execution commands
	virtualEnvName = "mkdir"
	pythonPipInstallCmd = "echo %s"

	// run
	err = updatePipeline(&p1)
	if err != nil {
		t.Fatal(err)
	}

	// check if file has been copied to correct place
	if _, err = os.Stat(filepath.Join(virtualEnvPath, p1.Name+".tar.gz")); err != nil {
		t.Fatalf("distribution file does not exist: %s", err.Error())
	}
}

func TestUpdatePipelineRuby(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestUpdatePipelineRuby")
	belaur.Cfg = new(belaur.Config)
	belaur.Cfg.HomePath = tmp
	buf := new(bytes.Buffer)
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Belaur",
	})

	p1 := belaur.Pipeline{
		Name:    "PipelinA",
		Type:    belaur.PTypeRuby,
		Created: time.Now(),
	}

	// Create fake test gem file.
	src := filepath.Join(tmp, "PipelineA_ruby")
	p1.ExecPath = src
	defer os.RemoveAll(tmp)
	_ = ioutil.WriteFile(src, []byte("testcontent"), 0666)

	// fake execution commands
	rubyGemName = "echo"

	// run
	err := updatePipeline(&p1)
	if err != nil {
		t.Fatal(err)
	}
}

func TestUpdatePipelineNodeJS(t *testing.T) {
	tmp, err := ioutil.TempDir("", "TestUpdatePipelineNodeJS")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmp)
	belaur.Cfg = new(belaur.Config)
	belaur.Cfg.HomePath = tmp
	buf := new(bytes.Buffer)
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: buf,
		Name:   "Belaur",
	})

	p1 := belaur.Pipeline{
		Name:    "PipelinA",
		Type:    belaur.PTypeNodeJS,
		Created: time.Now(),
	}

	// Create fake test nodejs archive file.
	src := filepath.Join(tmp, "PipelineA_nodejs")
	p1.ExecPath = src
	if err := ioutil.WriteFile(src, []byte("testcontent"), 0666); err != nil {
		t.Fatal(err)
	}

	// fake execution commands
	tarName = "echo"
	npmName = "echo"

	// run
	err = updatePipeline(&p1)
	if err != nil {
		t.Fatal(err)
	}
}
