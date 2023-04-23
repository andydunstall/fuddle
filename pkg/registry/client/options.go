package client

import (
	"time"

	"go.uber.org/zap"
)

type options struct {
	pendingUpdatesLimit int
	updateTimeout       time.Duration
	digestLimit         int
	logger              *zap.Logger
}

func defaultOptions() *options {
	return &options{
		pendingUpdatesLimit: 128,
		updateTimeout:       time.Second * 20,
		logger:              zap.NewNop(),
	}
}

type Option interface {
	apply(*options)
}

type pendingUpdatesLimitOption struct {
	limit int
}

func (o pendingUpdatesLimitOption) apply(opts *options) {
	opts.pendingUpdatesLimit = o.limit
}

func WithPendingUpdatesLimit(limit int) Option {
	return pendingUpdatesLimitOption{limit: limit}
}

type updateTimeoutOption struct {
	timeout time.Duration
}

func (o updateTimeoutOption) apply(opts *options) {
	opts.updateTimeout = o.timeout
}

func WithUpdateTimeoutOption(timeout time.Duration) Option {
	return updateTimeoutOption{timeout: timeout}
}

type digestLimitOption struct {
	limit int
}

func (o digestLimitOption) apply(opts *options) {
	opts.digestLimit = o.limit
}

func WithDigestLimit(limit int) Option {
	return digestLimitOption{limit: limit}
}

type loggerOption struct {
	log *zap.Logger
}

func (o loggerOption) apply(opts *options) {
	opts.logger = o.log
}

func WithLogger(log *zap.Logger) Option {
	return loggerOption{log: log}
}
