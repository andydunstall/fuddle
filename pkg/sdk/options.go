package fuddle

import (
	"time"

	"go.uber.org/zap"
)

type options struct {
	connectAttemptTimeout time.Duration
	keepAlivePingInterval time.Duration
	keepAlivePingTimeout  time.Duration

	onConnectionStateChange func(state ConnState)

	logger              *zap.Logger
	grpcLoggerVerbosity int
}

func defaultOptions() *options {
	return &options{
		connectAttemptTimeout:   time.Second * 4,
		keepAlivePingInterval:   time.Second * 10,
		keepAlivePingTimeout:    time.Second * 4,
		onConnectionStateChange: nil,
		logger:                  zap.NewNop(),
		grpcLoggerVerbosity:     0,
	}
}

type Option interface {
	apply(*options)
}

type connectAttemptTimeoutOption struct {
	timeout time.Duration
}

func (o connectAttemptTimeoutOption) apply(opts *options) {
	opts.connectAttemptTimeout = o.timeout
}

// WithConnectAttemptTimeout is the timeout for each connect attempt.
// This is different from the overall connect timeout which may attempt multiple
// addresses.
//
// Defaults to 4 seconds.
func WithConnectAttemptTimeout(timeout time.Duration) Option {
	return connectAttemptTimeoutOption{timeout: timeout}
}

type keepAlivePingIntervalOption struct {
	interval time.Duration
}

func (o keepAlivePingIntervalOption) apply(opts *options) {
	opts.keepAlivePingInterval = o.interval
}

// WithKeepAlivePingInterval is the interval to send keepAlive pings to the
// connected Fuddle node, which is used to detect an unresponsive connection
// and trigger a reconnection attempt.
//
// Defaults to 10 seconds.
func WithKeepAlivePingInterval(interval time.Duration) Option {
	return keepAlivePingIntervalOption{interval: interval}
}

type keepAlivePingTimeoutOption struct {
	timeout time.Duration
}

func (o keepAlivePingTimeoutOption) apply(opts *options) {
	opts.keepAlivePingTimeout = o.timeout
}

// WithKeepAlivePingTimeout is the time to wait for a keepalive ping response
// before timing out and considering the connection as failed.
//
// Defaults to 4 seconds.
func WithKeepAlivePingTimeout(timeout time.Duration) Option {
	return keepAlivePingTimeoutOption{timeout: timeout}
}

type onConnectionStateChangeOption struct {
	cb func(state ConnState)
}

func (o onConnectionStateChangeOption) apply(opts *options) {
	opts.onConnectionStateChange = o.cb
}

// WithOnConnectionStateChange adds an optional callback to receive updates when
// the clients connection state changes.
func WithOnConnectionStateChange(cb func(state ConnState)) Option {
	return &onConnectionStateChangeOption{
		cb: cb,
	}
}

type loggerOption struct {
	logger *zap.Logger
}

func (o loggerOption) apply(opts *options) {
	opts.logger = o.logger
}

func WithLogger(logger *zap.Logger) Option {
	return loggerOption{logger: logger}
}

type grpcLoggerVerbosityOption struct {
	v int
}

func (o grpcLoggerVerbosityOption) apply(opts *options) {
	opts.grpcLoggerVerbosity = o.v
}

// WithGRPCLoggerVerbosity adds gRPC logging to stdout and stderr. Note this is
// independent from WithLogger and should only be used for debugging rather
// than production code.
//
// Defaults to 0 to disable gRPC logs.
func WithGRPCLoggerVerbosity(v int) Option {
	return grpcLoggerVerbosityOption{v: v}
}
