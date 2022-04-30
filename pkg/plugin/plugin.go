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
	"errors"
	"fmt"
	"io"
	"os"
	"os/exec"
	"sync"
	"time"

	belaur "github.com/bhojpur/belaur"
	proto "github.com/bhojpur/belaur/pkg/api/v1/plugin"
	"github.com/bhojpur/belaur/pkg/security"
	"github.com/hashicorp/go-plugin"
)

const (
	pluginMapKey = "Plugin"

	// env variable key names for TLS cert path
	serverCertEnv = "BELAUR_PLUGIN_CERT"
	serverKeyEnv  = "BELAUR_PLUGIN_KEY"
	rootCACertEnv = "BELAUR_PLUGIN_CA_CERT"
)

var handshake = plugin.HandshakeConfig{
	ProtocolVersion: 2,
	MagicCookieKey:  "BELAUR_PLUGIN",
	// This cookie should never be changed again
	MagicCookieValue: "FdXjW27mN6XuG2zDBP4LixXUwDAGCEkidxwqBGYpUhxiWHzctATYZvpz4ZJdALmh",
}

var pluginMap = map[string]plugin.Plugin{
	pluginMapKey: &BelaurPluginImpl{},
}

// timeFormat is the logging time format.
const timeFormat = "2006/01/02 15:04:05"

// BelaurLogWriter represents a concurrent safe log writer which can be shared with go-plugin.
type BelaurLogWriter struct {
	mu     sync.RWMutex
	buffer *bytes.Buffer
	writer *bufio.Writer
}

// Write locks and writes to the underlying writer.
func (g *BelaurLogWriter) Write(p []byte) (n int, err error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.writer.Write(p)
}

// Flush locks and flushes the underlying writer.
func (g *BelaurLogWriter) Flush() error {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.writer.Flush()
}

// WriteString locks and passes on the string to write to the underlying writer.
func (g *BelaurLogWriter) WriteString(s string) (int, error) {
	g.mu.Lock()
	defer g.mu.Unlock()
	return g.writer.WriteString(s)
}

// GoPlugin represents a single plugin instance which uses gRPC
// to connect to exactly one plugin.
type GoPlugin struct {
	// Client is an instance of the go-plugin client.
	client *plugin.Client

	// Client protocol instance used to open gRPC connections.
	clientProtocol plugin.ClientProtocol

	// Interface to the connected plugin.
	pluginConn BelaurPlugin

	// Log file where all output is stored.
	logFile *os.File

	// Writer used to write logs from execution to file or buffer
	logger BelaurLogWriter

	// CA instance used to handle certificates
	ca security.CAAPI

	// Created certificates path for pipeline run
	certPath       string
	keyPath        string
	serverCertPath string
	serverKeyPath  string
}

// Plugin represents the plugin implementation.
type Plugin interface {
	// NewPlugin creates a new instance of plugin
	NewPlugin(ca security.CAAPI) Plugin

	// Init initializes the go-plugin client and generates a
	// new certificate pair for Bhojpur Belaur and the plugin/pipeline.
	Init(command *exec.Cmd, logPath *string) error

	// Validate validates the plugin interface.
	Validate() error

	// Execute executes one job of a pipeline.
	Execute(j *belaur.Job) error

	// GetJobs returns all real jobs from the pipeline.
	GetJobs() ([]*belaur.Job, error)

	// FlushLogs flushes the logs.
	FlushLogs() error

	// Close closes the connection and cleans open file writes.
	Close()
}

// NewPlugin creates a new instance of Plugin.
// One Plugin instance represents one connection to a plugin.
func (p *GoPlugin) NewPlugin(ca security.CAAPI) Plugin {
	return &GoPlugin{ca: ca}
}

// Init prepares the log path, set's up new certificates for both Bhojpur Belaur and
// plugin, and prepares the go-plugin client.
//
// It expects the start command for the plugin and the path where
// the log file should be stored.
//
// It's up to the caller to call plugin.Close to shutdown the plugin
// and close the gRPC connection.
func (p *GoPlugin) Init(command *exec.Cmd, logPath *string) error {
	// Initialise the logger
	p.logger = BelaurLogWriter{}

	// Create log file and open it.
	// We will close this file in the close method.
	if logPath != nil {
		var err error
		p.logFile, err = os.OpenFile(
			*logPath,
			os.O_CREATE|os.O_WRONLY,
			0666,
		)
		if err != nil {
			return err
		}

		// Create new writer
		p.logger.writer = bufio.NewWriter(p.logFile)
	} else {
		// If no path is provided, write output to buffer
		p.logger.buffer = new(bytes.Buffer)
		p.logger.writer = bufio.NewWriter(p.logger.buffer)
	}

	// Create and sign a new pair of certificates for the server
	var err error
	p.serverCertPath, p.serverKeyPath, err = p.ca.CreateSignedCert()
	if err != nil {
		return err
	}

	// Expose path of server certificates as well as public CA cert.
	// This allows the plugin to grab the certificates.
	caCert, _ := p.ca.GetCACertPath()
	command.Env = append(command.Env, serverCertEnv+"="+p.serverCertPath)
	command.Env = append(command.Env, serverKeyEnv+"="+p.serverKeyPath)
	command.Env = append(command.Env, rootCACertEnv+"="+caCert)

	// Create and sign a new pair of certificates for the client
	p.certPath, p.keyPath, err = p.ca.CreateSignedCert()
	if err != nil {
		return err
	}

	// Generate TLS config
	tlsConfig, err := p.ca.GenerateTLSConfig(p.certPath, p.keyPath)
	if err != nil {
		return err
	}

	// Get new client
	p.client = plugin.NewClient(&plugin.ClientConfig{
		HandshakeConfig:  handshake,
		Plugins:          pluginMap,
		Cmd:              command,
		AllowedProtocols: []plugin.Protocol{plugin.ProtocolGRPC},
		Stderr:           &p.logger,
		TLSConfig:        tlsConfig,
	})

	// Connect via gRPC
	p.clientProtocol, err = p.client.Client()
	if err != nil {
		_ = p.logger.Flush()
		return fmt.Errorf("%s\n\n--- output ---\n%s", err.Error(), p.logger.buffer.String())
	}

	return nil
}

// Validate validates the interface of the plugin.
func (p *GoPlugin) Validate() error {
	// Request the plugin
	raw, err := p.clientProtocol.Dispense(pluginMapKey)
	if err != nil {
		return err
	}

	// Convert plugin to interface
	if pC, ok := raw.(BelaurPlugin); ok {
		p.pluginConn = pC
		return nil
	}

	return errors.New("plugin is not compatible with plugin interface")
}

// Execute triggers the execution of one single job
// for the given plugin.
func (p *GoPlugin) Execute(j *belaur.Job) error {
	// Transform arguments
	var args []*proto.Argument
	for _, arg := range j.Args {
		a := &proto.Argument{
			Key:   arg.Key,
			Value: arg.Value,
		}

		args = append(args, a)
	}

	// Create new proto job object.
	job := &proto.Job{
		UniqueId: j.ID,
		Args:     args,
	}

	// Execute the job
	resultObj, err := p.pluginConn.ExecuteJob(job)

	// Check and set job status
	if resultObj != nil && resultObj.ExitPipeline {
		// ExitPipeline is true that indicates that the job failed.
		j.Status = belaur.JobFailed

		// Failed was set so the pipeline will now be marked as failed.
		if resultObj.Failed {
			j.FailPipeline = true
		}

		// Generate error message and attach it to logs.
		timeString := time.Now().Format(timeFormat)
		_, _ = p.logger.WriteString(fmt.Sprintf("%s Job '%s' threw an error: %s\n", timeString, j.Title, resultObj.Message))
	} else if err != nil {
		// An error occurred during the send or somewhere else.
		// The job itself usually does not return an error here.
		// We mark the job as failed.
		j.Status = belaur.JobFailed

		// Generate error message and attach it to logs.
		timeString := time.Now().Format(timeFormat)
		_, _ = p.logger.WriteString(fmt.Sprintf("%s Job '%s' threw an error: %s\n", timeString, j.Title, err.Error()))
	} else {
		j.Status = belaur.JobSuccess
	}

	return nil
}

// GetJobs receives all implemented jobs from the given plugin.
func (p *GoPlugin) GetJobs() ([]*belaur.Job, error) {
	l := make([]*belaur.Job, 0)

	// Get the stream
	stream, err := p.pluginConn.GetJobs()
	if err != nil {
		return nil, err
	}

	// receive all jobs
	pList := make([]*proto.Job, 0)
	jobsMap := make(map[uint32]*belaur.Job)
	for {
		job, err := stream.Recv()

		// Got all jobs
		if err == io.EOF {
			break
		}

		// Error during stream
		if err != nil {
			return nil, err
		}

		// Transform arguments
		args := make([]*belaur.Argument, 0, len(job.Args))
		for _, arg := range job.Args {
			a := &belaur.Argument{
				Description: arg.Description,
				Key:         arg.Key,
				Type:        arg.Type,
			}

			args = append(args, a)
		}

		// add proto object to separate list to rebuild dep later.
		pList = append(pList, job)

		// Convert proto object to belaur.Job struct
		j := &belaur.Job{
			ID:          job.UniqueId,
			Title:       job.Title,
			Description: job.Description,
			Status:      belaur.JobWaitingExec,
			Args:        args,
		}
		l = append(l, j)
		jobsMap[j.ID] = j
	}

	// Rebuild dependencies
	for _, pbJob := range pList {
		// Get job
		j := jobsMap[pbJob.UniqueId]

		// Iterate all dependencies
		j.DependsOn = make([]*belaur.Job, 0, len(pbJob.Dependson))
		for _, depJob := range pbJob.Dependson {
			// Get dependency
			depJ := jobsMap[depJob]

			// Set dependency
			j.DependsOn = append(j.DependsOn, depJ)
		}
	}

	// return list
	return l, nil
}

// FlushLogs flushes the logs.
func (p *GoPlugin) FlushLogs() error {
	return p.logger.Flush()
}

// Close shutdown the plugin and kills the gRPC connection.
// Remember to call this when you call plugin.Connect.
func (p *GoPlugin) Close() {
	// We start the kill command in a goroutine because kill
	// is blocking until the subprocess successfully exits.
	// The user should not wait for this.
	go func() {
		p.client.Kill()

		// Flush the writer
		_ = p.logger.Flush()

		// Close log file
		_ = p.logFile.Close()

		// Cleanup certificates
		_ = p.ca.CleanupCerts(p.certPath, p.keyPath)
		_ = p.ca.CleanupCerts(p.serverCertPath, p.serverKeyPath)
	}()
}
