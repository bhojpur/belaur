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
	"io/ioutil"
	"os"
	"path/filepath"
	"testing"
	"time"

	belaur "github.com/bhojpur/belaur"
	"github.com/bhojpur/belaur/pkg/helper/pipelinehelper"
)

func TestAppend(t *testing.T) {
	ap := NewActivePipelines()

	p1 := belaur.Pipeline{
		Name:    "Pipeline A",
		Type:    belaur.PTypeGolang,
		Created: time.Now(),
	}
	ap.Append(p1)

	ret := ap.GetByName("Pipeline A")

	if p1.Name != ret.Name || p1.Type != ret.Type {
		t.Fatalf("Appended pipeline is not present in active pipelines.")
	}

}

func TestUpdate(t *testing.T) {
	ap := NewActivePipelines()

	p1 := belaur.Pipeline{
		Name:    "Pipeline A",
		Type:    belaur.PTypeGolang,
		Created: time.Now(),
	}
	ap.Append(p1)

	p2 := belaur.Pipeline{
		Name:    "Pipeline B",
		Type:    belaur.PTypeGolang,
		Created: time.Now(),
	}

	err := ap.Update(0, p2)
	if err != nil {
		t.Fatal(err)
	}

	ret := ap.GetByName("Pipeline B")

	if p2.Name != ret.Name {
		t.Fatalf("Pipeline should have been updated.")
	}

}

func TestUpdateIndexOutOfBounds(t *testing.T) {
	ap := NewActivePipelines()

	p1 := belaur.Pipeline{
		Name:    "Pipeline A",
		Type:    belaur.PTypeGolang,
		Created: time.Now(),
	}
	ap.Append(p1)

	p2 := belaur.Pipeline{
		Name:    "Pipeline B",
		Type:    belaur.PTypeGolang,
		Created: time.Now(),
	}

	err := ap.Update(1, p2)
	if err == nil {
		t.Fatal("expected error to occur since we are out of bounds")
	}
}

func TestRemove(t *testing.T) {
	ap := NewActivePipelines()

	p1 := belaur.Pipeline{
		Name:    "Pipeline A",
		Type:    belaur.PTypeGolang,
		Created: time.Now(),
	}
	ap.Append(p1)

	p2 := belaur.Pipeline{
		Name:    "Pipeline B",
		Type:    belaur.PTypeGolang,
		Created: time.Now(),
	}
	ap.Append(p2)

	err := ap.Remove(1)
	if err != nil {
		t.Fatal(err)
	}

	count := 0
	for _, pipeline := range ap.GetAll() {
		count++
		if pipeline.Name == "Pipeline B" {
			t.Fatalf("Pipeline B still exists. It should have been removed.")
		}
	}

	if count != 1 {
		t.Fatalf("Expected pipeline count to be %v. Got %v.", 1, count)
	}
}

func TestRemoveInvalidIndex(t *testing.T) {
	ap := NewActivePipelines()

	p1 := belaur.Pipeline{
		Name:    "Pipeline A",
		Type:    belaur.PTypeGolang,
		Created: time.Now(),
	}
	ap.Append(p1)

	p2 := belaur.Pipeline{
		Name:    "Pipeline B",
		Type:    belaur.PTypeGolang,
		Created: time.Now(),
	}
	ap.Append(p2)

	err := ap.Remove(2)
	if err == nil {
		t.Fatal("expected error when accessing something outside the length ")
	}

	err = ap.Remove(3)
	if err == nil {
		t.Fatal("expected error when accessing something outside the length ")
	}
}

func TestGetByName(t *testing.T) {
	ap := NewActivePipelines()

	p1 := belaur.Pipeline{
		Name:    "Pipeline A",
		Type:    belaur.PTypeGolang,
		Created: time.Now(),
	}
	ap.Append(p1)

	ret := ap.GetByName("Pipeline A")

	if p1.Name != ret.Name || p1.Type != ret.Type {
		t.Fatalf("Pipeline A should have been retrieved.")
	}

	ret = ap.GetByName("Pipeline B")
	if ret != nil {
		t.Fatalf("Pipeline B should not have been retrieved.")
	}
}

func TestReplace(t *testing.T) {
	ap := NewActivePipelines()

	p1 := belaur.Pipeline{
		Name: "Pipeline A",
		Type: belaur.PTypeGolang,
		Repo: &belaur.GitRepo{
			URL:       "https://github.com/bhojpur/pipeline-test-1",
			LocalDest: "tmp",
		},
		Created: time.Now(),
	}
	ap.Append(p1)

	p2 := belaur.Pipeline{
		Name: "Pipeline A",
		Type: belaur.PTypeGolang,
		Repo: &belaur.GitRepo{
			URL:       "https://github.com/bhojpur/pipeline-test-2",
			LocalDest: "tmp",
		},
		Created: time.Now(),
	}
	ap.Append(p2)

	ret := ap.Replace(p2)
	if !ret {
		t.Fatalf("The pipeline could not be replaced")
	}

	p := ap.GetByName("Pipeline A")
	if p.Repo.URL != "https://github.com/bhojpur/pipeline-test-2" {
		t.Fatalf("The pipeline repo URL should have been replaced")
	}
}

func TestReplaceByName(t *testing.T) {
	ap := NewActivePipelines()

	p1 := belaur.Pipeline{
		Name:    "Pipeline A",
		Type:    belaur.PTypeGolang,
		Created: time.Now(),
	}
	ap.Append(p1)

	p2 := belaur.Pipeline{
		Name:    "Pipeline B",
		Type:    belaur.PTypeGolang,
		Created: time.Now(),
	}

	ap.ReplaceByName("Pipeline A", p2)

	ret := ap.GetByName("Pipeline B")

	if p2.Name != ret.Name {
		t.Fatalf("Pipeline should have been updated.")
	}
}

func TestIter(t *testing.T) {
	ap := NewActivePipelines()

	var pipelineNames = []string{"Pipeline A", "Pipeline B", "Pipeline C"}
	var retrievedNames []string

	for _, n := range pipelineNames {
		p := belaur.Pipeline{
			Name:    n,
			Type:    belaur.PTypeGolang,
			Created: time.Now(),
		}
		ap.Append(p)
	}

	count := 0
	for _, pipeline := range ap.GetAll() {
		count++
		retrievedNames = append(retrievedNames, pipeline.Name)
	}

	if count != len(pipelineNames) {
		t.Fatalf("Expected %d pipelines. Got %d.", len(pipelineNames), count)
	}

	for i := range retrievedNames {
		if pipelineNames[i] != retrievedNames[i] {
			t.Fatalf("The pipeline names do not match")
		}
	}
}

func TestContains(t *testing.T) {
	ap := NewActivePipelines()

	p1 := belaur.Pipeline{
		Name:    "Pipeline A",
		Type:    belaur.PTypeGolang,
		Created: time.Now(),
	}
	ap.Append(p1)

	ret := ap.Contains("Pipeline A")
	if !ret {
		t.Fatalf("Expected Pipeline A to be present in active pipelines.")
	}
}

func TestRemoveDeletedPipelines(t *testing.T) {
	ap := NewActivePipelines()

	p1 := belaur.Pipeline{
		Name:    "Pipeline A",
		Type:    belaur.PTypeGolang,
		Created: time.Now(),
	}
	ap.Append(p1)

	p2 := belaur.Pipeline{
		Name:    "Pipeline B",
		Type:    belaur.PTypeGolang,
		Created: time.Now(),
	}
	ap.Append(p2)

	p3 := belaur.Pipeline{
		Name:    "Pipeline C",
		Type:    belaur.PTypeGolang,
		Created: time.Now(),
	}
	ap.Append(p3)

	// Let's assume Pipeline B was deleted.
	existingPipelineNames := []string{"Pipeline A", "Pipeline C"}

	ap.RemoveDeletedPipelines(existingPipelineNames)

	for _, pipeline := range ap.GetAll() {
		if pipeline.Name == "Pipeline B" {
			t.Fatalf("Pipeline B still exists. It should have been removed.")
		}
	}

}

func TestRenameBinary(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestRenameBinary")
	belaur.Cfg = new(belaur.Config)
	belaur.Cfg.PipelinePath = tmp
	belaur.Cfg.HomePath = tmp
	belaur.Cfg.DataPath = tmp
	defer os.Remove("_golang")

	p := belaur.Pipeline{
		Name:    "PipelineA",
		Type:    belaur.PTypeGolang,
		Created: time.Now(),
	}

	newName := "PipelineB"

	src := filepath.Join(tmp, pipelinehelper.AppendTypeToName(p.Name, p.Type))
	dst := filepath.Join(tmp, pipelinehelper.AppendTypeToName(newName, p.Type))
	f, _ := os.Create(src)
	defer f.Close()
	defer os.Remove(src)
	defer os.Remove(dst)

	_ = ioutil.WriteFile(src, []byte("testcontent"), 0666)

	err := RenameBinary(p, newName)
	if err != nil {
		t.Fatal("an error occurred while renaming the binary: ", err)
	}

	content, err := ioutil.ReadFile(dst)
	if err != nil {
		t.Fatal("an error occurred while reading destination file: ", err)
	}
	if string(content) != "testcontent" {
		t.Fatal("file content does not equal src content. was: ", string(content))
	}

}

func TestDeleteBinary(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestDeleteBinary")
	belaur.Cfg = new(belaur.Config)
	belaur.Cfg.PipelinePath = tmp
	belaur.Cfg.HomePath = tmp
	belaur.Cfg.DataPath = tmp

	p := belaur.Pipeline{
		Name:    "PipelineA",
		Type:    belaur.PTypeGolang,
		Created: time.Now(),
	}

	src := filepath.Join(tmp, pipelinehelper.AppendTypeToName(p.Name, p.Type))
	f, _ := os.Create(src)
	defer f.Close()
	defer os.Remove(src)

	_ = ioutil.WriteFile(src, []byte("testcontent"), 0666)

	err := DeleteBinary(p)
	if err != nil {
		t.Fatal("an error occurred while deleting the binary: ", err)
	}

	_, err = os.Stat(src)
	if !os.IsNotExist(err) {
		t.Fatal("the binary file still exists. It should have been deleted")
	}
}

func TestGetExecPath(t *testing.T) {
	tmp, _ := ioutil.TempDir("", "TestGetExecPath")
	belaur.Cfg = new(belaur.Config)
	belaur.Cfg.PipelinePath = tmp
	belaur.Cfg.DataPath = tmp
	belaur.Cfg.HomePath = tmp

	p := belaur.Pipeline{
		Name:    "Pipeline A",
		Type:    belaur.PTypeGolang,
		Created: time.Now(),
	}

	expectedPath := filepath.Join(tmp, pipelinehelper.AppendTypeToName(p.Name, p.Type))
	execPath := GetExecPath(p)

	if execPath != expectedPath {
		t.Fatalf("expected execpath to be %s. got %s", expectedPath, execPath)
	}
}

func TestNewBuildPipeline(t *testing.T) {
	goBuildPipeline := newBuildPipeline(belaur.PTypeGolang)
	if goBuildPipeline == nil {
		t.Fatalf("should be of type %s but is nil\n", belaur.PTypeGolang)
	}
	javaBuildPipeline := newBuildPipeline(belaur.PTypeJava)
	if javaBuildPipeline == nil {
		t.Fatalf("should be of type %s but is nil\n", belaur.PTypeJava)
	}
	pythonBuildPipeline := newBuildPipeline(belaur.PTypePython)
	if pythonBuildPipeline == nil {
		t.Fatalf("should be of type %s but is nil\n", belaur.PTypePython)
	}
	cppBuildPipeline := newBuildPipeline(belaur.PTypeCpp)
	if cppBuildPipeline == nil {
		t.Fatalf("should be of type %s but is nil\n", belaur.PTypeCpp)
	}
	rubyBuildPipeline := newBuildPipeline(belaur.PTypeRuby)
	if rubyBuildPipeline == nil {
		t.Fatalf("should be of type %s but is nil\n", belaur.PTypeRuby)
	}
	nodeJSBuildPipeline := newBuildPipeline(belaur.PTypeNodeJS)
	if nodeJSBuildPipeline == nil {
		t.Fatalf("should be of type %s but is nil\n", belaur.PTypeNodeJS)
	}
}
