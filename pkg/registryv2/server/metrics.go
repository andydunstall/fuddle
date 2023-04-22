package server

import (
	"github.com/fuddle-io/fuddle/pkg/metrics"
)

type Metrics struct {
	ClientUpdatesInbound  *metrics.Counter
	ClientUpdatesOutbound *metrics.Counter

	ReplicaUpdatesInbound *metrics.Counter

	RepairUpdatesOutbound *metrics.Counter
}

func NewMetrics() *Metrics {
	return &Metrics{
		ClientUpdatesInbound: metrics.NewCounter(
			"registry",
			"client.updates.inbound",
			[]string{},
			"Number of inbound updates received from a client",
		),
		ClientUpdatesOutbound: metrics.NewCounter(
			"registry",
			"client.updates.outbound",
			[]string{},
			"Number of outbound updates sent to a client",
		),

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

func (m *Metrics) Register(collector metrics.Collector) {
	collector.AddCounter(m.ClientUpdatesInbound)
	collector.AddCounter(m.ClientUpdatesOutbound)
	collector.AddCounter(m.ReplicaUpdatesInbound)
	collector.AddCounter(m.RepairUpdatesOutbound)
}
