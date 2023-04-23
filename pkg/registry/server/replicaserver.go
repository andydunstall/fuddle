package server

import (
	"context"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"github.com/fuddle-io/fuddle/pkg/registry/registry"
	"go.uber.org/zap"
)

type ReplicaServer struct {
	registry *registry.Registry

	logger *zap.Logger

	rpc.UnimplementedReplicaRegistry2Server
}

func NewReplicaServer(reg *registry.Registry, opts ...Option) *ReplicaServer {
	options := defaultOptions()
	for _, o := range opts {
		o.apply(options)
	}

	return &ReplicaServer{
		registry: reg,
		logger:   options.logger,
	}
}

func (s *ReplicaServer) Update(ctx context.Context, req *rpc.UpdateRequest) (*rpc.UpdateResponse, error) {
	s.registry.RemoteUpdate(req.Member)
	return &rpc.UpdateResponse{}, nil
}

func (s *ReplicaServer) Sync(ctx context.Context, req *rpc.ReplicaSyncRequest) (*rpc.ReplicaSyncResponse, error) {
	return &rpc.ReplicaSyncResponse{}, nil
}
