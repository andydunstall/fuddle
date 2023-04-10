//go:build all || integration

package sdk

import (
	"context"
	"fmt"
	"math/rand"
	"sort"
	"testing"
	"time"

	fuddle "github.com/fuddle-io/fuddle-go"
	"github.com/fuddle-io/fuddle/pkg/testutils"
	"github.com/fuddle-io/fuddle/tests/cluster"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestRegister_RegisterMembers(t *testing.T) {
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

	var expectedMembers []fuddle.Member
	for _, client := range clients {
		if expectedMembers == nil {
			members := client.Members()
			sortMembers(members)
			expectedMembers = members
		}

		members := client.Members()
		sortMembers(members)
		assert.Equal(t, expectedMembers, members)
	}
}

func TestRegister_UnregisterOnClose(t *testing.T) {
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
		randomMember("member-base"),
		fuddle.WithLogger(testutils.Logger()),
	)
	require.NoError(t, err)
	defer client.Close()

	var clients []*fuddle.Fuddle
	for i := 0; i != 10; i++ {
		cl, err := fuddle.Register(
			ctx,
			c.RPCAddrs(),
			randomMember(fmt.Sprintf("member-%d", i)),
			fuddle.WithLogger(testutils.Logger()),
		)
		require.NoError(t, err)
		clients = append(clients, cl)
	}

	ctx, cancel = context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	assert.NoError(t, waitForMembers(ctx, client, 16))

	for _, cl := range clients {
		cl.Close()
	}

	assert.NoError(t, waitForMembers(ctx, client, 6))
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
