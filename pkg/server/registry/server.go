package registry

import (
	rpc "github.com/fuddle-io/fuddle-rpc/go"
)

type Server struct {
	rpc.UnimplementedRegistryServer
}

func NewServer(opts ...ServerOption) *Server {
	options := defaultServerOptions()
	for _, o := range opts {
		o.apply(options)
	}
	return &Server{}
}

func (s *Server) Subscribe(req *rpc.SubscribeRequest, stream rpc.Registry_SubscribeServer) error {
	<-stream.Context().Done()
	return nil
}
