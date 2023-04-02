//go:build all || integration

package gossip_test

import (
	"context"
	"sort"
	"testing"
	"time"

	"github.com/fuddle-io/fuddle/pkg/gossip"
	"github.com/fuddle-io/fuddle/pkg/gossip/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestGossip_JoinCluster(t *testing.T) {
	c, err := testutils.NewCluster(5)
	require.Nil(t, err)
	defer c.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	assert.NoError(t, c.WaitForHealthy(ctx))

	nodesCh := make(chan gossip.Node, 128)
	require.Nil(t, c.AddNode(func(node gossip.Node) {
		nodesCh <- node
	}, nil))

	var joinedIDs []string
	for i := 0; i != 6; i++ {
		select {
		case n := <-nodesCh:
			joinedIDs = append(joinedIDs, n.ID)
		case <-ctx.Done():
			t.Error(ctx.Err())
			return
		}
	}

	sort.Strings(joinedIDs)
	assert.Equal(t, []string{
		"node-0", "node-1", "node-2", "node-3", "node-4", "node-5",
	}, joinedIDs)

	// Check all other nodes have discovered the new node.
	assert.NoError(t, c.WaitForHealthy(ctx))
}

func TestGossip_LeaveCluster(t *testing.T) {
	c, err := testutils.NewCluster(5)
	require.Nil(t, err)
	defer c.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*30)
	defer cancel()
	assert.NoError(t, c.WaitForHealthy(ctx))

	nodesCh := make(chan gossip.Node, 128)
	require.Nil(t, c.AddNode(nil, func(node gossip.Node) {
		nodesCh <- node
	}))
	require.Nil(t, err)

	// Check all other nodes have discovered the new node.
	assert.NoError(t, c.WaitForHealthy(ctx))

	c.RemoveNode(0)

	select {
	case n := <-nodesCh:
		assert.Equal(t, n.ID, "node-0")
	case <-ctx.Done():
		t.Error(ctx.Err())
		return
	}
}

// Tests creating a 5 node cluster and waits for all nodes to discover each
// another.
func TestGossip_WaitForHealthy(t *testing.T) {
	c, err := testutils.NewCluster(5)
	require.Nil(t, err)
	defer c.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*3)
	defer cancel()

	assert.NoError(t, c.WaitForHealthy(ctx))
}
