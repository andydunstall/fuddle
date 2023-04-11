package registry

import (
	"github.com/fuddle-io/fuddle/pkg/metrics"
)

type Metrics struct {
	MembersCount *metrics.Gauge
}

func NewMetrics() *Metrics {
	return &Metrics{
		MembersCount: metrics.NewGauge(
			"registry",
			"members.count",
			[]string{"status"},
			"Number of registered members in the cluster",
		),
	}
}

func (m *Metrics) Register(collector metrics.Collector) {
	collector.AddGauge(m.MembersCount)
}
