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

import (
	"github.com/andydunstall/fuddle/pkg/util/wildcard"
)

// Filter specifies a node filter.
//
// This maps a service name (which may include wildcards with '*') to a service
// filter.
//
// Any nodes whose service don't match any of those listed are discarded.
type Filter map[string]ServiceFilter

func (f *Filter) Match(node NodeState) bool {
	// Must match at least one service, and all service filters where there
	// is a service name match.
	match := false
	for filterService, filter := range *f {
		if wildcard.Match(filterService, node.Service) {
			match = true

			if !filter.Match(node) {
				return false
			}
		}
	}
	return match
}

// ServiceFilter specifies a node filter that applies to all nodes in a service.
type ServiceFilter struct {
	// Locality is a list of localities (which may include wildcards with '*'),
	// where the nodes locality must match at least on of the listed localities.
	Locality []string

	// State contains the state filter.
	State StateFilter
}

func (f *ServiceFilter) Match(node NodeState) bool {
	// If there are no localites allow all.
	if f.Locality != nil {
		// The node locality must match at least one filter locality.
		match := false
		for _, filterLoc := range f.Locality {
			if wildcard.Match(filterLoc, node.Locality) {
				match = true
			}
		}
		if !match {
			return false
		}
	}

	return f.State.Match(node)
}

// StateFilter specifies a node filter that discards nodes whose state doesn't
// match the state listed.
//
// To match, for each filter key, the node must include a value for that key
// and match at least on of the filters for that key.
//
// The filter values may include wildcards, though the keys cannot.
type StateFilter map[string][]string

func (f *StateFilter) Match(node NodeState) bool {
	for filterKey, filterValues := range *f {
		v, ok := node.State[filterKey]
		// If the filter key is not in the node, its not a match.
		if !ok {
			return false
		}

		// The value must match at least one filter value.
		match := false
		for _, filterValue := range filterValues {
			if wildcard.Match(filterValue, v) {
				match = true
			}
		}
		if !match {
			return false
		}
	}

	return true
}
