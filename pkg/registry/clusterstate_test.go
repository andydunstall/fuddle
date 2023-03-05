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
	"sort"
	"testing"

	"github.com/andydunstall/fuddle/pkg/rpc"
	"github.com/stretchr/testify/assert"
)

func TestClusterState_Node(t *testing.T) {
	local := NodeState{
		ID:       "local-123",
		Service:  "foo",
		Locality: "us-east-1-a",
		Revision: "v0.1.0",
		State:    make(map[string]string),
	}
	cs := NewClusterState(local)

	// Verify ClusterState.Node returns the node with the given ID.
	node, ok := cs.Node("local-123")
	assert.True(t, ok)
	assert.Equal(t, local, node)

	// Verify the returned node is a copy by modifying its state and getting
	// the name ndoe again.
	node.State["a"] = "5"
	nodeCopy, ok := cs.Node("local-123")
	assert.True(t, ok)
	assert.Equal(t, local, nodeCopy)
}

func TestClusterState_NodesNoQuery(t *testing.T) {
	local := NodeState{
		ID: "local-123",
		State: map[string]string{
			"foo": "bar",
		},
	}
	cs := NewClusterState(local)

	joinedNodes := []NodeState{
		{
			ID:       "remote-1",
			Service:  "foo",
			Locality: "us-east-1-a",
			Revision: "v0.1.0",
			State: map[string]string{
				"addr.foo": "10.26.104.54:8138",
				"addr.bar": "10.26.104.23:1122",
			},
		},
		{
			ID:       "remote-2",
			Service:  "bar",
			Locality: "us-east-1-a",
			Revision: "v0.1.0",
			State: map[string]string{
				"addr.foo": "10.26.104.54:8138",
				"addr.bar": "10.26.104.23:1122",
			},
		},
	}
	for _, node := range joinedNodes {
		update := &rpc.NodeUpdate{
			NodeId:     node.ID,
			UpdateType: rpc.NodeUpdateType_JOIN,
			Attributes: &rpc.Attributes{
				Service:  node.Service,
				Locality: node.Locality,
				Revision: node.Revision,
			},
			State: node.State,
		}
		assert.Nil(t, cs.ApplyUpdate(update))
	}

	// Check Nodes returns all joined nodes and the local node.
	joinedNodes = append([]NodeState{local}, joinedNodes...)

	nodes := cs.Nodes(nil)
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].ID < nodes[j].ID
	})
	assert.Equal(t, joinedNodes, nodes)
}

func TestClusterState_NodesWithQuery(t *testing.T) {
	local := NodeState{
		ID: "local-123",
		State: map[string]string{
			"foo": "bar",
		},
	}
	cs := NewClusterState(local)

	joinedNodes := []NodeState{
		{
			ID:       "remote-1",
			Service:  "foo",
			Locality: "us-east-1-a",
			Revision: "v0.1.0",
			State: map[string]string{
				"addr.foo": "10.26.104.54:8138",
				"addr.bar": "10.26.104.23:1122",
			},
		},
		{
			ID:       "remote-2",
			Service:  "bar",
			Locality: "us-east-1-a",
			Revision: "v0.1.0",
			State: map[string]string{
				"addr.foo": "10.26.104.54:8138",
				"addr.bar": "10.26.104.23:1122",
			},
		},
	}
	for _, node := range joinedNodes {
		update := &rpc.NodeUpdate{
			NodeId:     node.ID,
			UpdateType: rpc.NodeUpdateType_JOIN,
			Attributes: &rpc.Attributes{
				Service:  node.Service,
				Locality: node.Locality,
				Revision: node.Revision,
			},
			State: node.State,
		}
		assert.Nil(t, cs.ApplyUpdate(update))
	}

	nodes := cs.Nodes(&Query{
		"foo": &ServiceQuery{
			State: []string{"addr.bar"},
		},
	})
	assert.Equal(t, []NodeState{{
		ID:       "remote-1",
		Service:  "foo",
		Locality: "us-east-1-a",
		Revision: "v0.1.0",
		State: map[string]string{
			"addr.bar": "10.26.104.23:1122",
		},
	}}, nodes)
}

func TestClusterState_NodeNotFound(t *testing.T) {
	cs := NewClusterState(NodeState{
		ID: "local-123",
	})
	_, ok := cs.Node("not-found")
	assert.False(t, ok)
}

// Tests applying a join update adds the node to the cluster state.
func TestClusterState_ApplyJoinUpdate(t *testing.T) {
	cs := NewClusterState(NodeState{
		ID: "local-123",
	})

	joinedNode := NodeState{
		ID:       "remote-123",
		Service:  "foo",
		Locality: "us-east-1-a",
		Revision: "v0.1.0",
		State: map[string]string{
			"addr.foo": "10.26.104.54:8138",
			"addr.bar": "10.26.104.23:1122",
		},
	}
	update := &rpc.NodeUpdate{
		NodeId:     joinedNode.ID,
		UpdateType: rpc.NodeUpdateType_JOIN,
		Attributes: &rpc.Attributes{
			Service:  joinedNode.Service,
			Locality: joinedNode.Locality,
			Revision: joinedNode.Revision,
		},
		State: joinedNode.State,
	}
	assert.Nil(t, cs.ApplyUpdate(update))

	// Verify ClusterState.Node returns the added node.
	node, ok := cs.Node("remote-123")
	assert.True(t, ok)
	assert.Equal(t, joinedNode, node)
}

// Tests applying a join update with no ID returns an error.
func TestClusterState_ApplyJoinUpdateMissingID(t *testing.T) {
	cs := NewClusterState(NodeState{
		ID: "local-123",
	})
	err := cs.ApplyUpdate(&rpc.NodeUpdate{
		UpdateType: rpc.NodeUpdateType_JOIN,
		Attributes: &rpc.Attributes{},
	})
	assert.NotNil(t, err)
}

// Tests applying a join update with no attributes returns an error.
func TestClusterState_ApplyJoinUpdateMissingAttributes(t *testing.T) {
	cs := NewClusterState(NodeState{
		ID: "local-123",
	})
	err := cs.ApplyUpdate(&rpc.NodeUpdate{
		NodeId:     "remote-123",
		UpdateType: rpc.NodeUpdateType_JOIN,
	})
	assert.NotNil(t, err)
}

// Tests applying a leave update removes the node to the cluster state.
func TestClusterState_ApplyLeaveUpdate(t *testing.T) {
	cs := NewClusterState(NodeState{
		ID: "local-123",
	})

	// Apply a join update the check the node is added.
	update := &rpc.NodeUpdate{
		NodeId:     "remote-123",
		UpdateType: rpc.NodeUpdateType_JOIN,
		Attributes: &rpc.Attributes{},
	}
	assert.Nil(t, cs.ApplyUpdate(update))
	_, ok := cs.Node("remote-123")
	assert.True(t, ok)

	// Apply a leave update the check the node is removed.
	assert.Nil(t, cs.ApplyUpdate(&rpc.NodeUpdate{
		NodeId:     "remote-123",
		UpdateType: rpc.NodeUpdateType_LEAVE,
		Attributes: &rpc.Attributes{},
	}))
	_, ok = cs.Node("remote-123")
	assert.False(t, ok)
}

func TestClusterState_ApplyStateUpdate(t *testing.T) {
	cs := NewClusterState(NodeState{
		ID: "local-123",
	})

	// Apply a join update the check the node is added.
	update := &rpc.NodeUpdate{
		NodeId:     "remote-123",
		UpdateType: rpc.NodeUpdateType_JOIN,
		Attributes: &rpc.Attributes{},
	}
	assert.Nil(t, cs.ApplyUpdate(update))
	_, ok := cs.Node("remote-123")
	assert.True(t, ok)

	// Apply state updates and check the node state is updated.
	assert.Nil(t, cs.ApplyUpdate(&rpc.NodeUpdate{
		NodeId:     "remote-123",
		UpdateType: rpc.NodeUpdateType_STATE,
		State: map[string]string{
			"foo": "1",
			"bar": "2",
		},
	}))
	assert.Nil(t, cs.ApplyUpdate(&rpc.NodeUpdate{
		NodeId:     "remote-123",
		UpdateType: rpc.NodeUpdateType_STATE,
		State: map[string]string{
			"car": "3",
		},
	}))

	node, ok := cs.Node("remote-123")
	assert.True(t, ok)
	assert.Equal(t, map[string]string{
		"foo": "1",
		"bar": "2",
		"car": "3",
	}, node.State)
}

// Tests applying a state update where the node is not found.
func TestClusterState_ApplyStateUpdateNodeNotFound(t *testing.T) {
	cs := NewClusterState(NodeState{
		ID: "local-123",
	})

	err := cs.ApplyUpdate(&rpc.NodeUpdate{
		NodeId:     "remote-123",
		UpdateType: rpc.NodeUpdateType_STATE,
		State: map[string]string{
			"foo": "1",
		},
	})
	assert.NotNil(t, err)
}

// Tests applying an update of unknown type returns an error.
func TestClusterState_ApplyUnknownUpdate(t *testing.T) {
	cs := NewClusterState(NodeState{
		ID: "local-123",
	})
	err := cs.ApplyUpdate(&rpc.NodeUpdate{
		UpdateType: 999,
	})
	assert.NotNil(t, err)
}

// Tests subscribing to cluster state updates by applying the applied updates to
// another cluster state and checking they are equal.
func TestClusterState_SubscribeUpdates(t *testing.T) {
	cs1 := NewClusterState(NodeState{
		ID: "local-node",
	})
	cs2 := NewClusterState(NodeState{
		ID: "local-node",
	})
	// Subscribe to updates from the first cluster state and apply to the
	// second.
	cs1.SubscribeUpdates(false, func(update *rpc.NodeUpdate) {
		assert.Nil(t, cs2.ApplyUpdate(update))
	})

	// Apply JOIN updates and check applied to both maps.
	assert.Nil(t, cs1.ApplyUpdate(&rpc.NodeUpdate{
		NodeId:     "node-1",
		UpdateType: rpc.NodeUpdateType_JOIN,
		Attributes: &rpc.Attributes{
			Service: "foo",
		},
		State: map[string]string{
			"a": "1",
		},
	}))
	assert.Nil(t, cs1.ApplyUpdate(&rpc.NodeUpdate{
		NodeId:     "node-2",
		UpdateType: rpc.NodeUpdateType_JOIN,
		Attributes: &rpc.Attributes{
			Service: "bar",
		},
		State: map[string]string{
			"b": "2",
		},
	}))

	nodes1 := cs1.Nodes(nil)
	sort.Slice(nodes1, func(i, j int) bool {
		return nodes1[i].ID < nodes1[j].ID
	})
	nodes2 := cs2.Nodes(nil)
	sort.Slice(nodes2, func(i, j int) bool {
		return nodes2[i].ID < nodes2[j].ID
	})
	assert.Equal(t, nodes1, nodes2)

	// Apply STATE updates and check applied to both maps.
	assert.Nil(t, cs1.ApplyUpdate(&rpc.NodeUpdate{
		NodeId:     "node-1",
		UpdateType: rpc.NodeUpdateType_STATE,
		State: map[string]string{
			"a": "10",
		},
	}))
	assert.Nil(t, cs1.ApplyUpdate(&rpc.NodeUpdate{
		NodeId:     "node-2",
		UpdateType: rpc.NodeUpdateType_STATE,
		State: map[string]string{
			"b": "20",
		},
	}))

	nodes1 = cs1.Nodes(nil)
	sort.Slice(nodes1, func(i, j int) bool {
		return nodes1[i].ID < nodes1[j].ID
	})
	nodes2 = cs2.Nodes(nil)
	sort.Slice(nodes2, func(i, j int) bool {
		return nodes2[i].ID < nodes2[j].ID
	})
	assert.Equal(t, nodes1, nodes2)

	// Apply LEAVE updates and check applied to both maps.
	assert.Nil(t, cs1.ApplyUpdate(&rpc.NodeUpdate{
		NodeId:     "node-1",
		UpdateType: rpc.NodeUpdateType_LEAVE,
	}))

	nodes1 = cs1.Nodes(nil)
	sort.Slice(nodes1, func(i, j int) bool {
		return nodes1[i].ID < nodes1[j].ID
	})
	nodes2 = cs2.Nodes(nil)
	sort.Slice(nodes2, func(i, j int) bool {
		return nodes2[i].ID < nodes2[j].ID
	})
	assert.Equal(t, nodes1, nodes2)
}

// Tests subscribing to cluster state with rewind and applying updates to the
// other cluster has the same state.
func TestClusterState_SubscribeUpdatesWithRewind(t *testing.T) {
	cs1 := NewClusterState(NodeState{
		ID: "local-node",
	})
	cs2 := NewClusterState(NodeState{
		ID: "local-node",
	})

	// Apply JOIN updates and check applied to both maps.
	assert.Nil(t, cs1.ApplyUpdate(&rpc.NodeUpdate{
		NodeId:     "node-1",
		UpdateType: rpc.NodeUpdateType_JOIN,
		Attributes: &rpc.Attributes{
			Service: "foo",
		},
		State: map[string]string{
			"a": "1",
		},
	}))
	assert.Nil(t, cs1.ApplyUpdate(&rpc.NodeUpdate{
		NodeId:     "node-2",
		UpdateType: rpc.NodeUpdateType_JOIN,
		Attributes: &rpc.Attributes{
			Service: "bar",
		},
		State: map[string]string{
			"b": "2",
		},
	}))

	// Subscribe to updates from the first cluster with rewind and apply to
	// the second. Note only subscribing after the state updates.
	cs1.SubscribeUpdates(true, func(update *rpc.NodeUpdate) {
		assert.Nil(t, cs2.ApplyUpdate(update))
	})

	nodes1 := cs1.Nodes(nil)
	sort.Slice(nodes1, func(i, j int) bool {
		return nodes1[i].ID < nodes1[j].ID
	})
	nodes2 := cs2.Nodes(nil)
	sort.Slice(nodes2, func(i, j int) bool {
		return nodes2[i].ID < nodes2[j].ID
	})
	assert.Equal(t, nodes1, nodes2)

	// Apply STATE updates and check applied to both maps.
	assert.Nil(t, cs1.ApplyUpdate(&rpc.NodeUpdate{
		NodeId:     "node-1",
		UpdateType: rpc.NodeUpdateType_STATE,
		State: map[string]string{
			"a": "10",
		},
	}))
	assert.Nil(t, cs1.ApplyUpdate(&rpc.NodeUpdate{
		NodeId:     "node-2",
		UpdateType: rpc.NodeUpdateType_STATE,
		State: map[string]string{
			"b": "20",
		},
	}))

	nodes1 = cs1.Nodes(nil)
	sort.Slice(nodes1, func(i, j int) bool {
		return nodes1[i].ID < nodes1[j].ID
	})
	nodes2 = cs2.Nodes(nil)
	sort.Slice(nodes2, func(i, j int) bool {
		return nodes2[i].ID < nodes2[j].ID
	})
	assert.Equal(t, nodes1, nodes2)
}

func TestClusterState_SubscribeNodes(t *testing.T) {
	local := NodeState{
		ID: "local-123",
		State: map[string]string{
			"foo": "bar",
		},
	}
	cs := NewClusterState(local)

	var nodes []NodeState
	cs.SubscribeNodes(&Query{
		"foo": &ServiceQuery{
			State: []string{"addr.bar"},
		},
	}, func(n []NodeState) {
		nodes = n
	})

	joinedNodes := []NodeState{
		{
			ID:       "remote-1",
			Service:  "foo",
			Locality: "us-east-1-a",
			Revision: "v0.1.0",
			State: map[string]string{
				"addr.foo": "10.26.104.54:8138",
				"addr.bar": "10.26.104.23:1122",
			},
		},
		{
			ID:       "remote-2",
			Service:  "bar",
			Locality: "us-east-1-a",
			Revision: "v0.1.0",
			State: map[string]string{
				"addr.foo": "10.26.104.54:8138",
				"addr.bar": "10.26.104.23:1122",
			},
		},
	}
	for _, node := range joinedNodes {
		update := &rpc.NodeUpdate{
			NodeId:     node.ID,
			UpdateType: rpc.NodeUpdateType_JOIN,
			Attributes: &rpc.Attributes{
				Service:  node.Service,
				Locality: node.Locality,
				Revision: node.Revision,
			},
			State: node.State,
		}
		assert.Nil(t, cs.ApplyUpdate(update))
	}

	assert.Equal(t, []NodeState{{
		ID:       "remote-1",
		Service:  "foo",
		Locality: "us-east-1-a",
		Revision: "v0.1.0",
		State: map[string]string{
			"addr.bar": "10.26.104.23:1122",
		},
	}}, nodes)
}
