//go:build all || integration

package registry

import (
	"context"
	"testing"
	"time"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"github.com/fuddle-io/fuddle/tests/cluster"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

func TestClient_Register(t *testing.T) {
	c, err := cluster.NewCluster(cluster.WithNodes(5))
	require.Nil(t, err)
	defer c.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	assert.NoError(t, c.WaitForHealthy(ctx))

	node := c.Nodes()[0]

	conn, err := grpc.DialContext(
		context.Background(),
		node.Fuddle.Config.RPC.JoinAdvAddr(),
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	require.Nil(t, err)
	client := rpc.NewRegistryClient(conn)

	updatesCh := make(chan *rpc.RemoteMemberUpdate, 10)
	req := &rpc.SubscribeRequest{}
	node.Fuddle.Registry().Subscribe(req, func(update *rpc.RemoteMemberUpdate) {
		updatesCh <- update
	})

	stream, err := client.Register(context.Background())
	require.Nil(t, err)

	err = stream.Send(&rpc.ClientUpdate{
		UpdateType: rpc.ClientUpdateType_CLIENT_REGISTER,
		Member: &rpc.Member{
			Id: "member-1",
		},
		SeqId: 1,
	})
	require.Nil(t, err)

	for {
		select {
		case u := <-updatesCh:
			if u.Member.Id == "member-1" {
				return
			}
		case <-time.After(time.Second):
			t.Error("timeout")
		}
	}
}
