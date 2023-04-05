package registry

import (
	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"go.uber.org/zap"
)

type registryOptions struct {
	localMember *rpc.Member
}

func defaultRegistryOptions() *registryOptions {
	return &registryOptions{}
}

type Option interface {
	apply(*registryOptions)
}

type localMemberRegistryOption struct {
	member *rpc.Member
}

func (o localMemberRegistryOption) apply(opts *registryOptions) {
	opts.localMember = o.member
}

func WithRegistryLocalMember(m *rpc.Member) Option {
	return localMemberRegistryOption{member: m}
}

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

type serverOptions struct {
	logger *zap.Logger
}

func defaultServerOptions() *serverOptions {
	return &serverOptions{
		logger: zap.NewNop(),
	}
}

type ServerOption interface {
	apply(*serverOptions)
}

type loggerServerOption struct {
	Log *zap.Logger
}

func (o loggerServerOption) apply(opts *serverOptions) {
	opts.logger = o.Log
}

func WithServerLogger(log *zap.Logger) ServerOption {
	return loggerServerOption{Log: log}
}
