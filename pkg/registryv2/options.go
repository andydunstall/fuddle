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

package registry

import (
	"time"

	"go.uber.org/zap"
)

type options struct {
	logger           *zap.Logger
	time             time.Time
	heartbeatTimeout time.Duration
	reconnectTimeout time.Duration
}

type Option interface {
	apply(*options)
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

type timeOption struct {
	time time.Time
}

func (o timeOption) apply(opts *options) {
	opts.time = o.time
}

func WithTime(t time.Time) Option {
	return timeOption{time: t}
}

type heartbeatTimeoutOption struct {
	timeout time.Duration
}

func (o heartbeatTimeoutOption) apply(opts *options) {
	opts.heartbeatTimeout = o.timeout
}

func WithHeartbeatTimeout(timeout time.Duration) Option {
	return heartbeatTimeoutOption{timeout: timeout}
}

type reconnectTimeoutOption struct {
	timeout time.Duration
}

func (o reconnectTimeoutOption) apply(opts *options) {
	opts.reconnectTimeout = o.timeout
}

func WithReconnectTimeout(timeout time.Duration) Option {
	return reconnectTimeoutOption{timeout: timeout}
}
