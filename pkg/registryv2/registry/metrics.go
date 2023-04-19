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
			[]string{"liveness", "service", "owner"},
			"Number of members in the registry",
		),
		MembersOwned: metrics.NewGauge(
			"registry",
			"members.owned",
			[]string{"liveness", "service"},
			"Number of members owned by the node",
		),
	}
}

func (m *Metrics) Register(collector metrics.Collector) {
	collector.AddGauge(m.MembersCount)
	collector.AddGauge(m.MembersOwned)
}
