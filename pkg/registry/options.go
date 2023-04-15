package registry

import (
	"time"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"github.com/fuddle-io/fuddle/pkg/metrics"
	"go.uber.org/zap"
)

type registryOptions struct {
	localMember      *rpc.Member
	heartbeatTimeout int64
	reconnectTimeout int64
	tombstoneTimeout int64
	now              int64
	collector        metrics.Collector
	logger           *zap.Logger
}

func defaultRegistryOptions() *registryOptions {
	return &registryOptions{
		heartbeatTimeout: 20 * 1000,
		reconnectTimeout: 5 * 60 * 1000,
		tombstoneTimeout: 30 * 60 * 1000,
		now:              time.Now().UnixMilli(),
		collector:        nil,
		logger:           zap.NewNop(),
	}
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

type registryHeartbeatTimeoutOption struct {
	timeout int64
}

func (o registryHeartbeatTimeoutOption) apply(opts *registryOptions) {
	opts.heartbeatTimeout = o.timeout
}

func WithHeartbeatTimeout(timeout int64) Option {
	return registryHeartbeatTimeoutOption{timeout: timeout}
}

type registryReconnectTimeoutOption struct {
	timeout int64
}

func (o registryReconnectTimeoutOption) apply(opts *registryOptions) {
	opts.reconnectTimeout = o.timeout
}

func WithReconnectTimeout(timeout int64) Option {
	return registryReconnectTimeoutOption{timeout: timeout}
}

type registryTombstoneTimeoutOption struct {
	timeout int64
}

func (o registryTombstoneTimeoutOption) apply(opts *registryOptions) {
	opts.tombstoneTimeout = o.timeout
}

func WithTombstoneTimeout(timeout int64) Option {
	return registryTombstoneTimeoutOption{timeout: timeout}
}

type registryNowTimeOption struct {
	now int64
}

func (o registryNowTimeOption) apply(opts *registryOptions) {
	opts.now = o.now
}

// WithRegistryNowTime sets the time 'now' to the given timestamp. This can be
// useful for testing.
func WithRegistryNowTime(now int64) Option {
	return registryNowTimeOption{now: now}
}

type collectorOption struct {
	collector metrics.Collector
}

func (o collectorOption) apply(opts *registryOptions) {
	opts.collector = o.collector
}

func WithCollector(c metrics.Collector) Option {
	return collectorOption{collector: c}
}

type loggerRegistryOption struct {
	Log *zap.Logger
}

func (o loggerRegistryOption) apply(opts *registryOptions) {
	opts.logger = o.Log
}

func WithRegistryLogger(log *zap.Logger) Option {
	return loggerRegistryOption{Log: log}
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
