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

const nodeJSInternalCloneFolder = "jsclone"

// BuildPipelineNodeJS is the real implementation of BuildPipeline for NodeJS
type BuildPipelineNodeJS struct {
	Type belaur.PipelineType
}

// PrepareEnvironment prepares the environment before we start the build process.
func (b *BuildPipelineNodeJS) PrepareEnvironment(p *belaur.CreatePipeline) error {
	// create uniqueName for destination folder
	v4, err := uuid.NewV4()
	if err != nil {
		belaur.Cfg.Logger.Debug("unable to generate uuid", "error", err.Error())
		return err
	}
	uniqueName := uuid.Must(v4, nil)

	// Create local temp folder for clone
	cloneFolder := filepath.Join(belaur.Cfg.HomePath, belaur.TmpFolder, belaur.TmpNodeJSFolder, belaur.SrcFolder, uniqueName.String(), nodeJSInternalCloneFolder)
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

// ExecuteBuild executes the NodeJS build process
func (b *BuildPipelineNodeJS) ExecuteBuild(p *belaur.CreatePipeline) error {
	// Look for Tar binary executable
	path, err := exec.LookPath(tarName)
	if err != nil {
		belaur.Cfg.Logger.Debug("cannot find tar binary executable", "error", err.Error())
		return err
	}

	// Set local destination
	localDest := ""
	if p.Pipeline.Repo != nil {
		localDest = p.Pipeline.Repo.LocalDest
	}

	// Set command args for archive process
	pipelineFileName := pipelinehelper.AppendTypeToName(p.Pipeline.Name, p.Pipeline.Type)
	args := []string{
		"--exclude=.git",
		"-czvf",
		pipelineFileName,
		"-C",
		localDest,
		".",
	}

	// Execute and wait until finish or timeout
	uniqueFolder := filepath.Join(belaur.Cfg.HomePath, belaur.TmpFolder, belaur.TmpNodeJSFolder, belaur.SrcFolder, p.Pipeline.UUID)
	output, err := executeCmd(path, args, os.Environ(), uniqueFolder)
	p.Output = string(output[:])
	if err != nil {
		belaur.Cfg.Logger.Debug("cannot build pipeline", "error", err.Error(), "output", string(output[:]))
		return err
	}

	// Build has been finished. Set execution path to the build result archive.
	// This will be used during pipeline verification phase which will happen after this step.
	p.Pipeline.ExecPath = filepath.Join(uniqueFolder, pipelineFileName)

	// Set the the local destination variable to the unique folder because this is now the place
	// where our binary is located.
	p.Pipeline.Repo.LocalDest = uniqueFolder

	return nil
}

// CopyBinary copies the final compiled binary to the
// destination folder.
func (b *BuildPipelineNodeJS) CopyBinary(p *belaur.CreatePipeline) error {
	// Define src and destination
	src := filepath.Join(p.Pipeline.Repo.LocalDest, pipelinehelper.AppendTypeToName(p.Pipeline.Name, p.Pipeline.Type))
	dest := filepath.Join(belaur.Cfg.PipelinePath, pipelinehelper.AppendTypeToName(p.Pipeline.Name, p.Pipeline.Type))

	// Copy binary
	if err := filehelper.CopyFileContents(src, dest); err != nil {
		return err
	}

	// Set +x (execution right) for pipeline
	return os.Chmod(dest, belaur.ExecutablePermission)
}

// SavePipeline saves the current pipeline configuration.
func (b *BuildPipelineNodeJS) SavePipeline(p *belaur.Pipeline) error {
	dest := filepath.Join(belaur.Cfg.PipelinePath, pipelinehelper.AppendTypeToName(p.Name, p.Type))
	p.ExecPath = dest
	p.Type = belaur.PTypeNodeJS
	p.Name = strings.TrimSuffix(filepath.Base(dest), typeDelimiter+belaur.PTypeNodeJS.String())
	p.Created = time.Now()
	// Our pipeline is finished constructing. Save it.
	storeService, _ := services.StorageService()
	return storeService.PipelinePut(p)
}
