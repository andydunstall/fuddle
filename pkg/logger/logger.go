package logger

import (
	"fmt"

	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type Logger struct {
	baseLogger *zap.Logger
	metrics    *Metrics
}

func NewLogger(opts ...Option) (*Logger, error) {
	options := defaultOptions()
	for _, o := range opts {
		o.apply(options)
	}

	conf := zap.NewProductionConfig()
	conf.Level.SetLevel(options.level)

	metrics := NewMetrics()
	if options.collector != nil {
		metrics.Register(options.collector)
	}

	logger, err := conf.Build(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return newMetricsCore(metrics)
	}))
	if err != nil {
		return nil, fmt.Errorf("logger: %w", err)
	}

	return &Logger{
		baseLogger: logger,
		metrics:    metrics,
	}, nil
}

func (l *Logger) Metrics() *Metrics {
	return l.metrics
}

func (l *Logger) Logger(subsystem string) *zap.Logger {
	return l.baseLogger.With(zap.String("subsystem", subsystem))
}
