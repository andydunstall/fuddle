//go:build all || integration

package sdk

import (
	"context"
	"testing"
	"time"

	fuddle "github.com/fuddle-io/fuddle/pkg/sdk"
	"github.com/fuddle-io/fuddle/tests/cluster"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSDK_RegisterMember(t *testing.T) {
	c, err := cluster.NewCluster(cluster.WithNodes(5))
	require.Nil(t, err)
	defer c.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	assert.NoError(t, c.WaitForHealthy(ctx))

	ctx, cancel = context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	clientSub, err := fuddle.Connect(ctx, c.RPCAddrs())
	assert.NoError(t, err)

	clientRegister, err := fuddle.Connect(ctx, c.RPCAddrs())
	assert.NoError(t, err)

	assert.NoError(t, clientRegister.Register(ctx, fuddle.Member{
		ID: "foo",
	}))

	assert.NoError(t, waitForMembers(ctx, clientSub, 6))
}

func TestSDK_DiscoverFuddleNodes(t *testing.T) {
	c, err := cluster.NewCluster(cluster.WithNodes(5))
	require.Nil(t, err)
	defer c.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	assert.NoError(t, c.WaitForHealthy(ctx))

	ctx, cancel = context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	client, err := fuddle.Connect(ctx, c.RPCAddrs())
	assert.NoError(t, err)

	assert.NoError(t, waitForMembers(ctx, client, 5))
}

func TestSDK_Connect(t *testing.T) {
	c, err := cluster.NewCluster(cluster.WithNodes(5))
	require.Nil(t, err)
	defer c.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	assert.NoError(t, c.WaitForHealthy(ctx))

	ctx, cancel = context.WithTimeout(context.Background(), time.Second*1)
	defer cancel()
	_, err = fuddle.Connect(ctx, c.RPCAddrs())
	assert.NoError(t, err)
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
