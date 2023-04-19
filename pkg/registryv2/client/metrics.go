package client

import (
	"github.com/fuddle-io/fuddle/pkg/metrics"
)

type Metrics struct {
	ReplicaUpdatesOutbound *metrics.Counter

	RepairUpdatesInbound *metrics.Counter
}

func NewMetrics() *Metrics {
	return &Metrics{
		ReplicaUpdatesOutbound: metrics.NewCounter(
			"registry",
			"replica.updates.outbound",
			[]string{"target"},
			"Number of outbound updates sent to replicas",
		),

		RepairUpdatesInbound: metrics.NewCounter(
			"registry",
			"repair.updates.inbound",
			[]string{"source"},
			"Number of inbound updates from replica repair",
		),
	}
}

func (m *Metrics) Register(collector metrics.Collector) {
	collector.AddCounter(m.ReplicaUpdatesOutbound)
	collector.AddCounter(m.RepairUpdatesInbound)
}
