package server

import (
	"net"

	"go.uber.org/zap"
)

type options struct {
	listener net.Listener
	logger   *zap.Logger
}

func defaultOptions() *options {
	return &options{
		listener: nil,
		logger:   zap.NewNop(),
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

type loggerOption struct {
	Log *zap.Logger
}

func (o loggerOption) apply(opts *options) {
	opts.logger = o.Log
}

func WithLogger(log *zap.Logger) Option {
	return loggerOption{Log: log}
}
