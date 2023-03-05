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

// updateSubHandle is a handle for an RPC update subscriber.
type updateSubHandle struct {
	Callback func(update *rpc.NodeUpdate)
}

// nodesSubHandle is a handle for an nodes subscriber.
type nodesSubHandle struct {
	Callback func(nodes []NodeState)
	Query    *Query
}

// ClusterState represents the shared view of the nodes in the cluster.
type ClusterState struct {
	// nodes contains the node state for the nodes in the cluster, indexed by
	// node ID.
	nodes map[string]NodeState

	// updateSubs contains a set of active RPC update subscribers.
	updateSubs map[*updateSubHandle]interface{}

	// nodesSubs contains a set of active nodes subscribers.
	nodesSubs map[*nodesSubHandle]interface{}

	// mu protects the above fields.
	mu sync.Mutex
}

// NewClusterState returns a cluster state containing only the given local node.
func NewClusterState(localNode NodeState) *ClusterState {
	nodes := map[string]NodeState{
		localNode.ID: localNode,
	}
	return &ClusterState{
		nodes:      nodes,
		updateSubs: make(map[*updateSubHandle]interface{}),
		nodesSubs:  make(map[*nodesSubHandle]interface{}),
		mu:         sync.Mutex{},
	}
}

// Node returns the state of the node in the cluster with the given ID.
func (s *ClusterState) Node(id string) (NodeState, bool) {
	s.mu.Lock()
	defer s.mu.Unlock()

	node, ok := s.nodes[id]
	if !ok {
		return NodeState{}, false
	}
	return node.Copy(), true
}

func (s *ClusterState) Nodes(query *Query) []NodeState {
	s.mu.Lock()
	defer s.mu.Unlock()

	return s.nodesLocked(query)
}

// ApplyUpdate applies the given node state update and sends it to the
// subscribers.
func (s *ClusterState) ApplyUpdate(update *rpc.NodeUpdate) error {
	s.mu.Lock()
	defer s.mu.Unlock()

	switch update.UpdateType {
	case rpc.NodeUpdateType_JOIN:
		if err := s.applyJoinUpdateLocked(update); err != nil {
			return err
		}
	case rpc.NodeUpdateType_LEAVE:
		if err := s.applyLeaveUpdateLocked(update); err != nil {
			return err
		}
	case rpc.NodeUpdateType_STATE:
		if err := s.applyStateUpdateLocked(update); err != nil {
			return err
		}
	default:
		return fmt.Errorf("cluster state: unknown update type: %s", update.UpdateType)
	}

	// Notify the subscribers of the update. Note keeping mutex locked to
	// guarantee ordering.
	for sub := range s.updateSubs {
		sub.Callback(update)
	}

	for sub := range s.nodesSubs {
		nodes := s.nodesLocked(sub.Query)
		if len(nodes) != 0 {
			sub.Callback(nodes)
		}
	}

	return nil
}

// SubscribeUpdates subscribes to RPC to updates.
//
// The callback is called with the cluster state mutex held (to guarantee
// ordering) so it MUST NOT block and MUST NOT call back to the cluster state.
//
// If rewind is true the callback is called with join updates for existing
// nodes. Used to get the current node state and subscribe in one atomic
// transaction.
//
// Returns a function to unsubscribe.
func (s *ClusterState) SubscribeUpdates(rewind bool, cb func(update *rpc.NodeUpdate)) func() {
	s.mu.Lock()
	defer s.mu.Unlock()

	if rewind {
		// Send all existing nodes as join events. This must keep the mutex
		// locked to avoid breaking order guarantees (such as receiving a
		// state update event before a join event for a node).
		for _, node := range s.nodes {
			update := &rpc.NodeUpdate{
				NodeId:     node.ID,
				UpdateType: rpc.NodeUpdateType_JOIN,
				Attributes: &rpc.Attributes{
					Service:  node.Service,
					Locality: node.Locality,
					Revision: node.Revision,
				},
				// Copy state since the node state may be modified.
				State: CopyState(node.State),
			}
			cb(update)
		}
	}

	handle := &updateSubHandle{
		Callback: cb,
	}
	s.updateSubs[handle] = struct{}{}

	return func() {
		s.unsubscribeUpdates(handle)
	}
}

// SubscribeNodes subscribes too state updates matching the given node.
//
// Returns a function to unsubscribe.
func (s *ClusterState) SubscribeNodes(query *Query, cb func([]NodeState)) func() {
	s.mu.Lock()
	defer s.mu.Unlock()

	cb(s.nodesLocked(query))

	handle := &nodesSubHandle{
		Callback: cb,
		Query:    query,
	}
	s.nodesSubs[handle] = struct{}{}

	return func() {
		s.unsubscribeNodes(handle)
	}
}

func (s *ClusterState) nodesLocked(query *Query) []NodeState {
	var nodes []NodeState
	for _, node := range s.nodes {
		// If the query is nil include all nodes.
		if query == nil {
			nodes = append(nodes, node.Copy())
			continue
		}

		state, match := query.MatchingState(node)
		if !match {
			continue
		}

		node.State = state
		nodes = append(nodes, node)
	}

	return nodes
}

func (s *ClusterState) unsubscribeUpdates(handle *updateSubHandle) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.updateSubs, handle)
}

func (s *ClusterState) unsubscribeNodes(handle *nodesSubHandle) {
	s.mu.Lock()
	defer s.mu.Unlock()

	delete(s.nodesSubs, handle)
}

func (s *ClusterState) applyJoinUpdateLocked(update *rpc.NodeUpdate) error {
	if update.NodeId == "" {
		return fmt.Errorf("cluster state: join update: missing id")
	}

	if update.Attributes == nil {
		return fmt.Errorf("cluster state: join update: missing attributes")
	}

	node := NodeState{
		ID:       update.NodeId,
		Service:  update.Attributes.Service,
		Locality: update.Attributes.Locality,
		Revision: update.Attributes.Revision,
		// Copy the state to avoid modifying the update. If update.State is
		// nil this returns an empty map.
		State: CopyState(update.State),
	}
	s.nodes[node.ID] = node

	return nil
}

func (s *ClusterState) applyLeaveUpdateLocked(update *rpc.NodeUpdate) error {
	delete(s.nodes, update.NodeId)

	return nil
}

func (s *ClusterState) applyStateUpdateLocked(update *rpc.NodeUpdate) error {
	node, ok := s.nodes[update.NodeId]
	if !ok {
		return fmt.Errorf("cluster state: state update: node does not exist")
	}

	// If the update is missing state must ignore it.
	if update.State == nil {
		return nil
	}
	for k, v := range update.State {
		node.State[k] = v
	}

	return nil
}
