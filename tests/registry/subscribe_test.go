//go:build all || integration

package registry

import (
	"context"
	"fmt"
	"testing"
	"time"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"github.com/fuddle-io/fuddle/pkg/registry"
	"github.com/fuddle-io/fuddle/pkg/testutils"
	"github.com/fuddle-io/fuddle/tests/cluster"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestSubscribe_ReceiveNodesLocalMember(t *testing.T) {
	c, err := cluster.NewCluster(cluster.WithNodes(1))
	require.Nil(t, err)
	defer c.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	assert.NoError(t, c.WaitForHealthy(ctx))

	node := c.Nodes()[0]

	r := registry.NewRegistry("local")

	updatesCh := make(chan *rpc.RemoteMemberUpdate)
	req := &rpc.SubscribeRequest{}
	r.Subscribe(req, func(update *rpc.RemoteMemberUpdate) {
		updatesCh <- update
	})

	client, err := registry.Connect(
		node.Fuddle.Config.RPC.JoinAdvAddr(),
		r,
		registry.WithClientLogger(testutils.Logger()),
	)
	require.NoError(t, err)
	defer client.Close()

	// Wait to receive the servers local node.
	u, err := waitForUpdate(updatesCh)
	assert.NoError(t, err)
	assert.Equal(t, node.Fuddle.Config.NodeID, u.Member.Id)
	assert.Equal(t, rpc.MemberUpdateType_REGISTER, u.UpdateType)
}

func waitForUpdate(ch <-chan *rpc.RemoteMemberUpdate) (*rpc.RemoteMemberUpdate, error) {
	select {
	case u := <-ch:
		return u, nil
	case <-time.After(time.Second):
		return nil, fmt.Errorf("timeout")
	}
}
