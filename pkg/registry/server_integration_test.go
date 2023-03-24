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

package registry_test

import (
	"context"
	"net"
	"testing"

	fuddle "github.com/fuddle-io/fuddle-go"
	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"github.com/fuddle-io/fuddle/pkg/registry"
	"github.com/fuddle-io/fuddle/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
)

// The server integration tests tests the integration of the Go SDK client
// using the gRPC API.

func TestServer_Register(t *testing.T) {
	r := registry.NewRegistry()
	addr, stop := startServer(r, t)
	defer stop()

	c, err := fuddle.Connect([]string{addr}, fuddle.WithLogger(testutils.Logger()))
	require.NoError(t, err)

	registeredNode := testutils.RandomSDKNode()
	_, err = c.Register(context.TODO(), registeredNode)
	assert.NoError(t, err)

	node, err := r.Node(registeredNode.ID)
	assert.NoError(t, err)
	assert.Equal(t, registeredNode, testutils.RPCNodeToSDKNode(node))
}

func TestServer_Unregister(t *testing.T) {
	r := registry.NewRegistry()
	addr, stop := startServer(r, t)
	defer stop()

	c, err := fuddle.Connect([]string{addr}, fuddle.WithLogger(testutils.Logger()))
	require.NoError(t, err)

	registeredNode := testutils.RandomSDKNode()
	n, err := c.Register(context.TODO(), registeredNode)
	assert.NoError(t, err)

	assert.NoError(t, n.Unregister(context.TODO()))

	_, err = r.Node(registeredNode.ID)
	assert.Equal(t, err, registry.ErrNotFound)
}

func TestServer_UpdateMetadata(t *testing.T) {
	r := registry.NewRegistry()
	addr, stop := startServer(r, t)
	defer stop()

	c, err := fuddle.Connect([]string{addr}, fuddle.WithLogger(testutils.Logger()))
	require.NoError(t, err)

	registeredNode := testutils.RandomSDKNode()
	n, err := c.Register(context.TODO(), registeredNode)
	assert.NoError(t, err)

	metadata := testutils.RandomMetadata()
	assert.NoError(t, n.UpdateMetadata(context.TODO(), metadata))

	expectedNode := registeredNode
	for k, v := range metadata {
		expectedNode.Metadata[k] = v
	}

	node, err := r.Node(registeredNode.ID)
	assert.NoError(t, err)
	assert.Equal(t, expectedNode, testutils.RPCNodeToSDKNode(node))
}

func TestServer_StreamUpdates(t *testing.T) {
	r := registry.NewRegistry()
	addr, stop := startServer(r, t)
	defer stop()

	observerConn, err := fuddle.Connect([]string{addr}, fuddle.WithLogger(testutils.Logger()))
	require.NoError(t, err)
	defer observerConn.Close()

	registerConn, err := fuddle.Connect([]string{addr}, fuddle.WithLogger(testutils.Logger()))
	require.NoError(t, err)
	defer registerConn.Close()

	registeredNode := testutils.RandomSDKNode()
	n, err := registerConn.Register(context.TODO(), registeredNode)
	assert.NoError(t, err)

	receivedNode, err := testutils.WaitForNode(observerConn, registeredNode.ID)
	assert.NoError(t, err)

	assert.Equal(t, receivedNode, registeredNode)

	assert.NoError(t, n.Unregister(context.TODO()))

	assert.NoError(t, testutils.WaitForCount(observerConn, 0))
}

func TestServer_Connect(t *testing.T) {
	addr, stop := startServer(registry.NewRegistry(), t)
	defer stop()

	_, err := fuddle.Connect([]string{addr})
	require.NoError(t, err)
}

func startServer(r *registry.Registry, t *testing.T) (string, func()) {
	s := registry.NewServer(r, registry.WithLogger(testutils.Logger()))

	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	grpcServer := grpc.NewServer()
	rpc.RegisterRegistryServer(grpcServer, s)
	go func() {
		require.NoError(t, grpcServer.Serve(ln))
	}()

	return ln.Addr().String(), func() {
		grpcServer.Stop()
	}
}
