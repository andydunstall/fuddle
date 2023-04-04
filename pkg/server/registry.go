package server

import (
	rpc "github.com/fuddle-io/fuddle-rpc/go"
)

type registryServer struct {
	rpc.UnimplementedRegistryServer
}

func newRegistryServer() *registryServer {
	return &registryServer{}
}

func (s *registryServer) Subscribe(req *rpc.SubscribeRequest, stream rpc.Registry_SubscribeServer) error {
	<-stream.Context().Done()
	return nil
}
