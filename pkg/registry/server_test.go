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
	"context"
	"sort"
	"testing"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"github.com/fuddle-io/fuddle/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestServer_RegisterThenQueryNode(t *testing.T) {
	s := NewServer(NewRegistry())

	registeredNode := testutils.RandomRPCNode()
	_, err := s.Register(context.Background(), &rpc.RegisterRequest{
		Node: registeredNode,
	})
	assert.NoError(t, err)

	resp, err := s.Node(context.Background(), &rpc.NodeRequest{
		NodeId: registeredNode.Id,
	})
	assert.NoError(t, err)

	assert.True(t, proto.Equal(resp.Node, registeredNode))
}

func TestServer_RegisterThenUnregisterNode(t *testing.T) {
	s := NewServer(NewRegistry())

	registeredNode := testutils.RandomRPCNode()
	_, err := s.Register(context.Background(), &rpc.RegisterRequest{
		Node: registeredNode,
	})
	assert.NoError(t, err)

	_, err = s.Unregister(context.Background(), &rpc.UnregisterRequest{
		NodeId: registeredNode.Id,
	})
	assert.NoError(t, err)

	resp, err := s.Node(context.Background(), &rpc.NodeRequest{
		NodeId: registeredNode.Id,
	})
	assert.NoError(t, err)
	assert.Equal(t, rpc.ErrorStatus_NOT_FOUND, resp.Error.Status)
}

func TestServer_NodeNotFound(t *testing.T) {
	s := NewServer(NewRegistry())

	resp, err := s.Node(context.Background(), &rpc.NodeRequest{
		NodeId: "foo",
	})
	assert.NoError(t, err)
	assert.Equal(t, rpc.ErrorStatus_NOT_FOUND, resp.Error.Status)
}

func TestServer_RegisterInvalidNode(t *testing.T) {
	s := NewServer(NewRegistry())

	registeredNode := testutils.RandomRPCNode()
	// Set empty ID.
	registeredNode.Id = ""

	resp, err := s.Register(context.Background(), &rpc.RegisterRequest{
		Node: registeredNode,
	})
	assert.NoError(t, err)
	assert.Equal(t, rpc.ErrorStatus_PROTOCOL, resp.Error.Status)
}

func TestServer_RegisterAlreadyRegister(t *testing.T) {
	s := NewServer(NewRegistry())

	registeredNode := testutils.RandomRPCNode()

	resp, err := s.Register(context.Background(), &rpc.RegisterRequest{
		Node: registeredNode,
	})
	assert.NoError(t, err)
	assert.Nil(t, resp.Error)

	resp, err = s.Register(context.Background(), &rpc.RegisterRequest{
		Node: registeredNode,
	})
	assert.NoError(t, err)
	assert.Equal(t, rpc.ErrorStatus_ALREADY_REGISTERED, resp.Error.Status)
}

func TestServer_UpdateNode(t *testing.T) {
	s := NewServer(NewRegistry())

	registeredNode := testutils.RandomRPCNode()

	regResp, err := s.Register(context.Background(), &rpc.RegisterRequest{
		Node: registeredNode,
	})
	assert.NoError(t, err)
	assert.Nil(t, regResp.Error)

	update := testutils.RandomMetadata()
	updateResp, err := s.UpdateNode(context.Background(), &rpc.UpdateNodeRequest{
		NodeId:   registeredNode.Id,
		Metadata: update,
	})
	assert.NoError(t, err)
	assert.Nil(t, updateResp.Error)

	expectedNode := CopyNode(registeredNode)
	for k, v := range update {
		expectedNode.Metadata[k] = &rpc.VersionedValue{
			Value: v,
		}
	}

	nodeResp, err := s.Node(context.Background(), &rpc.NodeRequest{
		NodeId: registeredNode.Id,
	})
	assert.NoError(t, err)

	assert.True(t, proto.Equal(nodeResp.Node, expectedNode))
}

func TestServer_UpdateNodeNilMetadata(t *testing.T) {
	s := NewServer(NewRegistry())

	registeredNode := testutils.RandomRPCNode()

	regResp, err := s.Register(context.Background(), &rpc.RegisterRequest{
		Node: registeredNode,
	})
	assert.NoError(t, err)
	assert.Nil(t, regResp.Error)

	updateResp, err := s.UpdateNode(context.Background(), &rpc.UpdateNodeRequest{
		NodeId:   registeredNode.Id,
		Metadata: nil,
	})
	assert.NoError(t, err)
	assert.Equal(t, rpc.ErrorStatus_PROTOCOL, updateResp.Error.Status)
}

func TestServer_UpdateNodeNotFound(t *testing.T) {
	s := NewServer(NewRegistry())

	resp, err := s.UpdateNode(context.Background(), &rpc.UpdateNodeRequest{
		NodeId: "foo",
		Metadata: map[string]string{
			"bar": "car",
		},
	})
	assert.NoError(t, err)
	assert.Equal(t, rpc.ErrorStatus_NOT_FOUND, resp.Error.Status)
}

func TestServer_Nodes(t *testing.T) {
	s := NewServer(NewRegistry())

	var nodes []*rpc.Node
	for i := 0; i != 10; i++ {
		registeredNode := testutils.RandomRPCNode()

		_, err := s.Register(context.Background(), &rpc.RegisterRequest{
			Node: registeredNode,
		})
		assert.NoError(t, err)

		nodes = append(nodes, registeredNode)
	}

	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].Id < nodes[j].Id
	})

	resp, err := s.Nodes(context.Background(), &rpc.NodesRequest{
		IncludeMetadata: true,
	})
	assert.NoError(t, err)

	nodesWithMetadata := resp.Nodes
	sort.Slice(nodesWithMetadata, func(i, j int) bool {
		return nodesWithMetadata[i].Id < nodesWithMetadata[j].Id
	})

	assert.Equal(t, 10, len(nodesWithMetadata))
	for i := 0; i != 10; i++ {
		assert.True(t, proto.Equal(nodes[i], nodesWithMetadata[i]))
	}
}

func TestServer_HeartbeatOK(t *testing.T) {
	s := NewServer(NewRegistry())

	registeredNode := testutils.RandomRPCNode()
	_, err := s.Register(context.Background(), &rpc.RegisterRequest{
		Node: registeredNode,
	})
	assert.NoError(t, err)

	resp, err := s.Heartbeat(context.Background(), &rpc.HeartbeatRequest{
		Heartbeat: &rpc.Heartbeat{
			Timestamp: 10,
		},
		Nodes: []string{registeredNode.Id},
	})
	assert.NoError(t, err)

	expected := &rpc.HeartbeatResponse{
		Heartbeat: &rpc.Heartbeat{
			Timestamp: 10,
		},
	}
	assert.True(t, proto.Equal(expected, resp))
}

func TestServer_HeartbeatNotFound(t *testing.T) {
	s := NewServer(NewRegistry())

	resp, err := s.Heartbeat(context.Background(), &rpc.HeartbeatRequest{
		Heartbeat: &rpc.Heartbeat{
			Timestamp: 10,
		},
		Nodes: []string{"not-found"},
	})
	assert.NoError(t, err)

	expected := &rpc.HeartbeatResponse{
		Heartbeat: &rpc.Heartbeat{
			Timestamp: 10,
		},
		Errors: map[string]*rpc.Error{
			"not-found": &rpc.Error{
				Status:      rpc.ErrorStatus_NOT_REGISTERED,
				Description: "not found",
			},
		},
	}
	assert.True(t, proto.Equal(expected, resp))
}
