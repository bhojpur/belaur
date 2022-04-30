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
	gemBinaryName = "gem"
)

// gemspecNameKey is the variable key which is filtered for during build.
const gemspecNameKey = "${NAME}"

// gemInitFile is the initial file of the gem.
const gemInitFile = "belaur.rb"

// BuildPipelineRuby is the real implementation of BuildPipeline for Ruby
type BuildPipelineRuby struct {
	Type        belaur.PipelineType
	GemfileName string
}

// PrepareEnvironment prepares the environment before we start the build process.
func (b *BuildPipelineRuby) PrepareEnvironment(p *belaur.CreatePipeline) error {
	// create uniqueName for destination folder
	v4, err := uuid.NewV4()
	if err != nil {
		belaur.Cfg.Logger.Debug("unable to generate uuid", "error", err.Error())
		return err
	}
	uniqueName := uuid.Must(v4, nil)

	// Create local temp folder for clone
	cloneFolder := filepath.Join(belaur.Cfg.HomePath, belaur.TmpFolder, belaur.TmpRubyFolder, belaur.SrcFolder, uniqueName.String())
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

// ExecuteBuild executes the ruby build process
func (b *BuildPipelineRuby) ExecuteBuild(p *belaur.CreatePipeline) error {
	// Look for gem binary executable
	path, err := exec.LookPath(gemBinaryName)
	if err != nil {
		belaur.Cfg.Logger.Debug("cannot find gem binary executable", "error", err.Error())
		return err
	}

	// Set local destination
	localDest := ""
	if p.Pipeline.Repo != nil {
		localDest = p.Pipeline.Repo.LocalDest
	}

	// Get all gemspec files in cloned folder.
	gemspec, err := filterPathContentBySuffix(localDest, ".gemspec")
	if err != nil {
		belaur.Cfg.Logger.Error("cannot find gemspec file in cloned repository folder", "path", localDest)
		return err
	}

	// if we found more or less than one gemspec we have a problem.
	if len(gemspec) != 1 {
		belaur.Cfg.Logger.Debug("cannot find gemspec file in cloned repo", "foundGemspecs", len(gemspec), "gemspecs", gemspec)
		return errors.New("cannot find gemspec file in cloned repo")
	}

	// Generate a new UUID for the gem name to prevent conflicts with other gems.
	v4, err := uuid.NewV4()
	if err != nil {
		return err
	}
	uuid := uuid.Must(v4, nil).String()

	// Read gemspec file.
	gemspecContent, err := ioutil.ReadFile(gemspec[0])
	if err != nil {
		belaur.Cfg.Logger.Debug("cannot read gemspec file", "error", err.Error(), "pipeline", p.Pipeline.Name)
		return err
	}

	// Replace name variable with new UUID and write content to file.
	gemspecContentStr := strings.Replace(string(gemspecContent[:]), gemspecNameKey, uuid, 1)
	err = ioutil.WriteFile(gemspec[0], []byte(gemspecContentStr), 0644)
	if err != nil {
		belaur.Cfg.Logger.Debug("cannot write/edit gemspec file", "error", err.Error(), "pipeline", p.Pipeline.Name)
		return err
	}

	// The initial ruby file in the gem must be named like the gem name.
	// We expect that the init file is always `gemInitFile`.
	err = os.Rename(filepath.Join(localDest, "lib", gemInitFile), filepath.Join(localDest, "lib", uuid+".rb"))
	if err != nil {
		belaur.Cfg.Logger.Debug("cannot rename initial ruby file", "error", err.Error(), "pipeline", p.Pipeline)
		return err
	}

	// Set command args for build
	args := []string{
		"build",
		gemspec[0],
	}

	// Execute and wait until finish or timeout
	output, err := executeCmd(path, args, os.Environ(), localDest)
	p.Output = string(output)
	if err != nil {
		belaur.Cfg.Logger.Debug("cannot build pipeline", "error", err.Error(), "output", string(output))
		return err
	}

	// Search for resulting gem file.
	gemfile, err := findGemFileByGlob(localDest, uuid+"*.gem")
	if err != nil {
		belaur.Cfg.Logger.Error("cannot find final gem file after build", "path", p.Pipeline.Repo.LocalDest)
		return err
	}

	// fallback to the old method because we don't have a uuid based gem file.
	if gemfile == nil {
		gemfile, err = findGemFileByGlob(localDest, "*.gem")
		if err != nil {
			belaur.Cfg.Logger.Error("cannot find final gem file after build", "path", p.Pipeline.Repo.LocalDest)
			return err
		}
	}

	// if we found more or less than one gem file for the given uuid then we have a problem.
	if len(gemfile) != 1 {
		belaur.Cfg.Logger.Debug("cannot find gem file in cloned repo", "gems", gemfile)
		return errors.New("cannot find gem file in cloned repo")
	}

	// Build has been finished. Set execution path to the build result archive.
	// This will be used during pipeline verification phase which will happen after this step.
	p.Pipeline.ExecPath = gemfile[0]
	b.GemfileName = gemfile[0]
	return nil
}

// filterPathContentBySuffix reads the whole directory given by path and
// returns all files with full path which have the same suffix like provided.
func filterPathContentBySuffix(path, suffix string) ([]string, error) {
	filteredFiles := []string{}

	// Read complete directory.
	files, err := ioutil.ReadDir(path)
	if err != nil {
		return filteredFiles, err
	}

	// filter for files ending with given suffix.
	for _, file := range files {
		if strings.HasSuffix(file.Name(), suffix) {
			filteredFiles = append(filteredFiles, filepath.Join(path, file.Name()))
		}
	}
	return filteredFiles, nil
}

// findGemFileByGlob reads the whole directory given by path and
// returns all files with full path which have the same glob like provided.
func findGemFileByGlob(path, glob string) ([]string, error) {
	fullPath := filepath.Join(path, glob)
	return filepath.Glob(fullPath)
}

// CopyBinary copies the final compiled binary to the
// destination folder.
func (b *BuildPipelineRuby) CopyBinary(p *belaur.CreatePipeline) error {
	// Define src and destination
	src := b.GemfileName
	belaur.Cfg.Logger.Debug("Copying over ruby gem file", "gem", src)
	dest := filepath.Join(belaur.Cfg.PipelinePath, pipelinehelper.AppendTypeToName(p.Pipeline.Name, p.Pipeline.Type))

	// Copy binary
	if err := filehelper.CopyFileContents(src, dest); err != nil {
		return err
	}

	// Set +x (execution right) for pipeline
	return os.Chmod(dest, belaur.ExecutablePermission)
}

// SavePipeline saves the current pipeline configuration.
func (b *BuildPipelineRuby) SavePipeline(p *belaur.Pipeline) error {
	dest := filepath.Join(belaur.Cfg.PipelinePath, pipelinehelper.AppendTypeToName(p.Name, p.Type))
	p.ExecPath = dest
	p.Type = belaur.PTypeRuby
	p.Name = strings.TrimSuffix(filepath.Base(dest), typeDelimiter+belaur.PTypeRuby.String())
	p.Created = time.Now()
	// Our pipeline is finished constructing. Save it.
	storeService, _ := services.StorageService()
	return storeService.PipelinePut(p)
}
