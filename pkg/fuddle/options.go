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

package fuddle

import (
	"net"

	"go.uber.org/zap"
)

type options struct {
	logger   *zap.Logger
	listener net.Listener
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

type listenerOption struct {
	ln net.Listener
}

func (o listenerOption) apply(opts *options) {
	opts.listener = o.ln
}

// WithListener uses the given listener for the server instead of binding to a
// new listener.
func WithListener(ln net.Listener) Option {
	return listenerOption{ln: ln}
}
