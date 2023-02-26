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
	"sync"
)

// NodeMap maintains the registered nodes in the cluster.
type NodeMap struct {
	nodes map[string]interface{}

	mu sync.Mutex
}

func NewNodeMap() *NodeMap {
	return &NodeMap{
		nodes: make(map[string]interface{}),
		mu:    sync.Mutex{},
	}
}

func (m *NodeMap) NodeIDs() []string {
	m.mu.Lock()
	defer m.mu.Unlock()

	nodeIDs := make([]string, 0, len(m.nodes))
	for id := range m.nodes {
		nodeIDs = append(nodeIDs, id)
	}
	return nodeIDs
}

func (m *NodeMap) Register(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.nodes[id] = struct{}{}
}

func (m *NodeMap) Unregister(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.nodes, id)
}
