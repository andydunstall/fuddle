package server

import (
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

func (s *ClientWriteServer) Sync(stream rpc.ClientWriteRegistry2_SyncServer) error {
	var member *rpc.MemberState
	for {
		m, err := stream.Recv()
		if err != nil {
			return nil
		}

		switch m.UpdateType {
		case rpc.ClientMemberUpdateType_JOIN:
			if m.State == nil {
				// TODO(AD) log error and ignore
				continue
			}

			member = m.State
			s.registry.OwnedMemberAdd(member)
		case rpc.ClientMemberUpdateType_LEAVE:
			if member == nil {
				// TODO(AD) log error and ignore
				continue
			}
			s.registry.OwnedMemberLeave(member.Id)
		case rpc.ClientMemberUpdateType_HEARTBEAT:
			if member == nil {
				// TODO(AD) log error and ignore
				continue
			}

			s.registry.OwnedMemberHeartbeat(member.Id)
		}
	}
}
