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
	"fmt"
	"os"
	"os/exec"
	"path/filepath"

	belaur "github.com/bhojpur/belaur"
	"github.com/bhojpur/belaur/pkg/helper/filehelper"
)

var (
	// virtualEnvName is the binary name of virtual environment app.
	virtualEnvName = "virtualenv"

	// pythonPipInstallCmd is the command used to install the python distribution
	// package.
	pythonPipInstallCmd = ". bin/activate; python -m pip install '%s.tar.gz'"

	// Ruby gem binary name.
	rubyGemName = "gem"

	// Tar binary name.
	tarName = "tar"

	// NPM binary name.
	npmName = "npm"
)

// updatePipeline executes update steps dependent on the pipeline type.
// Some pipeline types may don't require this.
func updatePipeline(p *belaur.Pipeline) error {
	switch p.Type {
	case belaur.PTypePython:
		// Remove virtual environment if exists
		virtualEnvPath := filepath.Join(belaur.Cfg.HomePath, belaur.TmpFolder, belaur.TmpPythonFolder, p.Name)
		_ = os.RemoveAll(virtualEnvPath)

		// Create virtual environment
		path, err := exec.LookPath(virtualEnvName)
		if err != nil {
			return errors.New("cannot find virtualenv executable")
		}
		cmd := exec.Command(path, virtualEnvPath)
		if err := cmd.Run(); err != nil {
			return err
		}

		// copy distribution file to environment and remove pipeline type at the end.
		// we have to do this otherwise pip will fail.
		err = filehelper.CopyFileContents(p.ExecPath, filepath.Join(virtualEnvPath, p.Name+".tar.gz"))
		if err != nil {
			return err
		}

		// install plugin in this environment
		cmd = exec.Command("/bin/sh", "-c", fmt.Sprintf(pythonPipInstallCmd, filepath.Join(virtualEnvPath, p.Name)))
		cmd.Dir = virtualEnvPath
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("cannot install python plugin: %s", string(out[:]))
		}
	case belaur.PTypeRuby:
		// Find gem binary in path variable.
		path, err := exec.LookPath(rubyGemName)
		if err != nil {
			return err
		}

		// Gem expects that the file suffix is ".gem".
		// Copy gem file to temp folder before we install it.
		tmpFolder := filepath.Join(belaur.Cfg.HomePath, belaur.TmpFolder, belaur.TmpRubyFolder)
		err = os.MkdirAll(tmpFolder, 0700)
		if err != nil {
			return err
		}
		pipelineCopyPath := filepath.Join(tmpFolder, filepath.Base(p.ExecPath)+".gem")
		err = filehelper.CopyFileContents(p.ExecPath, pipelineCopyPath)
		if err != nil {
			return err
		}
		defer os.Remove(pipelineCopyPath)

		// Install gem forcefully.
		cmd := exec.Command(path, "install", "-f", pipelineCopyPath)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("cannot install ruby gem: %s", string(out[:]))
		}
	case belaur.PTypeNodeJS:
		// Find tar binary in path
		path, err := exec.LookPath(tarName)
		if err != nil {
			return err
		}

		// Delete old folders if exist
		tmpFolder := filepath.Join(belaur.Cfg.HomePath, belaur.TmpFolder, belaur.TmpNodeJSFolder, p.Name)
		_ = os.RemoveAll(tmpFolder)

		// Recreate the temp folder
		if err := os.MkdirAll(tmpFolder, 0700); err != nil {
			return err
		}

		// Unpack it
		cmd := exec.Command(path, "-xzvf", p.ExecPath, "-C", tmpFolder)
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("cannot unpack nodejs archive: %s", string(out[:]))
		}

		// Find npm binary in path
		path, err = exec.LookPath(npmName)
		if err != nil {
			return err
		}

		// Install dependencies
		cmd = &exec.Cmd{
			Path: path,
			Dir:  tmpFolder,
			Args: []string{path, "install"},
		}
		if out, err := cmd.CombinedOutput(); err != nil {
			return fmt.Errorf("cannot install dependencies: %s", string(out[:]))
		}
	}

	// Update checksum
	checksum, err := filehelper.GetSHA256Sum(p.ExecPath)
	if err != nil {
		return err
	}
	p.SHA256Sum = checksum

	return nil
}
