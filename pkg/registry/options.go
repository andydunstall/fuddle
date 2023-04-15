package registry

import (
	"time"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"github.com/fuddle-io/fuddle/pkg/metrics"
	"go.uber.org/zap"
)

type options struct {
	localMember      *rpc.Member
	heartbeatTimeout int64
	reconnectTimeout int64
	tombstoneTimeout int64
	now              int64
	collector        metrics.Collector
	logger           *zap.Logger
}

func defaultOptions() *options {
	return &options{
		heartbeatTimeout: 20 * 1000,
		reconnectTimeout: 5 * 60 * 1000,
		tombstoneTimeout: 30 * 60 * 1000,
		now:              time.Now().UnixMilli(),
		collector:        nil,
		logger:           zap.NewNop(),
	}
}

type Option interface {
	apply(*options)
}

type localMemberOption struct {
	member *rpc.Member
}

func (o localMemberOption) apply(opts *options) {
	opts.localMember = o.member
}

func WithLocalMember(m *rpc.Member) Option {
	return localMemberOption{member: m}
}

type heartbeatTimeoutOption struct {
	timeout int64
}

func (o heartbeatTimeoutOption) apply(opts *options) {
	opts.heartbeatTimeout = o.timeout
}

func WithHeartbeatTimeout(timeout int64) Option {
	return heartbeatTimeoutOption{timeout: timeout}
}

type reconnectTimeoutOption struct {
	timeout int64
}

func (o reconnectTimeoutOption) apply(opts *options) {
	opts.reconnectTimeout = o.timeout
}

func WithReconnectTimeout(timeout int64) Option {
	return reconnectTimeoutOption{timeout: timeout}
}

type tombstoneTimeoutOption struct {
	timeout int64
}

func (o tombstoneTimeoutOption) apply(opts *options) {
	opts.tombstoneTimeout = o.timeout
}

func WithTombstoneTimeout(timeout int64) Option {
	return tombstoneTimeoutOption{timeout: timeout}
}

type nowTimeOption struct {
	now int64
}

func (o nowTimeOption) apply(opts *options) {
	opts.now = o.now
}

// WithNowTime sets the time 'now' to the given timestamp. This can be
// useful for testing.
func WithNowTime(now int64) Option {
	return nowTimeOption{now: now}
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
