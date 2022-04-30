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
	"errors"
	"fmt"
	"path/filepath"
	"strings"

	belaur "github.com/bhojpur/belaur"
)

const (
	typeDelimiter = "_"
)

// GetRealPipelineName removes the suffix from the pipeline name.
func GetRealPipelineName(name string, pType belaur.PipelineType) string {
	return strings.TrimSuffix(name, typeDelimiter+pType.String())
}

// AppendTypeToName appends the type to the output binary name.
// This allows later to define the pipeline type by the pipeline binary name.
func AppendTypeToName(n string, pType belaur.PipelineType) string {
	return fmt.Sprintf("%s%s%s", n, typeDelimiter, pType.String())
}

// GetLocalDestinationForPipeline computes the local location of a pipeline on disk based on the pipeline's
// type and configuration of Bhojpur Belaur such as, temp folder and data folder.
func GetLocalDestinationForPipeline(p belaur.Pipeline) (string, error) {
	tmpFolder, err := tmpFolderFromPipelineType(p)
	if err != nil {
		belaur.Cfg.Logger.Error("Pipeline type invalid", "type", p.Type)
		return "", err
	}
	return filepath.Join(belaur.Cfg.HomePath, belaur.TmpFolder, tmpFolder, belaur.SrcFolder, p.UUID), nil
}

// tmpFolderFromPipelineType returns the Bhojpur Belaur specific tmp folder for a pipeline
// based on the type of the pipeline.
func tmpFolderFromPipelineType(foundPipeline belaur.Pipeline) (string, error) {
	switch foundPipeline.Type {
	case belaur.PTypeCpp:
		return belaur.TmpCppFolder, nil
	case belaur.PTypeGolang:
		return belaur.TmpGoFolder, nil
	case belaur.PTypeNodeJS:
		return belaur.TmpNodeJSFolder, nil
	case belaur.PTypePython:
		return belaur.TmpPythonFolder, nil
	case belaur.PTypeRuby:
		return belaur.TmpRubyFolder, nil
	case belaur.PTypeJava:
		return belaur.TmpJavaFolder, nil
	}
	return "", errors.New("invalid pipeline type")
}
