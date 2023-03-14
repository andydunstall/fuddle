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

import (
	"fmt"
	"sync"
)

// subHandle is a handle for an RPC update subscriber.
type subHandle struct {
	Callback func(update *NodeUpdate)
}

// Cluster represents the shared view of the nodes in the cluster.
type Cluster struct {
	// nodes contains the node state for the nodes in the cluster, indexed by
	// node ID.
	nodes map[string]Node

	// subs contains a set of active RPC update subscribers.
	subs map[*subHandle]interface{}

	// mu protects the above fields.
	mu sync.Mutex
}

// NewCluster returns a cluster state containing only the given local node.
func NewCluster(localNode Node) *Cluster {
	nodes := map[string]Node{
		localNode.ID: localNode,
	}
	return &Cluster{
		nodes: nodes,
		subs:  make(map[*subHandle]interface{}),
		mu:    sync.Mutex{},
	}
}

// Node returns the state of the node in the cluster with the given ID.
func (s *Cluster) Node(id string) (Node, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	node, ok := s.nodes[id]
	if !ok {
		return Node{}, false
	}
	return node.Copy(), true
}

func (s *Cluster) Nodes() []Node {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.nodesLocked()
}

// ApplyUpdate applies the given node state update and sends it to the
// subscribers.
func (s *Cluster) ApplyUpdate(update *NodeUpdate) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch update.UpdateType {
	case UpdateTypeRegister:
		if err := s.applyRegisterUpdateLocked(update); err != nil {
			return err
		}
	case UpdateTypeUnregister:
		if err := s.applyUnregisterUpdateLocked(update); err != nil {
			return err
		}
	case UpdateTypeMetadata:
		if err := s.applyMetadataUpdateLocked(update); err != nil {
			return err
		}
	default:
		return fmt.Errorf("cluster state: unknown update type: %s", update.UpdateType)
	}

	// Notify the subscribers of the update. Note keeping mutex locked to
	// guarantee ordering.
	for sub := range s.subs {
		sub.Callback(update)
	}

	return nil
}

// Subscribe subscribes to RPC to updates.
//
// The callback is called with the cluster state mutex held (to guarantee
// ordering) so it MUST NOT block and MUST NOT call back to the cluster state.
//
// If rewind is true the callback is called with join updates for existing
// nodes. Used to get the current node state and subscribe in one atomic
// transaction.
//
// Returns a function to unsubscribe.
func (s *Cluster) Subscribe(rewind bool, cb func(update *NodeUpdate)) func() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if rewind {
		// Send all existing nodes as join events. This must keep the mutex
		// locked to avoid breaking order guarantees (such as receiving a
		// state update event before a join event for a node).
		for _, node := range s.nodes {
			update := &NodeUpdate{
				ID:         node.ID,
				UpdateType: UpdateTypeRegister,
				Attributes: &NodeAttributes{
					Service:  node.Service,
					Locality: node.Locality,
					Created:  node.Created,
					Revision: node.Revision,
				},
				// Copy state since the node state may be modified.
				Metadata: CopyMetadata(node.Metadata),
			}
			cb(update)
		}
	}

	handle := &subHandle{
		Callback: cb,
	}
	s.subs[handle] = struct{}{}

	return func() {
		s.unsubscribeUpdates(handle)
	}
}

func (s *Cluster) nodesLocked() []Node {
	var nodes []Node
	for _, node := range s.nodes {
		nodes = append(nodes, node.Copy())
	}
	return nodes
}

func (s *Cluster) unsubscribeUpdates(handle *subHandle) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.subs, handle)
}

func (s *Cluster) applyRegisterUpdateLocked(update *NodeUpdate) error {
	if update.ID == "" {
		return fmt.Errorf("cluster state: register update: missing id")
	}

	if update.Attributes == nil {
		return fmt.Errorf("cluster state: register update: missing attributes")
	}

	node := Node{
		ID:       update.ID,
		Service:  update.Attributes.Service,
		Locality: update.Attributes.Locality,
		Created:  update.Attributes.Created,
		Revision: update.Attributes.Revision,
		// Copy the state to avoid modifying the update. If update.Metadata is
		// nil this returns an empty map.
		Metadata: CopyMetadata(update.Metadata),
	}
	s.nodes[node.ID] = node

	return nil
}

func (s *Cluster) applyUnregisterUpdateLocked(update *NodeUpdate) error {
	delete(s.nodes, update.ID)

	return nil
}

func (s *Cluster) applyMetadataUpdateLocked(update *NodeUpdate) error {
	node, ok := s.nodes[update.ID]
	if !ok {
		return fmt.Errorf("cluster state: metadata update: node does not exist")
	}

	// If the update is missing metadata just ignore it.
	if update.Metadata == nil {
		return nil
	}
	for k, v := range update.Metadata {
		node.Metadata[k] = v
	}

	return nil
}
