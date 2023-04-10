//go:build all || integration

package registry

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"testing"
	"time"

	fuddle "github.com/fuddle-io/fuddle-go"
	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"github.com/fuddle-io/fuddle/pkg/admin"
	"github.com/fuddle-io/fuddle/pkg/testutils"
	"github.com/fuddle-io/fuddle/tests/cluster"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestAdmin_ListMembers(t *testing.T) {
	t.Parallel()

	c, err := cluster.NewCluster(cluster.WithNodes(5))
	require.Nil(t, err)
	defer c.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	assert.NoError(t, c.WaitForHealthy(ctx))

	ctx, cancel = context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	var clients []*fuddle.Fuddle
	for i := 0; i != 10; i++ {
		client, err := fuddle.Register(
			ctx,
			c.RPCAddrs(),
			randomMember(fmt.Sprintf("member-%d", i)),
			fuddle.WithLogger(testutils.Logger()),
		)
		require.NoError(t, err)
		defer client.Close()

		clients = append(clients, client)
	}

	ctx, cancel = context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	for _, client := range clients {
		assert.NoError(t, waitForMembers(ctx, client, 15))
	}

	expectedMembers := clients[0].Members()

	adminClient, err := admin.Connect(c.Nodes()[0].Fuddle.Config.RPC.JoinAdvAddr())
	assert.NoError(t, err)

	// List the members via the admin client and check the result matches the
	// registry view.

	ctx, cancel = context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	members, err := adminClient.Members(ctx)
	assert.NoError(t, err)

	var sdkMembers []fuddle.Member
	for _, m := range members {
		sdkMembers = append(sdkMembers, fromRPC(m))
	}

	sortMembers(expectedMembers)
	sortMembers(sdkMembers)
	assert.Equal(t, expectedMembers, sdkMembers)
}

func TestAdmin_GetMember(t *testing.T) {
	t.Parallel()

	c, err := cluster.NewCluster(cluster.WithNodes(5))
	require.Nil(t, err)
	defer c.Shutdown()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	assert.NoError(t, c.WaitForHealthy(ctx))

	ctx, cancel = context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	var clients []*fuddle.Fuddle
	for i := 0; i != 10; i++ {
		client, err := fuddle.Register(
			ctx,
			c.RPCAddrs(),
			randomMember(fmt.Sprintf("member-%d", i)),
			fuddle.WithLogger(testutils.Logger()),
		)
		require.NoError(t, err)
		defer client.Close()

		clients = append(clients, client)
	}

	ctx, cancel = context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	for _, client := range clients {
		assert.NoError(t, waitForMembers(ctx, client, 15))
	}

	adminClient, err := admin.Connect(c.Nodes()[0].Fuddle.Config.RPC.JoinAdvAddr())
	assert.NoError(t, err)

	ctx, cancel = context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	member, err := adminClient.Member(ctx, "member-5")
	assert.NoError(t, err)

	expected, _ := c.Nodes()[0].Fuddle.Registry().Member("member-5")
	assert.True(t, proto.Equal(expected, member))
}

func randomMember(id string) fuddle.Member {
	if id == "" {
		id = uuid.New().String()
	}
	return fuddle.Member{
		ID:       id,
		Service:  uuid.New().String(),
		Locality: uuid.New().String(),
		Created:  rand.Int63(),
		Revision: uuid.New().String(),
		Metadata: map[string]string{
			uuid.New().String(): uuid.New().String(),
			uuid.New().String(): uuid.New().String(),
			uuid.New().String(): uuid.New().String(),
			uuid.New().String(): uuid.New().String(),
			uuid.New().String(): uuid.New().String(),
		},
	}
}

func sortMembers(m []fuddle.Member) {
	sort.Slice(m, func(i, j int) bool {
		return m[i].ID < m[j].ID
	})
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

func fromRPC(m *rpc.Member) fuddle.Member {
	return fuddle.Member{
		ID:       m.Id,
		Service:  m.Service,
		Locality: m.Locality,
		Created:  m.Created,
		Revision: m.Revision,
		Metadata: m.Metadata,
	}
}
