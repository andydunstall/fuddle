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

func TestNodeMap_RegisterAndUnregisterNode(t *testing.T) {
	m := NewNodeMap()
	assert.Equal(t, []string{}, m.NodeIDs())

	m.Register(&rpc.RegisterRequest{NodeId: "node-1"})
	m.Register(&rpc.RegisterRequest{NodeId: "node-2"})

	nodeIDs := m.NodeIDs()
	// Sort to make comparison easier.
	sort.Strings(nodeIDs)
	assert.Equal(t, []string{"node-1", "node-2"}, nodeIDs)

	m.Unregister("node-1")

	assert.Equal(t, []string{"node-2"}, m.NodeIDs())
}

func TestNodeMap_SubscribeToUpdates(t *testing.T) {
	m := NewNodeMap()

	notifiedCount := 0

	// Subscribe and check the callback is called once for each update.
	m.Subscribe("sub", func() {
		notifiedCount++
	})

	m.Register(&rpc.RegisterRequest{NodeId: "node-1"})
	m.Register(&rpc.RegisterRequest{NodeId: "node-2"})
	m.Unregister("node-2")
	assert.Equal(t, 3, notifiedCount)

	// Unsubscribe and check the callback is no longer called.
	m.Unsubscribe("sub")

	m.Register(&rpc.RegisterRequest{NodeId: "node-3"})
	m.Register(&rpc.RegisterRequest{NodeId: "node-4"})
	m.Unregister("node-2")
	assert.Equal(t, 3, notifiedCount)
}
