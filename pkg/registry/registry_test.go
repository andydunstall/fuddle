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
	"time"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"github.com/fuddle-io/fuddle/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestRegistry_RegisterThenQueryNode(t *testing.T) {
	r := NewRegistry()

	registeredNode := testutils.RandomRPCNode()
	assert.NoError(t, r.Register(registeredNode, time.Now()))

	n, err := r.Node(registeredNode.Id)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(n, registeredNode))
}

func TestRegistry_RegisterThenUnregister(t *testing.T) {
	r := NewRegistry()

	registeredNode := testutils.RandomRPCNode()
	assert.NoError(t, r.Register(registeredNode, time.Now()))
	assert.True(t, r.Unregister(registeredNode.Id))

	_, err := r.Node(registeredNode.Id)
	assert.Equal(t, ErrNotFound, err)
}

func TestRegistry_NodeNotFound(t *testing.T) {
	r := NewRegistry()

	_, err := r.Node("foo")
	assert.Equal(t, ErrNotFound, err)
}

func TestRegistry_RegisterInvalidUpdate(t *testing.T) {
	r := NewRegistry()

	nodeMissingID := testutils.RandomRPCNode()
	nodeMissingID.Id = ""

	nodeMissingAttrs := testutils.RandomRPCNode()
	nodeMissingAttrs.Attributes = nil

	nodeMissingMetadata := testutils.RandomRPCNode()
	nodeMissingMetadata.Metadata = nil

	tests := []struct {
		Name string
		Node *rpc.Node
	}{
		{
			Name: "missing id",
			Node: nodeMissingID,
		},
		{
			Name: "missing attrs",
			Node: nodeMissingAttrs,
		},
		{
			Name: "missing metadata",
			Node: nodeMissingMetadata,
		},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			assert.Equal(t, ErrInvalidUpdate, r.Register(tt.Node, time.Now()))
		})
	}
}

func TestRegistry_RegisterAlreadyRegister(t *testing.T) {
	r := NewRegistry()

	registeredNode := testutils.RandomRPCNode()
	assert.NoError(t, r.Register(registeredNode, time.Now()))

	// Register the same node again which should fail.
	assert.Equal(t, ErrAlreadyRegistered, r.Register(registeredNode, time.Now()))
}

func TestRegistry_UpdateNode(t *testing.T) {
	r := NewRegistry()

	registeredNode := testutils.RandomRPCNode()
	assert.NoError(t, r.Register(registeredNode, time.Now()))

	update := testutils.RandomMetadata()
	assert.NoError(t, r.UpdateNode(registeredNode.Id, update))

	expectedNode := CopyNode(registeredNode)
	for k, v := range update {
		expectedNode.Metadata[k] = &rpc.VersionedValue{
			Value: v,
		}
	}

	n, err := r.Node(registeredNode.Id)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(n, expectedNode))
}

func TestRegistry_UpdateNodeNilMetadata(t *testing.T) {
	r := NewRegistry()

	registeredNode := testutils.RandomRPCNode()
	assert.NoError(t, r.Register(registeredNode, time.Now()))

	assert.Equal(t, ErrInvalidUpdate, r.UpdateNode(registeredNode.Id, nil))
}

func TestRegistry_UpdateNodeNotFound(t *testing.T) {
	r := NewRegistry()

	assert.Equal(t, ErrNotFound, r.UpdateNode("foo", map[string]string{"a": "b"}))
}

func TestRegistry_Nodes(t *testing.T) {
	r := NewRegistry()

	var nodes []*rpc.Node
	for i := 0; i != 10; i++ {
		registeredNode := testutils.RandomRPCNode()
		assert.NoError(t, r.Register(registeredNode, time.Now()))

		nodes = append(nodes, registeredNode)
	}

	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Id < nodes[j].Id
	})

	nodesWithMetadata := r.Nodes(true)
	sort.Slice(nodesWithMetadata, func(i, j int) bool {
		return nodesWithMetadata[i].Id < nodesWithMetadata[j].Id
	})

	assert.Equal(t, 10, len(nodesWithMetadata))
	for i := 0; i != 10; i++ {
		assert.True(t, proto.Equal(nodes[i], nodesWithMetadata[i]))
	}
}

// Tests subscribign to register updates.
func TestRegistry_SubscribeToRegister(t *testing.T) {
	r := NewRegistry()

	var receivedUpdates []*rpc.NodeUpdate
	unsubscribe := r.Subscribe(func(update *rpc.NodeUpdate) {
		receivedUpdates = append(receivedUpdates, update)
	})
	defer unsubscribe()

	var expectedUpdates []*rpc.NodeUpdate
	for i := 0; i != 10; i++ {
		node := testutils.RandomRPCNode()
		assert.NoError(t, r.Register(node, time.Now()))

		expectedUpdates = append(expectedUpdates, &rpc.NodeUpdate{
			NodeId:     node.Id,
			UpdateType: rpc.NodeUpdateType_REGISTER,
			Attributes: CopyAttributes(node.Attributes),
			Metadata:   CopyMetadata(node.Metadata),
		})
	}
	sort.Slice(expectedUpdates, func(i, j int) bool {
		return expectedUpdates[i].NodeId < expectedUpdates[j].NodeId
	})

	sort.Slice(receivedUpdates, func(i, j int) bool {
		return receivedUpdates[i].NodeId < receivedUpdates[j].NodeId
	})

	assert.Equal(t, 10, len(receivedUpdates))
	for i := 0; i != 10; i++ {
		assert.True(t, proto.Equal(expectedUpdates[i], receivedUpdates[i]))
	}
}

// Tests subscribing to unregister updates.
func TestRegistry_SubscribeToUnregister(t *testing.T) {
	r := NewRegistry()

	var receivedUpdates []*rpc.NodeUpdate
	unsubscribe := r.Subscribe(func(update *rpc.NodeUpdate) {
		receivedUpdates = append(receivedUpdates, update)
	})
	defer unsubscribe()

	var registeredIDs []string
	for i := 0; i != 10; i++ {
		node := testutils.RandomRPCNode()
		assert.NoError(t, r.Register(node, time.Now()))
		registeredIDs = append(registeredIDs, node.Id)
	}

	var expectedUpdates []*rpc.NodeUpdate
	for _, id := range registeredIDs {
		assert.True(t, r.Unregister(id))

		expectedUpdates = append(expectedUpdates, &rpc.NodeUpdate{
			NodeId:     id,
			UpdateType: rpc.NodeUpdateType_UNREGISTER,
		})
	}
	sort.Slice(expectedUpdates, func(i, j int) bool {
		return expectedUpdates[i].NodeId < expectedUpdates[j].NodeId
	})

	// Discard the first 10 register updates.
	receivedUpdates = receivedUpdates[10:]
	sort.Slice(receivedUpdates, func(i, j int) bool {
		return receivedUpdates[i].NodeId < receivedUpdates[j].NodeId
	})

	assert.Equal(t, 10, len(receivedUpdates))
	for i := 0; i != 10; i++ {
		assert.True(t, proto.Equal(expectedUpdates[i], receivedUpdates[i]))
	}
}

// Tests subscribing to metadata updates.
func TestRegistry_SubscribeToMetadata(t *testing.T) {
	r := NewRegistry()

	var receivedUpdates []*rpc.NodeUpdate
	unsubscribe := r.Subscribe(func(update *rpc.NodeUpdate) {
		receivedUpdates = append(receivedUpdates, update)
	})
	defer unsubscribe()

	var registeredIDs []string
	for i := 0; i != 10; i++ {
		node := testutils.RandomRPCNode()
		assert.NoError(t, r.Register(node, time.Now()))
		registeredIDs = append(registeredIDs, node.Id)
	}

	var expectedUpdates []*rpc.NodeUpdate
	for _, id := range registeredIDs {
		metadata := testutils.RandomMetadata()
		versionedMetadataUpdate := make(map[string]*rpc.VersionedValue)
		for k, v := range metadata {
			versionedMetadataUpdate[k] = &rpc.VersionedValue{
				Value: v,
			}
		}

		assert.NoError(t, r.UpdateNode(id, metadata))

		expectedUpdates = append(expectedUpdates, &rpc.NodeUpdate{
			NodeId:     id,
			UpdateType: rpc.NodeUpdateType_METADATA,
			Metadata:   versionedMetadataUpdate,
		})
	}
	sort.Slice(expectedUpdates, func(i, j int) bool {
		return expectedUpdates[i].NodeId < expectedUpdates[j].NodeId
	})

	// Discard the first 10 register updates.
	receivedUpdates = receivedUpdates[10:]
	sort.Slice(receivedUpdates, func(i, j int) bool {
		return receivedUpdates[i].NodeId < receivedUpdates[j].NodeId
	})

	assert.Equal(t, 10, len(receivedUpdates))
	for i := 0; i != 10; i++ {
		assert.True(t, proto.Equal(expectedUpdates[i], receivedUpdates[i]))
	}
}

// Tests we receive register updates for all nodes in the registry when
// subscribing.
func TestRegistry_SubscribeBootstrap(t *testing.T) {
	r := NewRegistry()

	var expectedUpdates []*rpc.NodeUpdate
	for i := 0; i != 10; i++ {
		node := testutils.RandomRPCNode()
		assert.NoError(t, r.Register(node, time.Now()))

		expectedUpdates = append(expectedUpdates, &rpc.NodeUpdate{
			NodeId:     node.Id,
			UpdateType: rpc.NodeUpdateType_REGISTER,
			Attributes: CopyAttributes(node.Attributes),
			Metadata:   CopyMetadata(node.Metadata),
		})
	}
	sort.Slice(expectedUpdates, func(i, j int) bool {
		return expectedUpdates[i].NodeId < expectedUpdates[j].NodeId
	})

	var receivedUpdates []*rpc.NodeUpdate
	unsubscribe := r.Subscribe(func(update *rpc.NodeUpdate) {
		receivedUpdates = append(receivedUpdates, update)
	})
	defer unsubscribe()
	sort.Slice(receivedUpdates, func(i, j int) bool {
		return receivedUpdates[i].NodeId < receivedUpdates[j].NodeId
	})

	assert.Equal(t, 10, len(receivedUpdates))
	for i := 0; i != 10; i++ {
		assert.True(t, proto.Equal(expectedUpdates[i], receivedUpdates[i]))
	}
}

func TestRegistry_UnregisterDownNodes(t *testing.T) {
	r := NewRegistry()

	// Register two nodes, both should be up.

	n1 := testutils.RandomRPCNode()
	assert.NoError(t, r.Register(n1, time.Unix(1, 0)))

	n2 := testutils.RandomRPCNode()
	assert.NoError(t, r.Register(n2, time.Unix(2, 0)))

	assert.Equal(t, []string(nil), r.UnregisterDownNodes(
		time.Unix(3, 0), time.Second*10,
	))

	// Mark contact for only the first node. The second should be unregistered
	// when checking.

	assert.NoError(t, r.MarkContact(n1.Id, time.Unix(10, 0)))

	// Expect the second node to be unregistered.
	assert.Equal(t, []string{n2.Id}, r.UnregisterDownNodes(
		time.Unix(15, 0), time.Second*10,
	))

	_, err := r.Node(n2.Id)
	assert.Equal(t, ErrNotFound, err)
}

func TestRegistry_MarkContactNotFound(t *testing.T) {
	r := NewRegistry()

	assert.Equal(t, ErrNotFound, r.MarkContact("not found", time.Now()))
}
