package server

import (
	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"github.com/fuddle-io/fuddle/pkg/metrics"
	"github.com/fuddle-io/fuddle/pkg/registry"
	"go.uber.org/zap"
)

// ClientWriteRegistryServer receives updates from external clients.
type ClientWriteRegistryServer struct {
	registry *registry.Registry

	inboundUpdates *metrics.Counter
	logger         *zap.Logger

	rpc.UnimplementedClientWriteRegistryServer
}

func NewClientWriteRegistryServer(reg *registry.Registry, opts ...Option) *ClientWriteRegistryServer {
	options := defaultOptions()
	for _, o := range opts {
		o.apply(options)
	}

	inboundUpdates := metrics.NewCounter(
		"registry",
		"updates.client.inbound",
		[]string{"updatetype"},
		"Number of inbound updates from the client",
	)
	if options.collector != nil {
		options.collector.AddCounter(inboundUpdates)
	}

	return &ClientWriteRegistryServer{
		registry:       reg,
		inboundUpdates: inboundUpdates,
		logger:         options.logger,
	}
}

func (s *ClientWriteRegistryServer) Register(stream rpc.ClientWriteRegistry_RegisterServer) error {
	logger := s.logger.With(zap.String("rpc", "ClientWriteRegistryServer.Register"))
	logger.Debug("register stream")

	m, err := stream.Recv()
	if err != nil {
		return nil
	}

	if m.UpdateType != rpc.ClientUpdateType_CLIENT_REGISTER {
		return nil
	}

	s.inboundUpdates.Inc(map[string]string{
		"updatetype": clientUpdateTypeToString(m.UpdateType),
	})

	member := m.Member
	s.registry.AddMember(member)

	for {
		m, err := stream.Recv()
		if err != nil {
			return nil
		}

		s.inboundUpdates.Inc(map[string]string{
			"updatetype": clientUpdateTypeToString(m.UpdateType),
		})

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
	}
}
