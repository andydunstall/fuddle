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

	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"github.com/fuddle-io/fuddle/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestRegistry_RegisterThenQueryNode(t *testing.T) {
	r := NewRegistry()

	registeredNode := testutils.RandomRPCNode()
	assert.NoError(t, r.Register(registeredNode))

	n, err := r.Node(registeredNode.Id)
	assert.NoError(t, err)
	assert.True(t, proto.Equal(n, registeredNode))
}

func TestRegistry_RegisterThenUnregister(t *testing.T) {
	r := NewRegistry()

	registeredNode := testutils.RandomRPCNode()
	assert.NoError(t, r.Register(registeredNode))
	assert.NoError(t, r.Unregister(registeredNode.Id))

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
			assert.Equal(t, ErrInvalidUpdate, r.Register(tt.Node))
		})
	}
}

func TestRegistry_RegisterAlreadyRegister(t *testing.T) {
	r := NewRegistry()

	registeredNode := testutils.RandomRPCNode()
	assert.NoError(t, r.Register(registeredNode))

	// Register the same node again which should fail.
	assert.Equal(t, ErrAlreadyRegistered, r.Register(registeredNode))
}

func TestRegistry_UpdateNode(t *testing.T) {
	r := NewRegistry()

	registeredNode := testutils.RandomRPCNode()
	assert.NoError(t, r.Register(registeredNode))

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
	assert.NoError(t, r.Register(registeredNode))

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
		assert.NoError(t, r.Register(registeredNode))

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
