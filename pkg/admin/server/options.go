package server

import (
	"net"

	"github.com/fuddle-io/fuddle/pkg/metrics"
	"go.uber.org/zap"
)

type options struct {
	listener  net.Listener
	collector *metrics.PromCollector
	logger    *zap.Logger
}

func defaultOptions() *options {
	return &options{
		listener:  nil,
		collector: nil,
		logger:    zap.NewNop(),
	}
}

type Option interface {
	apply(*options)
}

type listenerOption struct {
	listener net.Listener
}

func (o listenerOption) apply(opts *options) {
	opts.listener = o.listener
}

func WithListener(ln net.Listener) Option {
	return listenerOption{listener: ln}
}

type collectorOption struct {
	collector *metrics.PromCollector
}

func (o collectorOption) apply(opts *options) {
	opts.collector = o.collector
}

func WithCollector(c *metrics.PromCollector) Option {
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
