package server

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
	"testing"
	"time"

	belaur "github.com/bhojpur/belaur"
	"github.com/bhojpur/belaur/pkg/security"

	hclog "github.com/hashicorp/go-hclog"
)

func TestStart(t *testing.T) {
	// Create tmp folder
	tmpFolder, err := ioutil.TempDir("", "TestStart")
	if err != nil {
		t.Fatal(err)
	}
	defer os.RemoveAll(tmpFolder)

	belaur.Cfg = &belaur.Config{
		Mode:              belaur.ModeServer,
		WorkerGRPCHostURL: "myhost:12345",
		HomePath:          tmpFolder,
		DataPath:          tmpFolder,
		CAPath:            tmpFolder,
	}
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level: hclog.Trace,
		Name:  "Belaur",
	})

	ca, _ := security.InitCA()
	// Init worker server
	server := InitWorkerServer(Dependencies{Certificate: ca})

	// Start server
	errChan := make(chan error)
	go func() {
		if err := server.Start(); err != nil {
			errChan <- err
		}
	}()
	time.Sleep(3 * time.Second)
	select {
	case err := <-errChan:
		t.Fatal(err)
	default:
	}
}
