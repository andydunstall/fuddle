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

package tests

import (
	"testing"

	fuddle "github.com/fuddle-io/fuddle/pkg/sdk"
	"github.com/fuddle-io/fuddle/pkg/testutils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// Tests registering a node. The node should register itself and receive a
// update about the fuddle server joining the cluster.
func TestRegistry_RegisterNode(t *testing.T) {
	server, err := testutils.StartServer()
	assert.Nil(t, err)
	defer server.GracefulStop()

	localNode := testutils.RandomNode()
	registry, err := fuddle.Register([]string{server.RPCAddr()}, localNode)
	assert.Nil(t, err)
	defer registry.Unregister()

	// Wait until the registry client learns about two nodes (itself and the
	// fuddle server).
	nodes, err := testutils.WaitForNodes(registry, 2)
	assert.Nil(t, err)

	// Verify the registry now has both the local node and server node.
	expectedNodeIDs := map[string]interface{}{
		localNode.ID: struct{}{},
		server.ID():  struct{}{},
	}
	nodeIDs := nodeIDsSet(nodes)
	assert.Equal(t, expectedNodeIDs, nodeIDs)
}

// Tests a registered node receives updates when other nodes join the cluster.
func TestRegistry_ReceiveNodeJoins(t *testing.T) {
	server, err := testutils.StartServer()
	assert.Nil(t, err)
	defer server.GracefulStop()

	addedNodeIDs := make(map[string]interface{})

	localNode := testutils.RandomNode()
	registry, err := fuddle.Register([]string{server.RPCAddr()}, localNode)
	assert.Nil(t, err)
	defer registry.Unregister()

	addedNodeIDs[localNode.ID] = struct{}{}
	addedNodeIDs[server.ID()] = struct{}{}

	// Add 10 more nodes to the cluster.
	for i := 0; i != 10; i++ {
		node := testutils.RandomNode()
		r, err := fuddle.Register([]string{server.RPCAddr()}, node)
		assert.Nil(t, err)
		defer r.Unregister()

		addedNodeIDs[node.ID] = struct{}{}
	}

	// Wait until the registry learns about all nodes in the cluster (itself,
	// the fuddle server, and the 10 new nodes).
	nodes, err := testutils.WaitForNodes(registry, 12)
	assert.Nil(t, err)

	nodeIDs := nodeIDsSet(nodes)
	assert.Equal(t, addedNodeIDs, nodeIDs)
}

// Tests a registered node receives updates when other nodes leave the cluster.
func TestRegistry_ReceiveNodeLeaves(t *testing.T) {
	server, err := testutils.StartServer()
	assert.Nil(t, err)
	defer server.GracefulStop()

	localNode := testutils.RandomNode()
	registry, err := fuddle.Register([]string{server.RPCAddr()}, localNode)
	assert.Nil(t, err)
	defer registry.Unregister()

	// Add 10 more nodes to the cluster.
	var addedRegistries []*fuddle.Registry
	for i := 0; i != 10; i++ {
		node := testutils.RandomNode()
		r, err := fuddle.Register([]string{server.RPCAddr()}, node)
		assert.Nil(t, err)

		addedRegistries = append(addedRegistries, r)
	}

	// Wait until the registry learns about all nodes in the cluster (itself,
	// the fuddle server, and the 10 new nodes).
	_, err = testutils.WaitForNodes(registry, 12)
	assert.Nil(t, err)

	// Unregister the nodes again.
	for _, r := range addedRegistries {
		assert.Nil(t, r.Unregister())
	}

	// Wait until the registry learns about the nodes leaving the cluster.
	nodes, err := testutils.WaitForNodes(registry, 2)
	assert.Nil(t, err)

	// Verify the registry now has only the local node and server node.
	expectedNodeIDs := map[string]interface{}{
		localNode.ID: struct{}{},
		server.ID():  struct{}{},
	}
	nodeIDs := nodeIDsSet(nodes)
	assert.Equal(t, expectedNodeIDs, nodeIDs)
}

// Tests adding 10 random nodes and verifying each node discovers one another.
// Also checks updating on of the node and verifying all nodes discover the
// updated node.
func TestRegistry_ClusterDiscovery(t *testing.T) {
	server, err := testutils.StartServer()
	assert.Nil(t, err)
	defer server.GracefulStop()

	var addedNodes []fuddle.Node
	var addedRegistries []*fuddle.Registry
	for i := 0; i != 10; i++ {
		node := testutils.RandomNode()
		r, err := fuddle.Register([]string{server.RPCAddr()}, node)
		assert.Nil(t, err)

		addedRegistries = append(addedRegistries, r)
		addedNodes = append(addedNodes, node)
		defer r.Unregister()
	}

	// Wait for all nodes to discovery each other and have the same cluster state.
	var nodes []fuddle.Node
	for _, r := range addedRegistries {
		discoveredNodes, err := testutils.WaitForNodes(r, 11)
		assert.Nil(t, err)

		if nodes != nil {
			assert.Equal(t, nodesMap(nodes), nodesMap(discoveredNodes))
		}
		nodes = discoveredNodes
	}

	// Update the state of the first node in the cluster and wait for all
	// nodes to discover the update.
	updatedNode := addedNodes[0]
	updatedNode.State["foo"] = uuid.New().String()
	assert.Nil(t, addedRegistries[0].UpdateLocalState(updatedNode.State))

	for _, r := range addedRegistries {
		assert.Nil(t, testutils.WaitForNode(r, updatedNode))
	}
}

// Tests the registry client tries all given addresses to connect.
func TestRegistry_AttemptMultipleAddresses(t *testing.T) {
	server, err := testutils.StartServer()
	assert.Nil(t, err)
	defer server.GracefulStop()

	// Seed with a list of addresses where only one is valid.
	addrs := []string{
		"invalid.1:123",
		"invalid.2:123",
		"invalid.3:123",
		server.RPCAddr(),
	}
	registry, err := fuddle.Register(addrs, testutils.RandomNode())
	assert.Nil(t, err)
	defer registry.Unregister()
}

func nodeIDsSet(nodes []fuddle.Node) map[string]interface{} {
	ids := make(map[string]interface{})
	for _, node := range nodes {
		ids[node.ID] = struct{}{}
	}
	return ids
}

func nodesMap(nodes []fuddle.Node) map[string]fuddle.Node {
	nodesMap := make(map[string]fuddle.Node)
	for _, node := range nodes {
		nodesMap[node.ID] = node
	}
	return nodesMap
}
