package metrics

type Collector interface {
	AddGauge(g *Gauge)
	AddCounter(c *Counter)
}
