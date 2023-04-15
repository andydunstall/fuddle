//go:build all || integration

package server

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	"github.com/fuddle-io/fuddle/pkg/fcm"
	"github.com/fuddle-io/fuddle/pkg/fcm/client"
	"github.com/fuddle-io/fuddle/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFCMCluster_CreateCluster(t *testing.T) {
	t.Parallel()

	ln, err := tcpListen()
	require.NoError(t, err)

	server, err := fcm.NewFCM(
		"127.0.0.1", 0,
		fcm.WithListener(ln),
		fcm.WithLogger(testutils.Logger()),
	)
	require.NoError(t, err)
	defer server.Shutdown()

	client := client.NewClient(ln.Addr().String())

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	clusterInfo, err := client.ClusterCreate(ctx, 3, 10)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(clusterInfo.FuddleNodes))
	assert.Equal(t, 10, len(clusterInfo.ClientNodes))

	clusterInfo, err = client.ClusterInfo(ctx, clusterInfo.ID)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(clusterInfo.FuddleNodes))
	assert.Equal(t, 10, len(clusterInfo.ClientNodes))
}

func TestFCMCluster_AddAndRemoveNodes(t *testing.T) {
	t.Parallel()

	ln, err := tcpListen()
	require.NoError(t, err)

	server, err := fcm.NewFCM(
		"127.0.0.1", 0,
		fcm.WithListener(ln),
		fcm.WithLogger(testutils.Logger()),
	)
	require.NoError(t, err)
	defer server.Shutdown()

	client := client.NewClient(ln.Addr().String())

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*15)
	defer cancel()

	clusterInfo, err := client.ClusterCreate(ctx, 3, 10)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(clusterInfo.FuddleNodes))
	assert.Equal(t, 10, len(clusterInfo.ClientNodes))

	nodesInfo, err := client.AddNodes(ctx, clusterInfo.ID, 2, 5)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(nodesInfo.FuddleNodes))
	assert.Equal(t, 5, len(nodesInfo.ClientNodes))

	clusterInfo, err = client.ClusterInfo(ctx, clusterInfo.ID)
	assert.NoError(t, err)
	assert.Equal(t, 5, len(clusterInfo.FuddleNodes))
	assert.Equal(t, 15, len(clusterInfo.ClientNodes))

	nodesInfo, err = client.RemoveNodes(ctx, clusterInfo.ID, 3, 7)
	assert.NoError(t, err)
	assert.Equal(t, 3, len(nodesInfo.FuddleNodes))
	assert.Equal(t, 7, len(nodesInfo.ClientNodes))

	clusterInfo, err = client.ClusterInfo(ctx, clusterInfo.ID)
	assert.NoError(t, err)
	assert.Equal(t, 2, len(clusterInfo.FuddleNodes))
	assert.Equal(t, 8, len(clusterInfo.ClientNodes))
}

func tcpListen() (*net.TCPListener, error) {
	ln, err := net.ListenTCP("tcp", &net.TCPAddr{
		IP:   net.ParseIP("127.0.0.1"),
		Port: 0,
	})
	if err != nil {
		return nil, fmt.Errorf("tcp listen: %w", err)
	}
	return ln, nil
}
