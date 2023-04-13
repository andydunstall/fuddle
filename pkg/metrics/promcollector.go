package metrics

import (
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
)

type PromCollector struct {
	reg *prometheus.Registry
}

func NewPromCollector() *PromCollector {
	reg := prometheus.NewRegistry()
	reg.MustRegister(collectors.NewGoCollector())
	return &PromCollector{
		reg: reg,
	}
}

func (c *PromCollector) AddGauge(g *Gauge) {
	c.reg.MustRegister(g.ToProm())
}

func (c *PromCollector) AddCounter(counter *Counter) {
	c.reg.MustRegister(counter.ToProm())
}

func (c *PromCollector) Registry() *prometheus.Registry {
	return c.reg
}

var _ Collector = &PromCollector{}
