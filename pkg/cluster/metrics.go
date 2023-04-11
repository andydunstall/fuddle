package cluster

import (
	"github.com/fuddle-io/fuddle/pkg/metrics"
)

type Metrics struct {
	NodesCount *metrics.Gauge
}

func NewMetrics() *Metrics {
	return &Metrics{
		NodesCount: metrics.NewGauge(
			"cluster.nodes.count",
			[]string{},
			"Number of Fuddle nodes in the cluster",
		),
	}
}

func (m *Metrics) Register(collector metrics.Collector) {
	collector.AddGauge(m.NodesCount)
}
