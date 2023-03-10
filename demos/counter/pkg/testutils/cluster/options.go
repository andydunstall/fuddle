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

package cluster

type options struct {
	fuddleNodes   int
	counterNodes  int
	frontendNodes int
}

type Option interface {
	apply(*options)
}

type fuddleNodesOption int

func (c fuddleNodesOption) apply(opts *options) {
	opts.fuddleNodes = int(c)
}

func WithFuddleNodes(c int) Option {
	return fuddleNodesOption(c)
}

type counterNodesOption int

func (c counterNodesOption) apply(opts *options) {
	opts.counterNodes = int(c)
}

func WithCounterNodes(c int) Option {
	return counterNodesOption(c)
}

type frontendNodesOption int

func (c frontendNodesOption) apply(opts *options) {
	opts.frontendNodes = int(c)
}

func WithFrontendNodes(c int) Option {
	return frontendNodesOption(c)
}
