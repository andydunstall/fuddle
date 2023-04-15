//go:build all || integration

package gossip

import (
	"context"
	"testing"
	"time"

	"github.com/fuddle-io/fuddle/pkg/fcm/cluster"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Creates a 5 node cluster and waits for each node to discover one another.
func TestGossip_ClusterDiscovery(t *testing.T) {
	c, err := cluster.NewCluster(cluster.WithFuddleNodes(5))
	require.Nil(t, err)
	defer c.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	assert.NoError(t, c.WaitForHealthy(ctx))
}

// Creates a 5 node cluster, then adds another node and waits for both the new
// node to discover the cluster, and the existing nodes to discover the new
// node.
func TestGossip_JoinCluster(t *testing.T) {
	c, err := cluster.NewCluster(cluster.WithFuddleNodes(5))
	require.Nil(t, err)
	defer c.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	assert.NoError(t, c.WaitForHealthy(ctx))

	// Add a new node.
	_, err = c.AddFuddleNode()
	require.Nil(t, err)

	ctx, cancel = context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	assert.NoError(t, c.WaitForHealthy(ctx))
}

// Creates a 5 node cluster, then adds another node and waits for discovery,
// then removes the node and waits for the rest of the cluster to detect it has
// left.
func TestGossip_LeaveCluster(t *testing.T) {
	c, err := cluster.NewCluster(cluster.WithFuddleNodes(5))
	require.Nil(t, err)
	defer c.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	assert.NoError(t, c.WaitForHealthy(ctx))

	// Add a new node.
	_, err = c.AddFuddleNode()
	require.Nil(t, err)

	ctx, cancel = context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	assert.NoError(t, c.WaitForHealthy(ctx))

	c.RemoveFuddleNode()

	ctx, cancel = context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	assert.NoError(t, c.WaitForHealthy(ctx))
}
