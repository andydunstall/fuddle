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

func TestUpdates_ReceiveMembers(t *testing.T) {
	t.Parallel()

	c, err := cluster.NewCluster(cluster.WithNodes(5))
	require.Nil(t, err)
	defer c.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	assert.NoError(t, c.WaitForHealthy(ctx))

	ctx, cancel = context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	client, err := fuddle.Register(
		ctx,
		c.RPCAddrs(),
		randomMember(""),
		fuddle.WithLogger(testutils.Logger()),
	)
	require.NoError(t, err)
	defer client.Close()

	ctx, cancel = context.WithTimeout(context.Background(), time.Second*2)
	defer cancel()

	assert.NoError(t, waitForMembers(ctx, client, 6))
}

func TestUpdates_ReceiveMissedMembersAfterReconnect(t *testing.T) {
	t.Skip("TODO")
}

func waitForMembers(ctx context.Context, client *fuddle.Fuddle, count int) error {
	found := false
	recvCh := make(chan interface{})
	unsubscribe := client.Subscribe(func() {
		if found {
			return
		}

		if len(client.Members()) == count {
			found = true
			close(recvCh)
			return
		}
	})
	defer unsubscribe()

	if err := waitWithContext(ctx, recvCh); err != nil {
		return err
	}
	return nil
}

func waitWithContext(ctx context.Context, ch chan interface{}) error {
	select {
	case <-ch:
		return nil
	case <-ctx.Done():
		return ctx.Err()
	}
}
