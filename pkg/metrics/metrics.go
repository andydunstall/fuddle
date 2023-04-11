package metrics

import (
	"sort"
	"strings"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
)

type labelledValue struct {
	Label string
	Value string
}

func (lv *labelledValue) String() string {
	return lv.Label + "=" + lv.Value
}

type Gauge struct {
	values map[string]float64

	// mu is a mutex protecting the fields above.
	mu sync.Mutex

	promGauge *prometheus.GaugeVec
}

func NewGauge(name string, labels []string, help string) *Gauge {
	return &Gauge{
		values: make(map[string]float64),
		promGauge: prometheus.NewGaugeVec(
			prometheus.GaugeOpts{
				Name: name,
				Help: help,
			},
			labels,
		),
	}
}

func (g *Gauge) Set(v float64, labels map[string]string) {
	g.mu.Lock()
	g.values[labelsToString(labels)] = v
	g.mu.Unlock()

	g.promGauge.With(prometheus.Labels(labels)).Set(v)
}

func (g *Gauge) Value(labels map[string]string) float64 {
	return g.values[labelsToString(labels)]
}

func (g *Gauge) ToProm() *prometheus.GaugeVec {
	return g.promGauge
}

func labelsToString(labels map[string]string) string {
	var labelledValues []labelledValue
	for l, v := range labels {
		labelledValues = append(labelledValues, labelledValue{
			Label: l,
			Value: v,
		})
	}
	sort.Slice(labelledValues, func(i, j int) bool {
		return labelledValues[i].Label < labelledValues[j].Label
	})
	var strs []string
	for _, lv := range labelledValues {
		strs = append(strs, lv.String())
	}
	return strings.Join(strs, ",")
}
