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

	subscribers map[string]func()

	mu sync.Mutex
}

func NewNodeMap() *NodeMap {
	return &NodeMap{
		nodes:       make(map[string]interface{}),
		subscribers: make(map[string]func()),
		mu:          sync.Mutex{},
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
	m.register(id)
	m.notifySubscribers()
}

func (m *NodeMap) Unregister(id string) {
	m.unregister(id)
	m.notifySubscribers()
}

// Subscribe to nodemap updates using the given ID to identify the subscriber.
// The subscriber must not block.
func (m *NodeMap) Subscribe(id string, cb func()) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.subscribers[id] = cb
}

// Unsubscribe the subscriber with the given ID.
func (m *NodeMap) Unsubscribe(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.subscribers, id)
}

func (m *NodeMap) register(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	m.nodes[id] = struct{}{}
}

func (m *NodeMap) unregister(id string) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.nodes, id)
}

func (m *NodeMap) notifySubscribers() {
	// Copy the subscribers to notify without the mutex held.
	m.mu.Lock()
	subs := make([]func(), 0, len(m.subscribers))
	for _, sub := range m.subscribers {
		subs = append(subs, sub)
	}
	m.mu.Unlock()

	for _, sub := range subs {
		sub()
	}
}
