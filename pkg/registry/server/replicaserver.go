package server

import (
	"context"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"github.com/fuddle-io/fuddle/pkg/metrics"
	"github.com/fuddle-io/fuddle/pkg/registry/registry"
	"go.uber.org/zap"
)

type ReplicaServerMetrics struct {
	ReplicaUpdatesInbound *metrics.Counter
	RepairUpdatesOutbound *metrics.Counter
}

func NewReplicaServerMetrics() *ReplicaServerMetrics {
	return &ReplicaServerMetrics{
		ReplicaUpdatesInbound: metrics.NewCounter(
			"registry",
			"replica.updates.inbound",
			[]string{"source"},
			"Number of inbound updates received from replicas",
		),

		RepairUpdatesOutbound: metrics.NewCounter(
			"registry",
			"repair.updates.outbound",
			[]string{"target"},
			"Number of outbound updates from replica repair",
		),
	}
}

func (m *ReplicaServerMetrics) Register(collector metrics.Collector) {
	collector.AddCounter(m.ReplicaUpdatesInbound)
	collector.AddCounter(m.RepairUpdatesOutbound)
}

type ReplicaServer struct {
	registry *registry.Registry

	metrics *ReplicaServerMetrics
	logger  *zap.Logger

	rpc.UnimplementedReplicaRegistry2Server
}

func NewReplicaServer(reg *registry.Registry, opts ...Option) *ReplicaServer {
	options := defaultOptions()
	for _, o := range opts {
		o.apply(options)
	}

	metrics := NewReplicaServerMetrics()
	if options.collector != nil {
		metrics.Register(options.collector)
	}

	return &ReplicaServer{
		registry: reg,
		metrics:  metrics,
		logger:   options.logger,
	}
}

func (s *ReplicaServer) Metrics() *ReplicaServerMetrics {
	return s.metrics
}

func (s *ReplicaServer) Update(ctx context.Context, req *rpc.UpdateRequest) (*rpc.UpdateResponse, error) {
	s.metrics.ReplicaUpdatesInbound.Inc(map[string]string{
		"source": req.SourceNodeId,
	})

	s.registry.RemoteUpdate(req.Member)
	return &rpc.UpdateResponse{}, nil
}

func (s *ReplicaServer) Sync(ctx context.Context, req *rpc.ReplicaSyncRequest) (*rpc.ReplicaSyncResponse, error) {
	delta := s.registry.Delta(req.Digest)

	s.metrics.RepairUpdatesOutbound.Add(len(delta), map[string]string{
		"target": req.SourceNodeId,
	})

	return &rpc.ReplicaSyncResponse{
		Members: delta,
	}, nil
}
