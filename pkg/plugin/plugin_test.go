package plugin

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
	"bufio"
	"bytes"
	"context"
	"crypto/tls"
	"io"
	"io/ioutil"
	"os/exec"
	"path/filepath"
	"strings"
	"testing"
	"time"

	belaur "github.com/bhojpur/belaur"
	proto "github.com/bhojpur/belaur/pkg/api/v1/plugin"
	"github.com/hashicorp/go-hclog"
	"google.golang.org/grpc/metadata"
)

type fakeCAAPI struct{}

func (c *fakeCAAPI) CreateSignedCertWithValidOpts(hostname string, hoursBeforeValid, hoursAfterValid time.Duration) (string, string, error) {
	return "", "", nil
}
func (c *fakeCAAPI) CreateSignedCert() (string, string, error) { return "", "", nil }
func (c *fakeCAAPI) GenerateTLSConfig(certPath, keyPath string) (*tls.Config, error) {
	return &tls.Config{}, nil
}
func (c *fakeCAAPI) CleanupCerts(crt, key string) error { return nil }
func (c *fakeCAAPI) GetCACertPath() (string, string)    { return "", "" }

type fakeClientProtocol struct{}

func (cp *fakeClientProtocol) Dispense(s string) (interface{}, error) {
	return &fakeBelaurPlugin{}, nil
}
func (cp *fakeClientProtocol) Ping() error  { return nil }
func (cp *fakeClientProtocol) Close() error { return nil }

type fakeBelaurPlugin struct{}

func (p *fakeBelaurPlugin) GetJobs() (proto.Plugin_GetJobsClient, error) {
	return &fakeJobsClient{}, nil
}
func (p *fakeBelaurPlugin) ExecuteJob(job *proto.Job) (*proto.JobResult, error) {
	return &proto.JobResult{}, nil
}

type fakeJobsClient struct {
	counter int
}

func (jc *fakeJobsClient) Recv() (*proto.Job, error) {
	j := &proto.Job{
		Args: []*proto.Argument{
			{
				Key:   "key",
				Value: "value",
			},
		},
	}

	if jc.counter == 0 {
		jc.counter++
		return j, nil
	}
	return nil, io.EOF
}
func (jc *fakeJobsClient) Header() (metadata.MD, error) { return nil, nil }
func (jc *fakeJobsClient) Trailer() metadata.MD         { return nil }
func (jc *fakeJobsClient) CloseSend() error             { return nil }
func (jc *fakeJobsClient) Context() context.Context     { return nil }
func (jc *fakeJobsClient) SendMsg(m interface{}) error  { return nil }
func (jc *fakeJobsClient) RecvMsg(m interface{}) error  { return nil }

func TestNewPlugin(t *testing.T) {
	p := &GoPlugin{}
	p.NewPlugin(new(fakeCAAPI))
}

func TestInit(t *testing.T) {
	belaur.Cfg = &belaur.Config{}
	tmp, _ := ioutil.TempDir("", "TestInit")
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: hclog.DefaultOutput,
		Name:   "Belaur",
	})
	emptyPlugin := &GoPlugin{}
	p := emptyPlugin.NewPlugin(new(fakeCAAPI))
	logpath := filepath.Join(tmp, "test")
	err := p.Init(exec.Command("echo", "world"), &logpath)
	if err == nil {
		t.Fatal("was expecting an error. non happened")
	}
	if !strings.Contains(err.Error(), "Unrecognized remote plugin message") {
		// Sometimes go-plugin throws this error instead...
		if !strings.Contains(err.Error(), "plugin exited before we could connect") {
			t.Fatalf("Error should contain 'Unrecognized remote plugin message' but was '%s'", err.Error())
		}
	}
}

func TestValidate(t *testing.T) {
	belaur.Cfg = &belaur.Config{}
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: hclog.DefaultOutput,
		Name:   "Belaur",
	})
	p := &GoPlugin{clientProtocol: new(fakeClientProtocol)}
	err := p.Validate()
	if err != nil {
		t.Fatal(err)
	}
}

func TestExecute(t *testing.T) {
	belaur.Cfg = &belaur.Config{}
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: hclog.DefaultOutput,
		Name:   "Belaur",
	})
	p := &GoPlugin{pluginConn: new(fakeBelaurPlugin)}
	buf := new(bytes.Buffer)
	p.logger = BelaurLogWriter{}
	p.logger.writer = bufio.NewWriter(buf)
	j := &belaur.Job{
		Args: []*belaur.Argument{
			{
				Key:   "key",
				Value: "value",
			},
		},
	}
	err := p.Execute(j)
	if err != nil {
		t.Fatal(err)
	}
}

func TestGetJobs(t *testing.T) {
	belaur.Cfg = &belaur.Config{}
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: hclog.DefaultOutput,
		Name:   "Belaur",
	})
	p := &GoPlugin{pluginConn: new(fakeBelaurPlugin)}
	buf := new(bytes.Buffer)
	p.logger = BelaurLogWriter{}
	p.logger.writer = bufio.NewWriter(buf)
	_, err := p.GetJobs()
	if err != nil {
		t.Fatal(err)
	}
}

func TestClose(t *testing.T) {
	belaur.Cfg = &belaur.Config{}
	tmp, _ := ioutil.TempDir("", "TestInit")
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: hclog.DefaultOutput,
		Name:   "Belaur",
	})
	emptyPlugin := &GoPlugin{}
	p := emptyPlugin.NewPlugin(new(fakeCAAPI))
	logpath := filepath.Join(tmp, "test")
	_ = p.Init(exec.Command("echo", "world"), &logpath)
	p.Close()
}

func TestFlushLogs(t *testing.T) {
	belaur.Cfg = &belaur.Config{}
	tmp, _ := ioutil.TempDir("", "TestInit")
	belaur.Cfg.Logger = hclog.New(&hclog.LoggerOptions{
		Level:  hclog.Trace,
		Output: hclog.DefaultOutput,
		Name:   "Belaur",
	})
	emptyPlugin := &GoPlugin{}
	p := emptyPlugin.NewPlugin(new(fakeCAAPI))
	logpath := filepath.Join(tmp, "test")
	_ = p.Init(exec.Command("echo", "world"), &logpath)
	err := p.FlushLogs()
	if err != nil {
		t.Fatal(err)
	}
}
