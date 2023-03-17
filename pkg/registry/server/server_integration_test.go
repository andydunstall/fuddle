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

//go:build integration

package server

import (
	"encoding/json"
	"math/rand"
	"net"
	"net/http"
	"sort"
	"testing"

	fuddle "github.com/fuddle-io/fuddle-go"
	"github.com/fuddle-io/fuddle/pkg/registry/cluster"
	"github.com/fuddle-io/fuddle/pkg/testutils"
	"github.com/google/uuid"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests registering a node. The node should register itself and receive a
// update about the fuddle server joining the cluster.
//
// /v1/api/register
func TestService_RegisterNode(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	serverNode := randomNode()
	c := cluster.NewCluster(serverNode)
	server := NewServer(ln.Addr().String(), c, WithListener(ln))
	require.NoError(t, server.Start())
	defer server.GracefulStop()

	localNode := randomSDKNode()
	registry, err := fuddle.Register([]string{ln.Addr().String()}, localNode)
	require.NoError(t, err)
	defer registry.Unregister()

	// Wait until the registry client learns about two nodes (itself and the
	// fuddle server).
	nodes, err := testutils.WaitForNodes(registry, 2)
	assert.Nil(t, err)

	// Verify the registry now has both the local node and server node.
	expectedNodeIDs := map[string]interface{}{
		localNode.ID:  struct{}{},
		serverNode.ID: struct{}{},
	}
	nodeIDs := nodeIDsSet(nodes)
	assert.Equal(t, expectedNodeIDs, nodeIDs)
}

// Tests a registered node receives updates when other nodes join the cluster.
//
// /v1/api/register
func TestService_RegisterReceiveNodeJoins(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	serverNode := randomNode()
	c := cluster.NewCluster(serverNode)
	server := NewServer(ln.Addr().String(), c, WithListener(ln))
	require.NoError(t, server.Start())
	defer server.GracefulStop()

	addedNodeIDs := make(map[string]interface{})

	localNode := testutils.RandomNode()
	registry, err := fuddle.Register([]string{ln.Addr().String()}, localNode)
	assert.Nil(t, err)
	defer registry.Unregister()

	addedNodeIDs[localNode.ID] = struct{}{}
	addedNodeIDs[serverNode.ID] = struct{}{}

	// Add 10 more nodes to the cluster.
	for i := 0; i != 10; i++ {
		node := testutils.RandomNode()
		r, err := fuddle.Register([]string{ln.Addr().String()}, node)
		assert.Nil(t, err)
		defer r.Unregister()

		addedNodeIDs[node.ID] = struct{}{}
	}

	// Wait until the registry learns about all nodes in the cluster (itself,
	// the sdk server, and the 10 new nodes).
	nodes, err := testutils.WaitForNodes(registry, 12)
	assert.Nil(t, err)

	nodeIDs := nodeIDsSet(nodes)
	assert.Equal(t, addedNodeIDs, nodeIDs)
}

// Tests a registered node receives updates when other nodes leave the cluster.
//
// /v1/api/register
func TestService_RegisterReceiveNodeLeaves(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	serverNode := randomNode()
	c := cluster.NewCluster(serverNode)
	server := NewServer(ln.Addr().String(), c, WithListener(ln))
	require.NoError(t, server.Start())
	defer server.GracefulStop()

	localNode := testutils.RandomNode()
	registry, err := fuddle.Register([]string{ln.Addr().String()}, localNode)
	assert.Nil(t, err)
	defer registry.Unregister()

	// Add 10 more nodes to the cluster.
	var addedRegistries []*fuddle.Registry
	for i := 0; i != 10; i++ {
		node := testutils.RandomNode()
		r, err := fuddle.Register([]string{ln.Addr().String()}, node)
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
		localNode.ID:  struct{}{},
		serverNode.ID: struct{}{},
	}
	nodeIDs := nodeIDsSet(nodes)
	assert.Equal(t, expectedNodeIDs, nodeIDs)
}

// Tests adding 10 random nodes and verifying each node discovers one another.
// Also checks updating on of the node and verifying all nodes discover the
// updated node.
//
// /v1/api/register
func TestService_RegisterClusterDiscovery(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	serverNode := randomNode()
	c := cluster.NewCluster(serverNode)
	server := NewServer(ln.Addr().String(), c, WithListener(ln))
	require.NoError(t, server.Start())
	defer server.GracefulStop()

	var addedNodes []fuddle.Node
	var addedRegistries []*fuddle.Registry
	for i := 0; i != 10; i++ {
		node := testutils.RandomNode()
		r, err := fuddle.Register([]string{ln.Addr().String()}, node)
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

// Tests /api/v1/cluster returns the correct cluster state.
func TestService_Cluster(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	var nodes []cluster.Node

	local := testutils.RandomRegistryNode()
	nodes = append(nodes, local)
	c := cluster.NewCluster(local)
	// Add 5 more random nodes.
	for i := 0; i != 5; i++ {
		node := testutils.RandomRegistryNode()
		nodes = append(nodes, node)
		c.ApplyUpdate(&cluster.NodeUpdate{
			ID:         node.ID,
			UpdateType: cluster.UpdateTypeRegister,
			Attributes: &cluster.NodeAttributes{
				Service:  node.Service,
				Locality: node.Locality,
				Created:  node.Created,
				Revision: node.Revision,
			},
			Metadata: node.Metadata,
		})
	}

	service := NewServer(ln.Addr().String(), c, WithListener(ln))
	require.NoError(t, service.Start())
	defer service.GracefulStop()

	resp, err := http.Get("http://" + ln.Addr().String() + "/api/v1/cluster")
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)

	var recvNodes []cluster.Node
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&recvNodes))

	// Sort the nodes to make comparison easier.
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].ID < nodes[j].ID
	})
	sort.Slice(recvNodes, func(i, j int) bool {
		return recvNodes[i].ID < recvNodes[j].ID
	})

	assert.Equal(t, nodes, recvNodes)
}

// Tests /api/v1/node/{id} returns the correct node state.
func TestService_Node(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	local := testutils.RandomRegistryNode()
	c := cluster.NewCluster(local)
	service := NewServer(ln.Addr().String(), c, WithListener(ln))
	require.NoError(t, service.Start())
	defer service.GracefulStop()

	resp, err := http.Get("http://" + ln.Addr().String() + "/api/v1/node/" + local.ID)
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)

	var node cluster.Node
	assert.NoError(t, json.NewDecoder(resp.Body).Decode(&node))

	assert.Equal(t, node, local)
}

// Tests /api/v1/node/{id} returns 404 when a node is not found.
func TestService_NodeNotFound(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	local := testutils.RandomRegistryNode()
	c := cluster.NewCluster(local)
	service := NewServer(ln.Addr().String(), c, WithListener(ln))
	require.NoError(t, service.Start())
	defer service.GracefulStop()

	resp, err := http.Get("http://" + ln.Addr().String() + "/api/v1/node/notfound")
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 404, resp.StatusCode)
}

// Tests /metrics returns 200 when metrics are registered.
func TestService_Metrics(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	promRegistry := prometheus.NewRegistry()

	c := cluster.NewCluster(testutils.RandomRegistryNode())
	service := NewServer(
		ln.Addr().String(),
		c,
		WithListener(ln),
		WithPromRegistry(promRegistry),
	)
	require.NoError(t, service.Start())
	defer service.GracefulStop()

	resp, err := http.Get("http://" + ln.Addr().String() + "/metrics")
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 200, resp.StatusCode)
}

// Tests /metrics returns 404 when metrics are not registered.
func TestService_MetricsNotRegistered(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	c := cluster.NewCluster(testutils.RandomRegistryNode())
	service := NewServer(
		ln.Addr().String(), c, WithListener(ln),
	)
	require.NoError(t, service.Start())
	defer service.GracefulStop()

	resp, err := http.Get("http://" + ln.Addr().String() + "/metrics")
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 404, resp.StatusCode)
}

// randomNode returns a node with random attributes and metadata.
func randomNode() cluster.Node {
	return cluster.Node{
		ID:       uuid.New().String(),
		Service:  uuid.New().String(),
		Locality: uuid.New().String(),
		Created:  rand.Int63(),
		Revision: uuid.New().String(),
		Metadata: map[string]string{
			uuid.New().String(): uuid.New().String(),
			uuid.New().String(): uuid.New().String(),
			uuid.New().String(): uuid.New().String(),
			uuid.New().String(): uuid.New().String(),
			uuid.New().String(): uuid.New().String(),
		},
	}
}

// randomSDKNode returns a node with random attributes and metadata.
func randomSDKNode() fuddle.Node {
	return fuddle.Node{
		ID:       uuid.New().String(),
		Service:  uuid.New().String(),
		Locality: uuid.New().String(),
		Created:  rand.Int63(),
		Revision: uuid.New().String(),
		State: map[string]string{
			uuid.New().String(): uuid.New().String(),
			uuid.New().String(): uuid.New().String(),
			uuid.New().String(): uuid.New().String(),
			uuid.New().String(): uuid.New().String(),
			uuid.New().String(): uuid.New().String(),
		},
	}
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
