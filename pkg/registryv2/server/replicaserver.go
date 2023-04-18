package server

import (
	"context"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"github.com/fuddle-io/fuddle/pkg/registryv2/registry"
)

type ReplicaServer struct {
	registry *registry.Registry

	rpc.UnimplementedReplicaRegistry2Server
}

func (s *ReplicaServer) Update(ctx context.Context, req *rpc.UpdateRequest) (*rpc.UpdateResponse, error) {
	s.registry.UpsertMember(req.Member)
	return &rpc.UpdateResponse{}, nil
}

func (s *ReplicaServer) Sync(ctx context.Context, req *rpc.ReplicaSyncRequest) (*rpc.ReplicaSyncResponse, error) {
	return &rpc.ReplicaSyncResponse{
		Members: s.registry.MembersDelta(req.Digest),
	}, nil
}
