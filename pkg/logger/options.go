package logger

import (
	"github.com/fuddle-io/fuddle/pkg/metrics"
	"go.uber.org/zap/zapcore"
)

type options struct {
	level     zapcore.Level
	path      string
	collector metrics.Collector
}

func defaultOptions() *options {
	return &options{
		level:     zapcore.InfoLevel,
		path:      "",
		collector: nil,
	}
}

type Option interface {
	apply(*options)
}

type levelOption struct {
	level zapcore.Level
}

func (o levelOption) apply(opts *options) {
	opts.level = o.level
}

func WithLevel(l zapcore.Level) Option {
	return levelOption{level: l}
}

type pathOption struct {
	path string
}

func (o pathOption) apply(opts *options) {
	opts.path = o.path
}

func WithPath(path string) Option {
	return pathOption{path: path}
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
