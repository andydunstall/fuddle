package registry

import (
	"go.uber.org/zap"
)

type clientOptions struct {
	onConnectionStateChange func(state ConnState)

	logger *zap.Logger
}

func defaultClientOptions() *clientOptions {
	return &clientOptions{
		onConnectionStateChange: nil,
		logger:                  zap.NewNop(),
	}
}

type ClientOption interface {
	apply(*clientOptions)
}

type onConnectionStateChangeClientOption struct {
	cb func(state ConnState)
}

func (o onConnectionStateChangeClientOption) apply(opts *clientOptions) {
	opts.onConnectionStateChange = o.cb
}

func WithOnClientConnectionStateChange(cb func(state ConnState)) ClientOption {
	return &onConnectionStateChangeClientOption{
		cb: cb,
	}
}

type loggerClientOption struct {
	Log *zap.Logger
}

func (o loggerClientOption) apply(opts *clientOptions) {
	opts.logger = o.Log
}

func WithClientLogger(log *zap.Logger) ClientOption {
	return loggerClientOption{Log: log}
}

type serverServerOptions struct {
	logger *zap.Logger
}

func defaultServerOptions() *serverServerOptions {
	return &serverServerOptions{
		logger: zap.NewNop(),
	}
}

type ServerOption interface {
	apply(*serverServerOptions)
}

type loggerServerOption struct {
	Log *zap.Logger
}

func (o loggerServerOption) apply(opts *serverServerOptions) {
	opts.logger = o.Log
}

func WithServerLogger(log *zap.Logger) ServerOption {
	return loggerServerOption{Log: log}
}
