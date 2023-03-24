// Copyright (C) 2023 Andrew Dunstall
//
// Fuddle is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Fuddle is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

//go:build integration

package client

import (
	"context"
	"net"
	"testing"
	"time"

	fuddle "github.com/fuddle-io/fuddle-go"
	"github.com/fuddle-io/fuddle/pkg/config"
	fuddleServer "github.com/fuddle-io/fuddle/pkg/fuddle"
	"github.com/fuddle-io/fuddle/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/protobuf/proto"
)

func TestAdmin_Cluster(t *testing.T) {
	addr, stop := startServer(t)
	defer stop()

	conn, err := fuddle.Connect(
		[]string{addr},
		fuddle.WithLogger(testutils.Logger()),
	)
	require.NoError(t, err)
	defer conn.Close()

	registeredNode := testutils.RandomRPCNode()
	_, err = conn.Register(
		context.TODO(), testutils.RPCNodeToSDKNode(registeredNode),
	)
	assert.NoError(t, err)

	admin, err := NewAdmin(addr)
	assert.NoError(t, err)

	nodes, err := admin.Cluster(context.TODO())
	assert.NoError(t, err)

	// Expect Cluster to not return node metadata.
	expectedNode := registeredNode
	expectedNode.Metadata = nil

	assert.Equal(t, 1, len(nodes))
	assert.True(t, proto.Equal(registeredNode, nodes[0]))
}

func TestAdmin_Node(t *testing.T) {
	addr, stop := startServer(t)
	defer stop()

	conn, err := fuddle.Connect(
		[]string{addr},
		fuddle.WithLogger(testutils.Logger()),
	)
	require.NoError(t, err)
	defer conn.Close()

	registeredNode := testutils.RandomSDKNode()
	_, err = conn.Register(context.TODO(), registeredNode)
	assert.NoError(t, err)

	admin, err := NewAdmin(addr)
	assert.NoError(t, err)

	respNode, err := admin.Node(context.TODO(), registeredNode.ID)
	assert.NoError(t, err)

	assert.Equal(t, testutils.RPCNodeToSDKNode(respNode), registeredNode)
}

func TestAdmin_NodeNotFound(t *testing.T) {
	addr, stop := startServer(t)
	defer stop()

	conn, err := fuddle.Connect(
		[]string{addr},
		fuddle.WithLogger(testutils.Logger()),
	)
	require.NoError(t, err)
	defer conn.Close()

	admin, err := NewAdmin(addr)
	assert.NoError(t, err)

	_, err = admin.Node(context.TODO(), "not-found")
	assert.Error(t, err)
}

func TestAdmin_BadConnection(t *testing.T) {
	// Blocked port.
	addr := "fuddle.io:12345"
	_, err := NewAdmin(addr, WithConnectTimeout(time.Millisecond*100))
	require.Error(t, err)
}

func startServer(t *testing.T) (string, func()) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)
	server := fuddleServer.New(
		&config.Config{},
		fuddleServer.WithListener(ln),
		fuddleServer.WithLogger(testutils.Logger()),
	)
	require.NoError(t, server.Start())
	return ln.Addr().String(), func() {
		server.GracefulStop()
	}
}
