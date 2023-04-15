//go:build all || integration

package registry

import (
	"context"
	"fmt"
	"testing"
	"time"

	"github.com/fuddle-io/fuddle/pkg/fcm/cluster"
	"github.com/fuddle-io/fuddle/pkg/registry"
	registryClient "github.com/fuddle-io/fuddle/pkg/registry/client"
	"github.com/fuddle-io/fuddle/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnection_Reconnect(t *testing.T) {
	c, err := cluster.NewCluster(cluster.WithFuddleNodes(1))
	require.Nil(t, err)
	defer c.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	assert.NoError(t, c.WaitForHealthy(ctx))

	node := c.FuddleNodes()[0]

	r := registry.NewRegistry("local")

	connStateCh := make(chan registryClient.ConnState)
	client, err := registryClient.Connect(
		node.Fuddle.Config.RPC.JoinAdvAddr(),
		r,
		registryClient.WithLogger(testutils.Logger()),
		registryClient.WithOnConnectionStateChange(func(state registryClient.ConnState) {
			connStateCh <- state
		}),
	)
	require.NoError(t, err)
	defer client.Close()

	state, err := waitForConnState(connStateCh)
	assert.NoError(t, err)
	assert.Equal(t, registryClient.StateConnected, state)

	// Drop the proxy connections and wait for the client to reconnect.

	node.RPCProxy.Drop()

	state, err = waitForConnState(connStateCh)
	assert.NoError(t, err)
	assert.Equal(t, registryClient.StateDisconnected, state)

	state, err = waitForConnState(connStateCh)
	assert.NoError(t, err)
	assert.Equal(t, registryClient.StateConnected, state)
}

func TestConnection_Connect(t *testing.T) {
	c, err := cluster.NewCluster(cluster.WithFuddleNodes(1))
	require.Nil(t, err)
	defer c.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	assert.NoError(t, c.WaitForHealthy(ctx))

	node := c.FuddleNodes()[0]

	r := registry.NewRegistry("local")

	connStateCh := make(chan registryClient.ConnState)
	client, err := registryClient.Connect(
		node.Fuddle.Config.RPC.JoinAdvAddr(),
		r,
		registryClient.WithLogger(testutils.Logger()),
		registryClient.WithOnConnectionStateChange(func(state registryClient.ConnState) {
			connStateCh <- state
		}),
	)
	require.NoError(t, err)
	defer client.Close()

	state, err := waitForConnState(connStateCh)
	assert.NoError(t, err)
	assert.Equal(t, registryClient.StateConnected, state)
}

func waitForConnState(ch <-chan registryClient.ConnState) (registryClient.ConnState, error) {
	select {
	case c := <-ch:
		return c, nil
	case <-time.After(time.Second):
		return registryClient.ConnState(""), fmt.Errorf("timeout")
	}
}
