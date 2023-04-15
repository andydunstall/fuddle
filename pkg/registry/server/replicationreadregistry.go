package server

import (
	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"github.com/fuddle-io/fuddle/pkg/registry"
	"go.uber.org/zap"
)

// ReplicaReadRegistryServer services updates to the registry to other Fuddle
// nodes in the cluster.
type ReplicaReadRegistryServer struct {
	registry *registry.Registry

	logger *zap.Logger

	rpc.UnimplementedReplicaReadRegistryServer
}

func NewReplicaReadRegistryServer(reg *registry.Registry, opts ...Option) *ReplicaReadRegistryServer {
	options := defaultOptions()
	for _, o := range opts {
		o.apply(options)
	}

	return &ReplicaReadRegistryServer{
		registry: reg,
		logger:   options.logger,
	}
}

func (s *ReplicaReadRegistryServer) Updates(req *rpc.SubscribeRequest, stream rpc.ReplicaReadRegistry_UpdatesServer) error {
	logger := s.logger.With(zap.String("rpc", "ReplicaReadRegistryServer.Updates"))
	logger.Debug("updates stream")

	unsubscribe := s.registry.Subscribe(req, func(update *rpc.RemoteMemberUpdate) {
		logger.Debug(
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
	logger.Debug("subscribe stream closed")

	return nil
}
