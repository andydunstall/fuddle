package registry

import (
	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"go.uber.org/zap"
)

type Server struct {
	registry *Registry

	logger *zap.Logger

	rpc.UnimplementedRegistryServer
}

func NewServer(registry *Registry, opts ...ServerOption) *Server {
	options := defaultServerOptions()
	for _, o := range opts {
		o.apply(options)
	}
	return &Server{
		registry: registry,
		logger:   options.logger,
	}
}

func (s *Server) Subscribe(req *rpc.SubscribeRequest, stream rpc.Registry_SubscribeServer) error {
	s.logger.Debug("subscribe stream")

	unsubscribe := s.registry.Subscribe(req, func(update *rpc.RemoteMemberUpdate) {
		s.logger.Debug(
			"send update",
			zap.String("id", update.Member.Id),
			zap.String("type", update.UpdateType.String()),
		)

		// Ignore return error, if the client closes the stream the context
		// will be cancelled.
		// nolint
		stream.Send(update)
	})
	defer unsubscribe()

	<-stream.Context().Done()
	s.logger.Debug("subscribe stream closed")
	return nil
}

func (s *Server) Register(stream rpc.Registry_RegisterServer) error {
	s.logger.Debug("register stream")

	m, err := stream.Recv()
	if err != nil {
		return nil
	}

	if m.UpdateType != rpc.ClientUpdateType_CLIENT_REGISTER {
		return nil
	}

	member := m.Member
	s.registry.AddMember(member)

	if err := stream.Send(&rpc.ClientAck{
		SeqId: m.SeqId,
	}); err != nil {
		return nil
	}

	for {
		m, err := stream.Recv()
		if err != nil {
			return nil
		}

		if m.UpdateType == rpc.ClientUpdateType_CLIENT_REGISTER {
			member = m.Member
			s.registry.AddMember(member)
		}

		if m.UpdateType == rpc.ClientUpdateType_CLIENT_HEARTBEAT {
			s.registry.AddMember(member)
		}

		if m.UpdateType == rpc.ClientUpdateType_CLIENT_UNREGISTER {
			s.registry.RemoveMember(member.Id)
			return nil
		}

		if err := stream.Send(&rpc.ClientAck{
			SeqId: m.SeqId,
		}); err != nil {
			return nil
		}
	}
}
