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
	"testing"
	"time"

	fuddle "github.com/fuddle-io/fuddle-go"
	"github.com/fuddle-io/fuddle/pkg/testutils"
	"github.com/fuddle-io/fuddle/pkg/testutils/cluster"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests a client will reconnect to another node after the original connected
// node closes.
func TestReconnect_ReconnectAfterDrop(t *testing.T) {
	cluster, err := cluster.NewCluster()
	require.NoError(t, err)
	defer cluster.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	connStateCh := make(chan fuddle.ConnState, 10)
	c, err := fuddle.Connect(
		ctx,
		cluster.Addrs(),
		fuddle.WithLogger(testutils.Logger()),
		fuddle.WithOnConnectionStateChange(func(state fuddle.ConnState) {
			connStateCh <- state
		}),
	)
	require.NoError(t, err)
	defer c.Close()

	assert.Equal(t, fuddle.StateConnected, <-connStateCh)

	// Close the node the client is connected to.
	cluster.CloseIfActive()

	assert.Equal(t, fuddle.StateDisconnected, <-connStateCh)
	assert.Equal(t, fuddle.StateConnected, <-connStateCh)
}

// Tests a client will reconnect after its connection is blocked by the proxy
// dropping all traffic (even though the connection remains open).
func TestReconnect_ReconnectAfterBlock(t *testing.T) {
	cluster, err := cluster.NewCluster()
	require.NoError(t, err)
	defer cluster.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	connStateCh := make(chan fuddle.ConnState, 10)
	c, err := fuddle.Connect(
		ctx,
		cluster.Addrs(),
		fuddle.WithLogger(testutils.Logger()),
		fuddle.WithKeepAlivePingInterval(time.Second),
		fuddle.WithKeepAlivePingTimeout(time.Millisecond*100),
		fuddle.WithOnConnectionStateChange(func(state fuddle.ConnState) {
			connStateCh <- state
		}),
	)
	require.NoError(t, err)
	defer c.Close()

	assert.Equal(t, fuddle.StateConnected, <-connStateCh)

	// Block all traffic on active connections.
	cluster.BlockActiveConns()

	assert.Equal(t, fuddle.StateDisconnected, <-connStateCh)
	assert.Equal(t, fuddle.StateConnected, <-connStateCh)
}
