package logger

import (
	"github.com/fuddle-io/fuddle/pkg/metrics"
	"go.uber.org/zap/zapcore"
)

type Metrics struct {
	WarningsCount *metrics.Counter
	ErrorsCount   *metrics.Counter
}

func NewMetrics() *Metrics {
	return &Metrics{
		WarningsCount: metrics.NewCounter(
			"",
			"warnings",
			[]string{"subsystem"},
			"Number of warnings in the system",
		),
		ErrorsCount: metrics.NewCounter(
			"",
			"errors",
			[]string{"subsystem"},
			"Number of errors in the system",
		),
	}
}

func (m *Metrics) Register(collector metrics.Collector) {
	collector.AddCounter(m.WarningsCount)
	collector.AddCounter(m.ErrorsCount)
}

type metricsCore struct {
	metrics   *Metrics
	subsystem string

	core zapcore.Core
}

func newMetricsCore(metrics *Metrics, core zapcore.Core) zapcore.Core {
	return &metricsCore{
		metrics: metrics,
		core:    core,
	}
}

func (c *metricsCore) With(fields []zapcore.Field) zapcore.Core {
	subsystem := ""
	for _, field := range fields {
		if field.Key == "subsystem" && field.String != "" {
			subsystem = field.String
		}
	}

	return &metricsCore{
		metrics:   c.metrics,
		subsystem: subsystem,
		core:      c.core.With(fields),
	}
}

func (c *metricsCore) Check(entry zapcore.Entry, ce *zapcore.CheckedEntry) *zapcore.CheckedEntry {
	if entry.Level >= zapcore.WarnLevel {
		ce = ce.AddCore(entry, c)
	}
	return c.core.Check(entry, ce)
}

func (c *metricsCore) Write(entry zapcore.Entry, fields []zapcore.Field) error {
	subsystem := c.subsystem
	if subsystem == "" {
		subsystem = "unknown"
	}

	if entry.Level == zapcore.WarnLevel {
		c.metrics.WarningsCount.Inc(map[string]string{
			"subsystem": subsystem,
		})
	}

	if entry.Level == zapcore.ErrorLevel {
		c.metrics.ErrorsCount.Inc(map[string]string{
			"subsystem": subsystem,
		})
	}

	return c.core.Write(entry, fields)
}

func (c *metricsCore) Enabled(lvl zapcore.Level) bool {
	if lvl >= zapcore.WarnLevel {
		return true
	}
	return c.core.Enabled(lvl)
}

func (c *metricsCore) Sync() error {
	return c.core.Sync()
}
