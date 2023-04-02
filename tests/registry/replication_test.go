package registry

import (
	"context"
	"testing"
	"time"

	"github.com/fuddle-io/fuddle/tests/cluster"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests creating a 5 node cluster and waiting for each node to discovery one
// anothers registry entry.
func TestReplication_DiscoveryRegistry(t *testing.T) {
	c, err := cluster.NewCluster(cluster.WithNodes(5))
	require.Nil(t, err)
	defer c.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	assert.NoError(t, c.WaitForRegistryDiscovery(ctx))

	srcNode, ok := c.Node(0)
	assert.True(t, ok)
	destNode, ok := c.Node(1)
	assert.True(t, ok)

	registeredMemberID := uuid.New().String()

	waitCh := make(chan interface{})
	unsub := destNode.Registry().Subscribe(false, func(id string) {
		if id == registeredMemberID {
			close(waitCh)
		}
	})
	defer unsub()

	srcNode.Registry().RegisterLocal(registeredMemberID)

	select {
	case <-waitCh:
	case <-ctx.Done():
		t.Error("timeout")
	}
}
