package client

import (
	"go.uber.org/zap"
)

type options struct {
	onConnectionStateChange func(state ConnState)

	logger *zap.Logger
}

func defaultOptions() *options {
	return &options{
		onConnectionStateChange: nil,
		logger:                  zap.NewNop(),
	}
}

type Option interface {
	apply(*options)
}

type onConnectionStateChangeOption struct {
	cb func(state ConnState)
}

func (o onConnectionStateChangeOption) apply(opts *options) {
	opts.onConnectionStateChange = o.cb
}

func WithOnConnectionStateChange(cb func(state ConnState)) Option {
	return &onConnectionStateChangeOption{
		cb: cb,
	}
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
