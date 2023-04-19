package server

import (
	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"github.com/fuddle-io/fuddle/pkg/registryv2/registry"
)

type ClientReadServer struct {
	registry *registry.Registry

	rpc.UnimplementedClientReadRegistry2Server
}

func NewClientReadServer() *ClientReadServer {
	return &ClientReadServer{}
}

func (s *ClientReadServer) Sync(req *rpc.ClientSyncRequest, stream rpc.ClientReadRegistry2_SyncServer) error {
	s.registry.SubscribeFromDigest(req.Digest, req.Filter, func(member *rpc.Member2) {
		// TODO(AD) can't block so queue and send in background
	})
	return nil
}
