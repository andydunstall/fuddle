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

package server

import (
	"net"

	"go.uber.org/zap"
)

type options struct {
	logger        *zap.Logger
	rpcListener   net.Listener
	adminListener net.Listener
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

type rpcListenerOption struct {
	ln net.Listener
}

func (o rpcListenerOption) apply(opts *options) {
	opts.rpcListener = o.ln
}

// WithRPCListener uses the given listener for the gRPC server instead of
// binding to a new listener.
func WithRPCListener(ln net.Listener) Option {
	return rpcListenerOption{ln: ln}
}

type adminListenerOption struct {
	ln net.Listener
}

func (o adminListenerOption) apply(opts *options) {
	opts.adminListener = o.ln
}

// WithAdminListener uses the given listener for the admin server instead of
// binding to a new listener.
func WithAdminListener(ln net.Listener) Option {
	return adminListenerOption{ln: ln}
}