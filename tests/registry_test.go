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
	"time"

	fuddle "github.com/andydunstall/fuddle/pkg/sdkv2"
	"github.com/andydunstall/fuddle/pkg/testutils"
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

	// Subscribe and wait until the registry client knows about two nodes
	// (itself and the fuddle server).
	recvCh := make(chan interface{}, 1)
	unsubscribe := registry.Subscribe(func(nodes []fuddle.NodeState) {
		if len(nodes) == 2 {
			close(recvCh)
		}
	})
	defer unsubscribe()

	assert.Nil(t, testutils.WaitWithTimeout(recvCh, time.Second))

	// Verify the registry now has both the local node and server node.
	expectedNodeIDs := map[string]interface{}{
		localNode.ID: struct{}{},
		server.ID():  struct{}{},
	}
	nodeIDs := nodeIDsSet(registry.Nodes())
	assert.Equal(t, expectedNodeIDs, nodeIDs)
}

func nodeIDsSet(nodes []fuddle.NodeState) map[string]interface{} {
	ids := make(map[string]interface{})
	for _, node := range nodes {
		ids[node.ID] = struct{}{}
	}
	return ids
}
