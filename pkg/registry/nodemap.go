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
	"fmt"
	"sync"

	"github.com/andydunstall/fuddle/pkg/rpc"
)

type subHandle struct {
	Callback func(update *rpc.NodeUpdate)
}

type NodeState struct {
	ID       string            `json:"id,omitempty"`
	Service  string            `json:"service,omitempty"`
	Locality string            `json:"locality,omitempty"`
	Revision string            `json:"revision,omitempty"`
	State    map[string]string `json:"state,omitempty"`
}

type NodeMap struct {
	nodes       map[string]NodeState
	subscribers map[*subHandle]interface{}

	mu sync.Mutex
}

func NewNodeMap(node NodeState) *NodeMap {
	nodes := make(map[string]NodeState)
	nodes[node.ID] = node
	return &NodeMap{
		nodes:       nodes,
		subscribers: make(map[*subHandle]interface{}),
	}
}

func (m *NodeMap) Node(id string) (NodeState, bool) {
	m.mu.Lock()
	defer m.mu.Unlock()

	node, ok := m.nodes[id]
	if !ok {
		return NodeState{}, false
	}

	nodeCopy := node
	nodeCopy.State = make(map[string]string)
	for k, v := range node.State {
		nodeCopy.State[k] = v
	}

	return nodeCopy, true
}

func (m *NodeMap) Nodes() []NodeState {
	m.mu.Lock()
	defer m.mu.Unlock()

	var nodes []NodeState
	for _, node := range m.nodes {
		nodeCopy := node
		nodeCopy.State = make(map[string]string)
		for k, v := range node.State {
			nodeCopy.State[k] = v
		}
		nodes = append(nodes, nodeCopy)
	}

	return nodes
}

func (m *NodeMap) Update(update *rpc.NodeUpdate) error {
	m.mu.Lock()
	defer m.mu.Unlock()

	switch update.UpdateType {
	case rpc.UpdateType_NODE_JOIN:
		if update.Attributes == nil {
			return fmt.Errorf("node join: missing attributes")
		}

		// Copy the state since the update will be propagated to subscribers
		// so must not be modified.
		state := map[string]string{}
		for k, v := range update.State {
			state[k] = v
		}

		m.nodes[update.NodeId] = NodeState{
			ID:       update.NodeId,
			Service:  update.Attributes.Service,
			Locality: update.Attributes.Locality,
			Revision: update.Attributes.Revision,
			State:    state,
		}
	case rpc.UpdateType_NODE_LEAVE:
		delete(m.nodes, update.NodeId)
	case rpc.UpdateType_NODE_UPDATE:
		if update.State == nil {
			return fmt.Errorf("node update: missing state")
		}

		node, ok := m.nodes[update.NodeId]
		if !ok {
			return fmt.Errorf("node update: node does not exist")
		}
		if node.State == nil {
			node.State = make(map[string]string)
		}
		for k, v := range update.State {
			node.State[k] = v
		}
	}

	var subs []*subHandle
	for sub := range m.subscribers {
		subs = append(subs, sub)
	}

	for _, sub := range subs {
		sub.Callback(update)
	}

	return nil
}

// Subscribe to node updates. The callback MUST NOT block, or modify the
// update.
//
// If rewind is true all existing nodes are send as JOIN events. This is to
// subscribe without missing updates.
func (m *NodeMap) Subscribe(rewind bool, cb func(update *rpc.NodeUpdate)) func() {
	m.mu.Lock()
	defer m.mu.Unlock()

	// If rewind send all existing node state as JOIN events. This must be
	// send before unlocking the mutex to avoid interleaving bootstrap updates
	// new updates. Therefore the callback must not block.
	if rewind {
		for _, node := range m.nodes {
			// Copy the node state since the update state must not be modified.
			state := map[string]string{}
			for k, v := range node.State {
				state[k] = v
			}

			update := &rpc.NodeUpdate{
				NodeId:     node.ID,
				UpdateType: rpc.UpdateType_NODE_JOIN,
				Attributes: &rpc.Attributes{
					Id:       node.ID,
					Service:  node.Service,
					Locality: node.Locality,
					Revision: node.Revision,
				},
				State: state,
			}
			cb(update)
		}
	}

	handle := &subHandle{
		Callback: cb,
	}
	m.subscribers[handle] = struct{}{}

	return func() {
		m.unsubscribe(handle)
	}
}

func (m *NodeMap) unsubscribe(handle *subHandle) {
	m.mu.Lock()
	defer m.mu.Unlock()

	delete(m.subscribers, handle)
}
