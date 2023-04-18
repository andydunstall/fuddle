package server

import (
	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"github.com/fuddle-io/fuddle/pkg/registryv2/registry"
)

type ClientWriteServer struct {
	registry *registry.Registry

	rpc.UnimplementedClientWriteRegistry2Server
}

func (s *ClientWriteServer) Sync(stream rpc.ClientWriteRegistry2_SyncServer) error {
	var member *rpc.Member2
	for {
		m, err := stream.Recv()
		if err != nil {
			return nil
		}

		switch m.UpdateType {
		case rpc.ClientMemberUpdateType_REGISTER:
			if m.Member == nil {
				// TODO(AD) log error and ignore
				continue
			}

			member = m.Member
			s.registry.UpsertMember(member)
		case rpc.ClientMemberUpdateType_UNREGISTER:
			if member == nil {
				// TODO(AD) log error and ignore
				continue
			}
			s.registry.MemberLeave(member.State.Id)
		case rpc.ClientMemberUpdateType_HEARTBEAT:
			if member == nil {
				// TODO(AD) log error and ignore
				continue
			}

			s.registry.MemberHeartbeat(member.State.Id)
		}
	}
}
