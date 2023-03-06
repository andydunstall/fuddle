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

package sdk

// Options contains options for configuring the registry client.
type Options struct {
}

// Option contains an option for configuring the registry client.
type Option func(*Options)

// NodeOptions contains options for querying the nodes in the registry.
type NodesOptions struct {
	filter Filter
}

// NodeOption contains an option for querying the nodes in the registry.
type NodesOption func(*NodesOptions)

func WithFilter(filter Filter) NodesOption {
	return func(opts *NodesOptions) {
		opts.filter = filter
	}
}
