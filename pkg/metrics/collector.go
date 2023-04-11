package metrics

type Collector interface {
	AddGauge(g *Gauge)
}
