package server

import (
	"context"
	"time"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"github.com/fuddle-io/fuddle/pkg/registryv2/registry"
)

type ClientWriteServer struct {
	registry *registry.Registry

	rpc.UnimplementedClientWriteRegistry2Server
}

func NewClientWriteServer() *ClientWriteServer {
	return &ClientWriteServer{}
}

func (s *ClientWriteServer) MemberJoin(ctx context.Context, req *rpc.ClientMemberJoinRequest) (*rpc.ClientMemberJoinResponse, error) {
	s.registry.OwnedMemberUpsert(req.Member, time.Now().UnixMilli())
	return &rpc.ClientMemberJoinResponse{}, nil
}

func (s *ClientWriteServer) MemberLeave(ctx context.Context, req *rpc.ClientMemberLeaveRequest) (*rpc.ClientMemberLeaveResponse, error) {
	s.registry.OwnedMemberLeave(req.MemberId, time.Now().UnixMilli())
	return &rpc.ClientMemberLeaveResponse{}, nil
}

func (s *ClientWriteServer) MemberHeartbeat(ctx context.Context, req *rpc.ClientMemberHeartbeatRequest) (*rpc.ClientMemberHeartbeatResponse, error) {
	s.registry.OwnedMemberHeartbeat(req.MemberId, time.Now().UnixMilli())
	return nil, nil
}
