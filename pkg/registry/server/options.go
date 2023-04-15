package server

import (
	"go.uber.org/zap"
)

type options struct {
	logger *zap.Logger
}

func defaultOptions() *options {
	return &options{
		logger: zap.NewNop(),
	}
}

type Option interface {
	apply(*options)
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
