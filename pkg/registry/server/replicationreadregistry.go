package server

import (
	"context"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"github.com/fuddle-io/fuddle/pkg/metrics"
	"github.com/fuddle-io/fuddle/pkg/registry/registry"
	"go.uber.org/zap"
)

// ReplicaReadRegistryServer services updates to the registry to other Fuddle
// nodes in the cluster.
type ReplicaReadRegistryServer struct {
	registry *registry.Registry

	outboundUpdates *metrics.Counter
	logger          *zap.Logger

	rpc.UnimplementedReplicaReadRegistryServer
}

func NewReplicaReadRegistryServer(reg *registry.Registry, opts ...Option) *ReplicaReadRegistryServer {
	options := defaultOptions()
	for _, o := range opts {
		o.apply(options)
	}

	outboundUpdates := metrics.NewCounter(
		"registry",
		"updates.replica.outbound",
		[]string{},
		"Number of outbound updates to a replica node",
	)
	if options.collector != nil {
		options.collector.AddCounter(outboundUpdates)
	}

	return &ReplicaReadRegistryServer{
		registry:        reg,
		outboundUpdates: outboundUpdates,
		logger:          options.logger,
	}
}

func (s *ReplicaReadRegistryServer) Updates(req *rpc.SubscribeRequest, stream rpc.ReplicaReadRegistry_UpdatesServer) error {
	logger := s.logger.With(zap.String("rpc", "ReplicaReadRegistryServer.Updates"))
	logger.Debug("updates stream")

	unsubscribe := s.registry.Subscribe(req, func(update *rpc.Member2) {
		logger.Debug(
			"send update",
			zap.String("id", update.State.Id),
		)

		s.outboundUpdates.Inc(map[string]string{})

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

func (s *ReplicaReadRegistryServer) Update(ctx context.Context, req *rpc.UpdateRequest) (*rpc.UpdateResponse, error) {
	s.logger.Debug(
		"replica update",
		zap.String("id", req.Member.State.Id),
	)

	s.registry.RemoteUpdate(req.Member)
	return &rpc.UpdateResponse{}, nil
}
