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

// ServiceQuery filters state within a service.
type ServiceQuery struct {
	State []string
}

// IsStateMatch returns true if the key matches the service query, false
// otherwise.
func (q *ServiceQuery) IsStateMatch(k string) bool {
	for _, queryKey := range q.State {
		if queryKey == k {
			return true
		}
	}
	return false
}

// Query filters subscriptions to only include the service, state and locality
// requested.
//
// The query maps services names to the state to include for that service.
type Query map[string]*ServiceQuery

// MatchingState returns all the given nodes state that matches the query, or
// false if either the node doesn't match or none of the nodes state matches.
func (q *Query) MatchingState(node Node) (state map[string]string, match bool) {
	serviceQuery, ok := (*q)[node.Service]
	// If the service is not in the query discard the node.
	if !ok {
		return nil, false
	}

	// If the service query is nil return all the state for the nodes matching
	// the services in the query.
	if serviceQuery == nil {
		return CopyState(node.State), true
	}

	// If there is a service query but there are no State fields return nil.
	if serviceQuery.State == nil || len(serviceQuery.State) == 0 {
		return nil, true
	}

	// Filter the nodes state. If there is no matching state discard the node.
	state = make(map[string]string)
	for k, v := range node.State {
		if serviceQuery.IsStateMatch(k) {
			state[k] = v
		}
	}
	if len(state) == 0 {
		return nil, false
	}
	return state, true
}
