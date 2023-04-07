//go:build all || integration

package registry

import (
	"context"
	"testing"
	"time"

	"github.com/fuddle-io/fuddle/tests/cluster"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestReplication(t *testing.T) {
	c, err := cluster.NewCluster(cluster.WithNodes(5))
	require.Nil(t, err)
	defer c.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	// Waits for registries to discover one another.
	assert.NoError(t, c.WaitForHealthy(ctx))
}
