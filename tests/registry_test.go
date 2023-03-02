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
	"context"
	"sort"
	"testing"

	"github.com/andydunstall/fuddle/pkg/client"
	"github.com/andydunstall/fuddle/pkg/rpc"
	"github.com/andydunstall/fuddle/pkg/server"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestRegistry_RegisterAndUnregisterNode(t *testing.T) {
	conf := testConfig()
	server := server.NewServer(conf, zap.NewNop())
	assert.Nil(t, server.Start())
	defer server.GracefulStop()

	registry, err := client.ConnectRegistry(conf.AdvAddr)
	assert.Nil(t, err)

	assert.Nil(t, registry.Register(context.TODO(), &rpc.NodeState{Id: "node-1"}))
	assert.Nil(t, registry.Register(context.TODO(), &rpc.NodeState{Id: "node-2"}))

	nodes, err := registry.Nodes(context.TODO())
	assert.Nil(t, err)
	nodeIDs := []string{}
	for _, node := range nodes {
		nodeIDs = append(nodeIDs, node.Id)
	}
	// Sort to make comparison easier.
	sort.Strings(nodeIDs)
	assert.Equal(t, []string{"node-1", "node-2"}, nodeIDs)

	assert.Nil(t, registry.Unregister(context.TODO(), "node-1"))

	nodes, err = registry.Nodes(context.TODO())
	assert.Nil(t, err)
	nodeIDs = []string{}
	for _, node := range nodes {
		nodeIDs = append(nodeIDs, node.Id)
	}
	assert.Equal(t, []string{"node-2"}, nodeIDs)
}
