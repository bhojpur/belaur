package pipelinehelper

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
	"testing"

	belaur "github.com/bhojpur/belaur"
	"github.com/bhojpur/belaur/pkg/helper/stringhelper"
)

func TestGetRealPipelineName(t *testing.T) {
	pipeGo := "my_pipeline_golang"
	if GetRealPipelineName(pipeGo, belaur.PTypeGolang) != "my_pipeline" {
		t.Fatalf("output should be my_pipeline but is %s", GetRealPipelineName(pipeGo, belaur.PTypeGolang))
	}
}

func TestAppendTypeToName(t *testing.T) {
	expected := []string{"my_pipeline_golang", "my_pipeline2_java", "my_pipeline_python"}
	input := []struct {
		name  string
		pType belaur.PipelineType
	}{
		{
			"my_pipeline",
			belaur.PTypeGolang,
		},
		{
			"my_pipeline2",
			belaur.PTypeJava,
		},
		{
			"my_pipeline",
			belaur.PTypePython,
		},
	}

	for _, inp := range input {
		got := AppendTypeToName(inp.name, inp.pType)
		if !stringhelper.IsContainedInSlice(expected, got, false) {
			t.Fatalf("expected name not contained: %s", got)
		}
	}
}
