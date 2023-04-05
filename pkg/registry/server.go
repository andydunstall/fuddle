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
			zap.String("id", update.Id),
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
