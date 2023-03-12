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

	"github.com/fuddle-io/fuddle/demos/counter/pkg/client/counter"
	"github.com/fuddle-io/fuddle/demos/counter/pkg/testutils/cluster"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests registering increments the count and receives update when the
// count changes.
func TestCounter_Register(t *testing.T) {
	c, err := cluster.NewCluster(
		cluster.WithFuddleNodes(1),
		cluster.WithCounterNodes(3),
	)
	require.NoError(t, err)
	defer c.Shutdown()

	partitioner := counter.NewMurmur3Partitioner()
	partitioner.SetNodes(c.CounterNodes())

	client := counter.NewClient(partitioner)
	defer client.Close()

	// Register and subscribe to updates.
	updates := make(chan uint64, 1)
	unsubscribe, err := client.Register("foo", func(c uint64) {
		updates <- c
	})
	require.NoError(t, err)
	defer func() {
		assert.Nil(t, unsubscribe())
	}()

	// Expect to receive an update that the user was registered.
	assert.Equal(t, uint64(1), waitTimeout(updates, t))

	// Register 15 more users with the same ID, split across multiple client
	// connections.
	var unregister []func() error
	for i := 0; i != 5; i++ {
		c := counter.NewClient(partitioner)
		defer c.Close()

		for j := 0; j != 3; j++ {
			unsub, err := c.Register("foo", func(c uint64) {})
			require.NoError(t, err)
			unregister = append(unregister, unsub)

			assert.Equal(t, uint64((i*3)+j+2), waitTimeout(updates, t))
		}
	}

	for i := len(unregister) - 1; i >= 0; i-- {
		require.Nil(t, unregister[i]())
		assert.Equal(t, uint64(i+1), waitTimeout(updates, t))
	}
}

func waitTimeout(ch <-chan uint64, t *testing.T) uint64 {
	select {
	case c := <-ch:
		return c
	case <-time.After(time.Second):
		t.Error("timeout waiting for update")
		return 0
	}
}
