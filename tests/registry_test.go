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

package tests

import (
	"context"
	"sort"
	"testing"
	"time"

	fuddle "github.com/fuddle-io/fuddle-go"
	"github.com/fuddle-io/fuddle/pkg/testutils"
	"github.com/fuddle-io/fuddle/pkg/testutils/cluster"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegistry_RegisterAndStreamUpdates(t *testing.T) {
	cluster, err := cluster.NewCluster()
	require.NoError(t, err)
	defer cluster.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	var clients []*fuddle.Fuddle
	for i := 0; i != 5; i++ {
		c, err := fuddle.Connect(
			ctx,
			cluster.Addrs(),
			fuddle.WithLogger(testutils.Logger()),
		)
		require.NoError(t, err)
		defer c.Close()

		clients = append(clients, c)
	}

	// Register 10 members in on each client.
	for _, c := range clients {
		for i := 0; i != 10; i++ {
			assert.NoError(t, c.Register(ctx, testutils.RandomSDKMember()))
		}
	}

	// Wait for each client to receive one anothers members and the Fuddle nodes
	// members.
	for _, c := range clients {
		assert.NoError(t, testutils.WaitForMembers(ctx, c, 51))
	}

	// Verify each set has the same view of the registry.
	expectedMembers := clients[0].Members()
	for _, c := range clients {
		assert.Equal(t, sortMembers(expectedMembers), sortMembers(c.Members()))
	}
}

// Tests registering and streaming updates works after all active connections
// are blocked. This will require register RPCs to retry and streaming updates
// to reconnect.
func TestRegistry_RegisterAndStreamUpdatesAfterConnectionUnresponsive(t *testing.T) {
	cluster, err := cluster.NewCluster()
	require.NoError(t, err)
	defer cluster.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*60)
	defer cancel()

	var clients []*fuddle.Fuddle
	for i := 0; i != 5; i++ {
		c, err := fuddle.Connect(
			ctx,
			cluster.Addrs(),
			fuddle.WithLogger(testutils.Logger()),
		)
		require.NoError(t, err)
		defer c.Close()

		clients = append(clients, c)
	}

	// Block all traffic on active connections.
	cluster.BlockActiveConns()

	// Register 10 members in on each client.
	for _, c := range clients {
		for i := 0; i != 10; i++ {
			assert.NoError(t, c.Register(ctx, testutils.RandomSDKMember()))
		}
	}

	// Wait for each client to receive one anothers members and the Fuddle nodes
	// members.
	for _, c := range clients {
		assert.NoError(t, testutils.WaitForMembers(ctx, c, 51))
	}

	// Verify each set has the same view of the registry.
	expectedMembers := clients[0].Members()
	for _, c := range clients {
		assert.Equal(t, sortMembers(expectedMembers), sortMembers(c.Members()))
	}
}

func TestRegistry_UnregisterAndStreamUpdates(t *testing.T) {
	cluster, err := cluster.NewCluster()
	require.NoError(t, err)
	defer cluster.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	var clients []*fuddle.Fuddle
	for i := 0; i != 5; i++ {
		c, err := fuddle.Connect(
			ctx,
			cluster.Addrs(),
			fuddle.WithLogger(testutils.Logger()),
		)
		require.NoError(t, err)
		defer c.Close()

		clients = append(clients, c)
	}

	// Register 10 members in on each client.
	for _, c := range clients {
		for i := 0; i != 10; i++ {
			assert.NoError(t, c.Register(ctx, testutils.RandomSDKMember()))
		}
	}

	// Wait for each client to receive one anothers members and the Fuddle nodes
	// members.
	for _, c := range clients {
		assert.NoError(t, testutils.WaitForMembers(ctx, c, 51))
	}

	// Unregister all members.
	for _, c := range clients {
		for _, m := range c.LocalMembers() {
			assert.NoError(t, c.Unregister(ctx, m.ID))
		}
	}

	// Wait for each client to receive unregister requests.
	for _, c := range clients {
		assert.NoError(t, testutils.WaitForMembers(ctx, c, 1))
	}

	// Verify each set has the same view of the registry.
	expectedMembers := clients[0].Members()
	for _, c := range clients {
		assert.Equal(t, sortMembers(expectedMembers), sortMembers(c.Members()))
	}
}

func TestRegistry_UnregisterOnClose(t *testing.T) {
	cluster, err := cluster.NewCluster()
	require.NoError(t, err)
	defer cluster.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	client, err := fuddle.Connect(
		ctx,
		cluster.Addrs(),
		fuddle.WithLogger(testutils.Logger()),
	)
	require.NoError(t, err)
	defer client.Close()

	var clients []*fuddle.Fuddle
	for i := 0; i != 5; i++ {
		c, err := fuddle.Connect(
			ctx,
			cluster.Addrs(),
			fuddle.WithLogger(testutils.Logger()),
		)
		require.NoError(t, err)

		clients = append(clients, c)
	}

	// Register 10 members in on each client.
	for _, c := range clients {
		for i := 0; i != 10; i++ {
			assert.NoError(t, c.Register(ctx, testutils.RandomSDKMember()))
		}
	}

	assert.NoError(t, testutils.WaitForMembers(ctx, client, 51))

	// Close all clients which should unregister all members.
	for _, c := range clients {
		c.Close()
	}

	assert.NoError(t, testutils.WaitForMembers(ctx, client, 1))
}

// Tests Fuddle clients receive the Fuddle node members when they connect.
func TestRegistry_ReceiveFuddleNodeMember(t *testing.T) {
	cluster, err := cluster.NewCluster()
	require.NoError(t, err)
	defer cluster.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	c, err := fuddle.Connect(
		ctx,
		cluster.Addrs(),
		fuddle.WithLogger(testutils.Logger()),
	)
	require.NoError(t, err)
	defer c.Close()

	// Wait to receive the Fuddle nodes members.
	assert.NoError(t, testutils.WaitForMembers(ctx, c, 1))

	members := c.Members()
	assert.Equal(t, 1, len(members))
	assert.Equal(t, "fuddle", members[0].Service)
}

func sortMembers(m []fuddle.Member) []fuddle.Member {
	sort.Slice(m, func(i, j int) bool {
		return m[i].ID < m[j].ID
	})
	return m
}
