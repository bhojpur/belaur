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
	"errors"
	"io/ioutil"
	"os"
	"os/exec"
	"path/filepath"
	"strings"
	"time"

	belaur "github.com/bhojpur/belaur"
	"github.com/bhojpur/belaur/pkg/helper/filehelper"
	"github.com/bhojpur/belaur/pkg/helper/pipelinehelper"

	"github.com/bhojpur/belaur/pkg/services"
	"github.com/gofrs/uuid"
)

var (
	pythonBinaryName = "python"
)

// BuildPipelinePython is the real implementation of BuildPipeline for python
type BuildPipelinePython struct {
	Type belaur.PipelineType
}

// PrepareEnvironment prepares the environment before we start the build process.
func (b *BuildPipelinePython) PrepareEnvironment(p *belaur.CreatePipeline) error {
	// create uniqueName for destination folder
	v4, err := uuid.NewV4()
	if err != nil {
		belaur.Cfg.Logger.Debug("unable to generate uuid", "error", err.Error())
		return err
	}
	uniqueName := uuid.Must(v4, nil)

	// Create local temp folder for clone
	rootPath := filepath.Join(belaur.Cfg.HomePath, belaur.TmpFolder, belaur.TmpPythonFolder)
	cloneFolder := filepath.Join(rootPath, belaur.SrcFolder, uniqueName.String())
	err = os.MkdirAll(cloneFolder, 0700)
	if err != nil {
		return err
	}

	// Set new generated path in pipeline obj for later usage
	if p.Pipeline.Repo == nil {
		p.Pipeline.Repo = &belaur.GitRepo{}
	}
	p.Pipeline.Repo.LocalDest = cloneFolder
	p.Pipeline.UUID = uniqueName.String()
	return nil
}

// ExecuteBuild executes the python build process
func (b *BuildPipelinePython) ExecuteBuild(p *belaur.CreatePipeline) error {
	// Look for python executeable
	path, err := exec.LookPath(pythonBinaryName)
	if err != nil {
		belaur.Cfg.Logger.Debug("cannot find python executeable", "error", err.Error())
		return err
	}

	// Set command args for build distribution package
	args := []string{
		"setup.py",
		"sdist",
	}

	// Set local destination
	localDest := ""
	if p.Pipeline.Repo != nil {
		localDest = p.Pipeline.Repo.LocalDest
	}

	// Execute and wait until finish or timeout
	output, err := executeCmd(path, args, os.Environ(), localDest)
	if err != nil {
		belaur.Cfg.Logger.Debug("cannot generate python distribution package", "error", err.Error(), "output", string(output))
		p.Output = string(output)
		return err
	}

	// Build has been finished. Set execution path to the build result archive.
	// This will be used during pipeline verification phase which will happen after this step.
	p.Pipeline.ExecPath, err = findPythonArchivePath(p)
	if err != nil {
		return err
	}

	return nil
}

// findPythonArchivePath filters the archives in the generated dist folder
// and looks for the final archive. It will return an error if less or more
// than one file(s) are found otherwise the full path to the file.
func findPythonArchivePath(p *belaur.CreatePipeline) (src string, err error) {
	// find all files in dist folder
	distFolder := filepath.Join(p.Pipeline.Repo.LocalDest, "dist")
	files, err := ioutil.ReadDir(distFolder)
	if err != nil {
		return
	}

	// filter for archives
	archive := []os.FileInfo{}
	for _, file := range files {
		if strings.HasSuffix(file.Name(), ".tar.gz") {
			archive = append(archive, file)
		}
	}

	// if we found more or less than one archive we have a problem
	if len(archive) != 1 {
		belaur.Cfg.Logger.Debug("cannot find python package", "foundPackages", len(archive), "archives", files)
		err = errors.New("cannot find python package")
		return
	}

	// Return full path
	src = filepath.Join(distFolder, archive[0].Name())
	return
}

// CopyBinary copies the final compiled archive to the
// destination folder.
func (b *BuildPipelinePython) CopyBinary(p *belaur.CreatePipeline) error {
	// Define src and destination
	src, err := findPythonArchivePath(p)
	if err != nil {
		return err
	}
	dest := filepath.Join(belaur.Cfg.PipelinePath, pipelinehelper.AppendTypeToName(p.Pipeline.Name, p.Pipeline.Type))

	// Copy binary
	if err := filehelper.CopyFileContents(src, dest); err != nil {
		return err
	}

	// Set +x (execution right) for pipeline
	return os.Chmod(dest, belaur.ExecutablePermission)
}

// SavePipeline saves the current pipeline configuration.
func (b *BuildPipelinePython) SavePipeline(p *belaur.Pipeline) error {
	dest := filepath.Join(belaur.Cfg.PipelinePath, pipelinehelper.AppendTypeToName(p.Name, p.Type))
	p.ExecPath = dest
	p.Type = belaur.PTypePython
	p.Name = strings.TrimSuffix(filepath.Base(dest), typeDelimiter+belaur.PTypePython.String())
	p.Created = time.Now()
	// Our pipeline is finished constructing. Save it.
	storeService, _ := services.StorageService()
	return storeService.PipelinePut(p)
}
