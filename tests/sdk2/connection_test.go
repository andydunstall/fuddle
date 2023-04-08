//go:build all || integration

package sdk2

import (
	"context"
	"testing"
	"time"

	fuddle "github.com/fuddle-io/fuddle/pkg/sdk2"
	"github.com/fuddle-io/fuddle/pkg/testutils"
	"github.com/fuddle-io/fuddle/tests/cluster"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnection_Connect(t *testing.T) {
	t.Parallel()

	c, err := cluster.NewCluster(cluster.WithNodes(5))
	require.Nil(t, err)
	defer c.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	assert.NoError(t, c.WaitForHealthy(ctx))

	ctx, cancel = context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	client, err := fuddle.Connect(
		ctx,
		c.RPCAddrs(),
		fuddle.WithLogger(testutils.Logger()),
	)
	require.NoError(t, err)
	defer client.Close()
}

// Tests the client connection will succeed even if some of the seed addresses
// are wrong.
func TestConnection_ConnectIgnoreBadAddrs(t *testing.T) {
	t.Parallel()

	c, err := cluster.NewCluster(cluster.WithNodes(1))
	require.NoError(t, err)
	defer c.Shutdown()

	addrs := []string{
		// Blocked port.
		"fuddle.io:12345",
		// Bad protocol.
		"fuddle.io:443",
		// No host.
		"notfound.fuddle.io:12345",
	}
	addrs = append(addrs, c.RPCAddrs()...)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	client, err := fuddle.Connect(
		ctx,
		addrs,
		fuddle.WithLogger(testutils.Logger()),
		fuddle.WithConnectAttemptTimeout(time.Millisecond*100),
	)
	require.NoError(t, err)
	defer client.Close()
}

// Tests connecting to an unreachable address fails.
func TestConnection_ConnectUnreachable(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	// Attempt to connect to a blocked port.
	_, err := fuddle.Connect(
		ctx,
		[]string{"fuddle.io:12345"},
		fuddle.WithLogger(testutils.Logger()),
	)
	assert.Error(t, err)
}

// Tests connecting with no seed addresses fails.
func TestConnection_ConnectNoSeeds(t *testing.T) {
	t.Parallel()

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	_, err := fuddle.Connect(
		ctx,
		[]string{},
		fuddle.WithLogger(testutils.Logger()),
	)
	assert.Error(t, err)
}
