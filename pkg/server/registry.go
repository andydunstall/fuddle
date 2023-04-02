package server

import (
	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"github.com/fuddle-io/fuddle/pkg/registry"
)

type registryServer struct {
	registry *registry.Registry

	rpc.UnimplementedRegistryServer
}

func newRegistryServer(registry *registry.Registry) *registryServer {
	return &registryServer{
		registry: registry,
	}
}

func (s *registryServer) Subscribe(req *rpc.SubscribeRequest, stream rpc.Registry_SubscribeServer) error {
	unsub := s.registry.Subscribe(req.OwnerOnly, func(id string) {
		// nolint
		stream.Send(&rpc.MemberUpdate{
			Id:         id,
			UpdateType: rpc.MemberUpdateType_REGISTER,
		})
	})
	defer unsub()

	<-stream.Context().Done()

	return nil
}
