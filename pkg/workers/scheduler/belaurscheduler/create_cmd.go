package belaurscheduler

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
	"os/exec"
	"path/filepath"

	belaur "github.com/bhojpur/belaur"
	"gopkg.in/yaml.v2"
)

// createPipelineCmd creates the execute command for the plugin system
// dependent on the plugin type.
func createPipelineCmd(p *belaur.Pipeline) *exec.Cmd {
	if p == nil {
		return nil
	}
	c := &exec.Cmd{}

	// Dependent on the pipeline type
	switch p.Type {
	case belaur.PTypeGolang:
		c.Path = p.ExecPath
	case belaur.PTypeJava:
		// Look for java executable
		path, err := exec.LookPath(javaExecName)
		if err != nil {
			belaur.Cfg.Logger.Error("cannot find java executable", "error", err.Error())
			return nil
		}

		// Build start command
		c.Path = path
		c.Args = []string{
			path,
			"-jar",
			p.ExecPath,
		}
	case belaur.PTypePython:
		// Build start command
		c.Path = "/bin/sh"
		c.Args = []string{
			"/bin/sh",
			"-c",
			". bin/activate; exec " + pythonExecName + " -c \"import pipeline; pipeline.main()\"",
		}
		c.Dir = filepath.Join(belaur.Cfg.HomePath, belaur.TmpFolder, belaur.TmpPythonFolder, p.Name)
	case belaur.PTypeCpp:
		c.Path = p.ExecPath
	case belaur.PTypeRuby:
		// Look for ruby executable
		path, err := exec.LookPath(rubyExecName)
		if err != nil {
			belaur.Cfg.Logger.Error("cannot find ruby executable", "error", err.Error())
			return nil
		}

		// Get the gem name from the gem file.
		gemName, err := findRubyGemName(p.ExecPath)
		if err != nil {
			belaur.Cfg.Logger.Error("cannot find the gem name from the gem file", "error", err.Error())
			return nil
		}

		// Build start command
		c.Path = path
		c.Args = []string{
			path,
			"-r",
			gemName,
			"-e",
			"Main.main",
		}
	case belaur.PTypeNodeJS:
		// Look for node executable
		path, err := exec.LookPath(nodeJSExecName)
		if err != nil {
			belaur.Cfg.Logger.Error("cannot find NodeJS executable", "error", err)
			return nil
		}

		// Define the folder where the nodejs plugin is located unpacked
		unpackedFolder := filepath.Join(belaur.Cfg.HomePath, belaur.TmpFolder, belaur.TmpNodeJSFolder, p.Name)

		// Build start command
		c.Path = path
		c.Args = []string{
			path,
			"index.js",
		}
		c.Dir = unpackedFolder
	default:
		c = nil
	}

	return c
}

var findRubyGemCommands = []string{"specification", "--yaml"}

// findRubyGemName finds the gem name of a ruby gem file.
func findRubyGemName(execPath string) (name string, err error) {
	// Find the gem binary path.
	path, err := exec.LookPath(rubyGemName)
	if err != nil {
		return
	}

	// Get the gem specification in YAML format.
	gemCommands := append(findRubyGemCommands, execPath)
	cmd := exec.Command(path, gemCommands...)
	output, err := cmd.CombinedOutput()
	if err != nil {
		belaur.Cfg.Logger.Debug("output", "output", string(output[:]))
		return
	}

	// Struct helper to filter for what we need.
	type gemSpecOutput struct {
		Name string
	}

	// Transform and filter the gem specification.
	gemSpec := gemSpecOutput{}
	err = yaml.Unmarshal(output, &gemSpec)
	if err != nil {
		return
	}
	name = gemSpec.Name
	return
}
