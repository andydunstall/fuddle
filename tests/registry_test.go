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
	"fmt"
	"sort"
	"testing"
	"time"

	"github.com/andydunstall/fuddle/pkg/rpc"
	fuddle "github.com/andydunstall/fuddle/pkg/sdk"
	"github.com/andydunstall/fuddle/pkg/server"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

// Tests registering a node checks the node is in the clients node map.
func TestRegistry_RegisterNode(t *testing.T) {
	conf := testConfig()
	server := server.NewServer(conf, zap.NewNop())
	assert.Nil(t, server.Start())
	defer server.GracefulStop()

	client, err := fuddle.Register(conf.AdvAddr, fuddle.Attributes{
		ID: "node-1",
	}, make(map[string]string), zap.NewNop())
	assert.Nil(t, err)
	defer func() {
		assert.Nil(t, client.Unregister())
	}()

	// Check when subscribing with rewind we receive ourselves.
	updates := make(chan *rpc.NodeUpdate, 1)
	client.Subscribe(true, func(update *rpc.NodeUpdate) {
		updates <- update
	})
	update := waitWithTimeout(updates)
	assert.Equal(t, "node-1", update.NodeId)
	assert.Equal(t, rpc.UpdateType_NODE_JOIN, update.UpdateType)

	assert.Nil(t, err)
}

func TestRegistry_SubscribeToClusterUpdates(t *testing.T) {
	conf := testConfig()
	server := server.NewServer(conf, zap.NewNop())
	assert.Nil(t, server.Start())
	defer server.GracefulStop()

	client, err := fuddle.Register(conf.AdvAddr, fuddle.Attributes{
		ID: "local-node",
	}, make(map[string]string), zap.NewNop())
	assert.Nil(t, err)
	defer func() {
		assert.Nil(t, client.Unregister())
	}()

	updates := make(chan *rpc.NodeUpdate, 64)
	client.Subscribe(false, func(update *rpc.NodeUpdate) {
		updates <- update
	})

	// Add more nodes to the registry, and check the first node receives
	// updates for each.

	var ids []string
	var clients []*fuddle.Fuddle
	for i := 0; i != 5; i++ {
		id := fmt.Sprintf("node-%d", i)
		ids = append(ids, id)
		client, err := fuddle.Register(conf.AdvAddr, fuddle.Attributes{
			ID: id,
		}, make(map[string]string), zap.NewNop())
		clients = append(clients, client)
		assert.Nil(t, err)
	}

	for _, id := range ids {
		update := waitWithTimeout(updates)
		assert.Equal(t, id, update.NodeId)
		assert.Equal(t, rpc.UpdateType_NODE_JOIN, update.UpdateType)
	}

	// Update each node.

	for _, client := range clients {
		assert.Nil(t, client.Update("foo", "bar"))
	}

	var updatedIDs []string
	for _, id := range ids {
		update := waitWithTimeout(updates)
		assert.Equal(t, rpc.UpdateType_NODE_UPDATE, update.UpdateType)
		updatedIDs = append(updatedIDs, id)
	}
	sort.Strings(updatedIDs)
	assert.Equal(t, ids, updatedIDs)

	// Remove each of the nodes in the registry, and check the first node
	// receives leave updates for each.

	for _, client := range clients {
		assert.Nil(t, client.Unregister())
	}

	var leftIDs []string
	for _, id := range ids {
		update := waitWithTimeout(updates)
		assert.Equal(t, rpc.UpdateType_NODE_LEAVE, update.UpdateType)
		leftIDs = append(leftIDs, id)
	}
	sort.Strings(leftIDs)
	assert.Equal(t, ids, leftIDs)
}

func waitWithTimeout(c chan *rpc.NodeUpdate) *rpc.NodeUpdate {
	select {
	case update := <-c:
		return update
	case <-time.After(time.Second):
		return nil
	}
}
