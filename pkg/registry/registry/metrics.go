package registry

import (
	"github.com/fuddle-io/fuddle/pkg/metrics"
)

type Metrics struct {
	MembersCount *metrics.Gauge
	MembersOwned *metrics.Gauge
}

func NewMetrics() *Metrics {
	return &Metrics{
		MembersCount: metrics.NewGauge(
			"registry",
			"members.count",
			[]string{"status", "owner"},
			"Number of registered members in the cluster",
		),
		MembersOwned: metrics.NewGauge(
			"registry",
			"members.owned",
			[]string{"status"},
			"Number of members owned by this node",
		),
	}
}

func (m *Metrics) Register(collector metrics.Collector) {
	collector.AddGauge(m.MembersCount)
	collector.AddGauge(m.MembersOwned)
}
