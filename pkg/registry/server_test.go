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

	"github.com/andydunstall/fuddle/pkg/rpc"
	"github.com/stretchr/testify/assert"
	"go.uber.org/zap"
)

func TestServer_RegisterAndUnregisterNode(t *testing.T) {
	m := NewNodeMap()
	s := NewServer(m, zap.NewNop())

	_, err := s.Register(context.TODO(), &rpc.RegisterRequest{
		NodeId: "node-1",
	})
	assert.Nil(t, err)
	_, err = s.Register(context.TODO(), &rpc.RegisterRequest{
		NodeId: "node-2",
	})
	assert.Nil(t, err)

	nodeIDs := m.NodeIDs()
	// Sort to make comparison easier.
	sort.Strings(nodeIDs)
	assert.Equal(t, []string{"node-1", "node-2"}, nodeIDs)

	_, err = s.Unregister(context.TODO(), &rpc.RegisterRequest{
		NodeId: "node-1",
	})
	assert.Nil(t, err)

	assert.Equal(t, []string{"node-2"}, m.NodeIDs())
}
