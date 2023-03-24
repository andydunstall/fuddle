// Copyright (C) 2023 Andrew Dunstall
//
// Fuddle is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Fuddle is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package client

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
