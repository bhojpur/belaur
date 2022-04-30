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
	"context"

	proto "github.com/bhojpur/belaur/pkg/api/v1/plugin"
	plugin "github.com/hashicorp/go-plugin"
	"google.golang.org/grpc"
)

// BelaurPlugin is the Bhojpur Belaur plugin interface used for communication
// with the plugin.
type BelaurPlugin interface {
	GetJobs() (proto.Plugin_GetJobsClient, error)
	ExecuteJob(job *proto.Job) (*proto.JobResult, error)
}

// BelaurPluginClient represents gRPC client
type BelaurPluginClient struct {
	client proto.PluginClient
}

// BelaurPluginImpl represents the plugin implementation on client side.
type BelaurPluginImpl struct {
	Impl BelaurPlugin

	plugin.NetRPCUnsupportedPlugin
}

// GRPCServer is needed here to implement hashicorp
// plugin.Plugin interface. Real implementation is
// in the plugin(s).
func (p *BelaurPluginImpl) GRPCServer(b *plugin.GRPCBroker, s *grpc.Server) error {
	// Real implementation defined in plugin
	return nil
}

// GRPCClient is the passing method for the gRPC client.
func (p *BelaurPluginImpl) GRPCClient(context context.Context, b *plugin.GRPCBroker, c *grpc.ClientConn) (interface{}, error) {
	return &BelaurPluginClient{client: proto.NewPluginClient(c)}, nil
}

// GetJobs requests all jobs from the plugin.
// We get a stream of proto.Job back.
func (m *BelaurPluginClient) GetJobs() (proto.Plugin_GetJobsClient, error) {
	return m.client.GetJobs(context.Background(), &proto.Empty{})
}

// ExecuteJob triggers the execution of the given job in the plugin.
func (m *BelaurPluginClient) ExecuteJob(job *proto.Job) (*proto.JobResult, error) {
	return m.client.ExecuteJob(context.Background(), job)
}
