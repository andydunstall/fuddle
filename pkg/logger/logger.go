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
	if options.path != "" {
		conf.OutputPaths = []string{options.path}
	}

	metrics := NewMetrics()
	if options.collector != nil {
		metrics.Register(options.collector)
	}

	logger, err := conf.Build(zap.WrapCore(func(core zapcore.Core) zapcore.Core {
		return newMetricsCore(metrics, core)
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

func StringToLevel(lvl string) zapcore.Level {
	switch lvl {
	case "debug":
		return zapcore.DebugLevel
	case "info":
		return zapcore.InfoLevel
	case "warn":
		return zapcore.WarnLevel
	case "error":
		return zapcore.ErrorLevel
	default:
		// If the level is invalid or not specified, default to info.
		return zapcore.InfoLevel
	}
}
