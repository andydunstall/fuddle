package admin

import (
	"time"
)

type options struct {
	connectTimeout time.Duration
}

type Option interface {
	apply(*options)
}

type connectTimeoutOption struct {
	timeout time.Duration
}

func (o connectTimeoutOption) apply(opts *options) {
	opts.connectTimeout = o.timeout
}

// WithConnectTimeout defines the time to wait for each connection attempt
// before timing out. Default to 1 second.
func WithConnectTimeout(timeout time.Duration) Option {
	return connectTimeoutOption{timeout: timeout}
}
