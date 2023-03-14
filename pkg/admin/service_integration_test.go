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

package admin

import (
	"encoding/json"
	"net"
	"net/http"
	"sort"
	"testing"

	"github.com/fuddle-io/fuddle/pkg/config"
	"github.com/fuddle-io/fuddle/pkg/registry/cluster"
	"github.com/fuddle-io/fuddle/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

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

	service := NewService(c, &config.Config{}, WithListener(ln))
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
	service := NewService(c, &config.Config{}, WithListener(ln))
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
	service := NewService(c, &config.Config{}, WithListener(ln))
	require.NoError(t, service.Start())
	defer service.GracefulStop()

	resp, err := http.Get("http://" + ln.Addr().String() + "/api/v1/node/notfound")
	assert.NoError(t, err)
	defer resp.Body.Close()
	assert.Equal(t, 404, resp.StatusCode)
}
