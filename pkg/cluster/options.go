package cluster

import (
	"github.com/fuddle-io/fuddle/pkg/metrics"
	"go.uber.org/zap"
)

type options struct {
	collector metrics.Collector
	logger    *zap.Logger
}

func defaultOptions() *options {
	return &options{
		collector: nil,
		logger:    zap.NewNop(),
	}
}

type Option interface {
	apply(*options)
}

type collectorOption struct {
	collector metrics.Collector
}

func (o collectorOption) apply(opts *options) {
	opts.collector = o.collector
}

func WithCollector(c metrics.Collector) Option {
	return collectorOption{collector: c}
}

type loggerOption struct {
	Log *zap.Logger
}

func (o loggerOption) apply(opts *options) {
	opts.logger = o.Log
}

func WithLogger(log *zap.Logger) Option {
	return loggerOption{Log: log}
}
