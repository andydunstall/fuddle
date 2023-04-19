package registry

import (
	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"github.com/fuddle-io/fuddle/pkg/metrics"
)

type options struct {
	localMember *rpc.MemberState
	collector   metrics.Collector
}

func defaultOptions() *options {
	return &options{
		localMember: nil,
		collector:   nil,
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

type collectorOption struct {
	collector metrics.Collector
}

func (o collectorOption) apply(opts *options) {
	opts.collector = o.collector
}

func WithCollector(c metrics.Collector) Option {
	return collectorOption{collector: c}
}
