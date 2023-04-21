package registry

import (
	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"github.com/fuddle-io/fuddle/pkg/metrics"
)

type options struct {
	localMember *rpc.MemberState

	tombstoneTimeout int64

	collector metrics.Collector
}

func defaultOptions() *options {
	return &options{
		localMember:      nil,
		tombstoneTimeout: 30 * 60 * 1000,
		collector:        nil,
	}
}

type Option interface {
	apply(*options)
}

type localMemberOption struct {
	member *rpc.MemberState
}

func (o localMemberOption) apply(opts *options) {
	opts.localMember = o.member
}

func WithLocalMember(m *rpc.MemberState) Option {
	return localMemberOption{member: m}
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

type collectorOption struct {
	collector metrics.Collector
}

func (o collectorOption) apply(opts *options) {
	opts.collector = o.collector
}

func WithCollector(c metrics.Collector) Option {
	return collectorOption{collector: c}
}
