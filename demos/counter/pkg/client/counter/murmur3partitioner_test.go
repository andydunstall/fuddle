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

package counter

import (
	"testing"
	"time"

	"github.com/stretchr/testify/assert"
)

func TestMurmur3Partitioner_Locate(t *testing.T) {
	p := NewMurmur3Partitioner()

	p.SetNodes(map[string]string{
		"node-1": "192.168.1.1",
		"node-2": "192.168.1.2",
		"node-3": "192.168.1.3",
	})

	// Lookup foo.
	addr, _, ok := p.Locate("foo", nil)
	assert.True(t, ok)
	assert.Equal(t, NodeAddr{"node-2", "192.168.1.2"}, addr)

	// Remove the node foo was located and verify it has moved.
	p.SetNodes(map[string]string{
		"node-1": "192.168.1.1",
		"node-3": "192.168.1.3",
	})
	addr, _, ok = p.Locate("foo", nil)
	assert.True(t, ok)
	assert.Equal(t, NodeAddr{"node-3", "192.168.1.3"}, addr)

	// Add the original node back and verify foo moves back to that node.
	p.SetNodes(map[string]string{
		"node-1": "192.168.1.1",
		"node-2": "192.168.1.2",
		"node-3": "192.168.1.3",
	})
	addr, _, ok = p.Locate("foo", nil)
	assert.True(t, ok)
	assert.Equal(t, NodeAddr{"node-2", "192.168.1.2"}, addr)
}

func TestMurmur3Partitioner_LocateWithOnRelocate(t *testing.T) {
	p := NewMurmur3Partitioner()

	relocates := make(chan NodeAddr, 1)

	p.SetNodes(map[string]string{
		"node-1": "192.168.1.1",
		"node-2": "192.168.1.2",
		"node-3": "192.168.1.3",
	})

	// Lookup foo.
	addr, unregister, ok := p.Locate("foo", func(addr NodeAddr, ok bool) {
		assert.True(t, ok)
		relocates <- addr
	})
	assert.True(t, ok)
	assert.Equal(t, NodeAddr{"node-2", "192.168.1.2"}, addr)
	defer unregister()

	// Remove the node foo was located and verify it has moved.
	p.SetNodes(map[string]string{
		"node-1": "192.168.1.1",
		"node-3": "192.168.1.3",
	})
	select {
	case addr = <-relocates:
		assert.Equal(t, NodeAddr{"node-3", "192.168.1.3"}, addr)
	case <-time.After(time.Second):
		t.Error("timeout")
	}

	// Add the original node back and verify foo moves back to that node.
	p.SetNodes(map[string]string{
		"node-1": "192.168.1.1",
		"node-2": "192.168.1.2",
		"node-3": "192.168.1.3",
	})
	select {
	case addr = <-relocates:
		assert.Equal(t, NodeAddr{"node-2", "192.168.1.2"}, addr)
	case <-time.After(time.Second):
		t.Error("timeout")
	}
}

func TestMurmur3Partitioner_LocateNoNodes(t *testing.T) {
	p := NewMurmur3Partitioner()
	_, _, ok := p.Locate("foo", nil)
	assert.False(t, ok)
}
