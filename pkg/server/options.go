package server

import (
	"net"

	"go.uber.org/zap"
)

type options struct {
	listener *net.TCPListener
	logger   *zap.Logger
}

func defaultOptions() options {
	return options{
		logger: zap.NewNop(),
	}
}

type Option interface {
	apply(*options)
}

type listenerOption struct {
	ln *net.TCPListener
}

func (o listenerOption) apply(opts *options) {
	opts.listener = o.ln
}

func WithListener(ln *net.TCPListener) Option {
	return listenerOption{ln: ln}
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
