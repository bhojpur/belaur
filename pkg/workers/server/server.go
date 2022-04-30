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
	"fmt"
	"net"
	"os"
	"path/filepath"
	"strings"

	belaur "github.com/bhojpur/belaur"
	pb "github.com/bhojpur/belaur/pkg/api/v1/worker"
	"github.com/bhojpur/belaur/pkg/helper/filehelper"
	"github.com/bhojpur/belaur/pkg/security"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials"
)

const (
	hoursBeforeValid = 2
	hoursAfterValid  = 87600 // 10 years
)

// Dependencies defines dependencies of this service.
type Dependencies struct {
	Certificate security.CAAPI
}

// WorkerServer represents an instance of the worker server implementation
type WorkerServer struct {
	Dependencies
}

// InitWorkerServer creates a new worker server instance.
func InitWorkerServer(deps Dependencies) *WorkerServer {
	return &WorkerServer{
		Dependencies: deps,
	}
}

// Start starts the gRPC worker server.
// It returns an error when something badly happens.
func (w *WorkerServer) Start() error {
	lis, err := net.Listen("tcp", ":"+belaur.Cfg.WorkerServerPort)
	if err != nil {
		belaur.Cfg.Logger.Error("cannot start worker gRPC server", "error", err)
		return err
	}

	// Print info message
	belaur.Cfg.Logger.Info("worker gRPC server about to start on port: " + belaur.Cfg.WorkerServerPort)

	// Check if certificates exist for the gRPC server
	certPath := filepath.Join(belaur.Cfg.DataPath, "worker_cert.pem")
	keyPath := filepath.Join(belaur.Cfg.DataPath, "worker_key.pem")
	_, certErr := os.Stat(certPath)
	_, keyErr := os.Stat(keyPath)
	if os.IsNotExist(certErr) || os.IsNotExist(keyErr) {
		// Parse hostname for the certificate
		s := strings.Split(belaur.Cfg.WorkerGRPCHostURL, ":")
		if len(s) != 2 {
			belaur.Cfg.Logger.Error("failed to parse configured gRPC worker host url", "url", belaur.Cfg.WorkerGRPCHostURL)
			return fmt.Errorf("failed to parse configured gRPC worker host url: %s", belaur.Cfg.WorkerGRPCHostURL)
		}

		// Generate certs
		certTmpPath, keyTmpPath, err := w.Certificate.CreateSignedCertWithValidOpts(s[0], hoursBeforeValid, hoursAfterValid)
		if err != nil {
			belaur.Cfg.Logger.Error("failed to generate cert pair for gRPC server", "error", err.Error())
			return err
		}

		// Move certs to correct place
		if err = filehelper.CopyFileContents(certTmpPath, certPath); err != nil {
			belaur.Cfg.Logger.Error("failed to copy gRPC server cert to data folder", "error", err.Error())
			return err
		}
		if err = filehelper.CopyFileContents(keyTmpPath, keyPath); err != nil {
			belaur.Cfg.Logger.Error("failed to copy gRPC server key to data folder", "error", err.Error())
			return err
		}
		if err = os.Remove(certTmpPath); err != nil {
			belaur.Cfg.Logger.Error("failed to remove temporary server cert file", "error", err)
			return err
		}
		if err = os.Remove(keyTmpPath); err != nil {
			belaur.Cfg.Logger.Error("failed to remove temporary key cert file", "error", err)
			return err
		}
	}

	// Generate tls config
	tlsConfig, err := w.Certificate.GenerateTLSConfig(certPath, keyPath)
	if err != nil {
		belaur.Cfg.Logger.Error("failed to generate tls config for gRPC server", "error", err.Error())
		return err
	}

	s := grpc.NewServer(grpc.Creds(credentials.NewTLS(tlsConfig)))
	pb.RegisterWorkerServer(s, &WorkServer{})
	if err := s.Serve(lis); err != nil {
		belaur.Cfg.Logger.Error("cannot start worker gRPC server", "error", err)
		return err
	}
	return nil
}
