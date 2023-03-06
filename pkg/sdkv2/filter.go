// Copyright (C) 2023 Andrew Dunstall
//
// Registry is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Registry is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package sdk

// Filter specifies a node filter.
//
// This maps a service name (which may include wildcards with '*') to a service
// filter.
//
// Any nodes whose service don't match any of those listed are discarded.
type Filter map[string]ServiceFilter

// ServiceFilter specifies a node filter that applies to all nodes in a service.
type ServiceFilter struct {
	// Locality is a list of localities (which may include wildcards with '*'),
	// where the nodes locality must match at least on of the listed localities.
	Locality []string

	// State contains the state filter.
	State StateFilter
}

// StateFilter specifies a node filter that discards nodes whose state doesn't
// match the state listed.
//
// The filter maps state keys (which may include wildcards with '*') to the
// set of allowed state values (which can also include wildcards).
type StateFilter map[string][]string
