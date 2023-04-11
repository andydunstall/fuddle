package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
)

type PromCollector struct {
	reg *prometheus.Registry
}

func NewPromCollector() *PromCollector {
	return &PromCollector{
		reg: prometheus.NewRegistry(),
	}
}

func (c *PromCollector) AddGauge(g *Gauge) {
	c.reg.MustRegister(g.ToProm())
}

func (c *PromCollector) Registry() *prometheus.Registry {
	return c.reg
}

var _ Collector = &PromCollector{}
