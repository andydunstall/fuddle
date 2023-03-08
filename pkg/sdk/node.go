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

// Node represents the state of a node in the cluster.
//
// It includes both fixed attributes of the node, and mutable application
// defined state.
type Node struct {
	// ID is a unique identifier for the node in the cluster.
	ID string

	// Service is the name of the service running on the node.
	Service string

	// Locality is the location of the node in the cluster.
	Locality string

	// Created is the time the node was created in UNIX milliseconds.
	Created int64

	// Revision identifies the version of the service running on the node.
	Revision string

	// State contains application defined key-value pairs.
	State map[string]string
}

func (s *Node) Equal(o Node) bool {
	if s.ID != o.ID {
		return false
	}
	if s.Service != o.Service {
		return false
	}
	if s.Locality != o.Locality {
		return false
	}
	if s.Created != o.Created {
		return false
	}
	if s.Revision != o.Revision {
		return false
	}
	if len(s.State) != len(o.State) {
		return false
	}
	for k, v1 := range s.State {
		v2, ok := o.State[k]
		if !ok {
			return false
		}
		if v1 != v2 {
			return false
		}
	}
	return true
}

func (s *Node) Copy() Node {
	cp := *s
	cp.State = CopyState(s.State)
	return cp
}

func CopyState(s map[string]string) map[string]string {
	if s == nil {
		return make(map[string]string)
	}

	cp := make(map[string]string)
	for k, v := range s {
		cp[k] = v
	}
	return cp
}
