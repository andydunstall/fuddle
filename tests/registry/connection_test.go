//go:build all || integration

package registry

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/fuddle-io/fuddle/pkg/registry"
	"github.com/fuddle-io/fuddle/pkg/testutils"
	"github.com/fuddle-io/fuddle/tests/cluster"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnection_Reconnect(t *testing.T) {
	c, err := cluster.NewCluster(cluster.WithNodes(1))
	require.Nil(t, err)
	defer c.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	assert.NoError(t, c.WaitForHealthy(ctx))

	node := c.Nodes()[0]

	r := registry.NewRegistry("local")

	connStateCh := make(chan registry.ConnState)
	client, err := registry.Connect(
		node.Fuddle.Config.RPC.JoinAdvAddr(),
		r,
		registry.WithClientLogger(testutils.Logger()),
		registry.WithOnClientConnectionStateChange(func(state registry.ConnState) {
			connStateCh <- state
		}),
	)
	require.NoError(t, err)
	defer client.Close()

	state, err := waitForConnState(connStateCh)
	assert.NoError(t, err)
	assert.Equal(t, registry.StateConnected, state)

	// Drop the proxy connections and wait for the client to reconnect.

	node.RPCProxy.Drop()

	state, err = waitForConnState(connStateCh)
	assert.NoError(t, err)
	assert.Equal(t, registry.StateDisconnected, state)

	state, err = waitForConnState(connStateCh)
	assert.NoError(t, err)
	assert.Equal(t, registry.StateConnected, state)
}

func TestConnection_Connect(t *testing.T) {
	c, err := cluster.NewCluster(cluster.WithNodes(1))
	require.Nil(t, err)
	defer c.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	assert.NoError(t, c.WaitForHealthy(ctx))

	node := c.Nodes()[0]

	r := registry.NewRegistry("local")

	connStateCh := make(chan registry.ConnState)
	client, err := registry.Connect(
		node.Fuddle.Config.RPC.JoinAdvAddr(),
		r,
		registry.WithClientLogger(testutils.Logger()),
		registry.WithOnClientConnectionStateChange(func(state registry.ConnState) {
			connStateCh <- state
		}),
	)
	require.NoError(t, err)
	defer client.Close()

	state, err := waitForConnState(connStateCh)
	assert.NoError(t, err)
	assert.Equal(t, registry.StateConnected, state)
}

func waitForConnState(ch <-chan registry.ConnState) (registry.ConnState, error) {
	select {
	case c := <-ch:
		return c, nil
	case <-time.After(time.Second):
		return registry.ConnState(""), fmt.Errorf("timeout")
	}
}
