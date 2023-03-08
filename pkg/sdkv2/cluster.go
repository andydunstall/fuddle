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
	"fmt"
	"sync"
)

// subHandle is a handle for a cluster update subscriber.
type subHandle struct {
	Callback func()
}

// cluster maintains the registry clients view of the cluster.
type cluster struct {
	localID string

	nodes map[string]NodeState

	subs map[*subHandle]interface{}

	// mu protects the above fields.
	mu sync.Mutex
}

// newCluster returns a cluster containing only the given local node state.
func newCluster(node NodeState) *cluster {
	nodes := map[string]NodeState{
		node.ID: node.Copy(),
	}
	return &cluster{
		localID: node.ID,
		nodes:   nodes,
		subs:    make(map[*subHandle]interface{}),
		mu:      sync.Mutex{},
	}
}

// Nodes returns the set of nodes in the cluster.
func (c *cluster) Nodes(opts ...NodesOption) []NodeState {
	c.mu.Lock()
	defer c.mu.Unlock()

	options := &nodesOptions{}
	for _, o := range opts {
		o.apply(options)
	}

	nodes := make([]NodeState, 0, len(c.nodes))
	for _, n := range c.nodes {
		if options.filter == nil || options.filter.Match(n) {
			nodes = append(nodes, n.Copy())
		}
	}
	return nodes
}

// Subscribe registers the given callback to fire when the registry state
// changes.
//
// Note the callback is called synchronously with the registry mutex held,
// therefore it must NOT block or callback to the registry (or it will
// deadlock).
func (c *cluster) Subscribe(cb func()) func() {
	c.mu.Lock()
	defer c.mu.Unlock()

	handle := &subHandle{
		Callback: cb,
	}
	c.subs[handle] = struct{}{}

	return func() {
		c.unsubscribe(handle)
	}
}

// UpdateLocalState adds the given state to the local node. Note this is only
// adds or updates state, it won't remove entries.
func (c *cluster) UpdateLocalState(update map[string]string) error {
	return c.UpdateState(c.localID, update)
}

// AddNode adds the given node to the cluster.
func (c *cluster) AddNode(node NodeState) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if node.ID == "" {
		return fmt.Errorf("cluster: add node: node missing id")
	}

	c.nodes[node.ID] = node
	c.notifySubscribersLocked()
	return nil
}

func (c *cluster) RemoveNode(id string) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.nodes, id)
	c.notifySubscribersLocked()
}

func (c *cluster) UpdateState(id string, update map[string]string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	node, ok := c.nodes[id]
	if !ok {
		return fmt.Errorf("cluster: update state: node not found")
	}

	for k, v := range update {
		node.State[k] = v
	}
	c.notifySubscribersLocked()
	return nil
}

func (c *cluster) notifySubscribersLocked() {
	for sub := range c.subs {
		sub.Callback()
	}
}

func (c *cluster) unsubscribe(handle *subHandle) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.subs, handle)
}
