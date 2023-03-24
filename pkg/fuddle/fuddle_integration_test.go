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

package fuddle_test

// The Fuddle integration tests test the services are integrated and configured
// correctly.

import (
	"context"
	"net"
	"testing"

	fuddle "github.com/fuddle-io/fuddle-go"
	"github.com/fuddle-io/fuddle/pkg/config"
	fuddleServer "github.com/fuddle-io/fuddle/pkg/fuddle"
	"github.com/fuddle-io/fuddle/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

// Tests the registry is configured and started.
func TestFuddle_Registry(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	server := fuddleServer.New(
		&config.Config{},
		fuddleServer.WithListener(ln),
		fuddleServer.WithLogger(testutils.Logger()),
	)
	require.NoError(t, server.Start())
	defer server.GracefulStop()

	observerConn, err := fuddle.Connect(
		[]string{ln.Addr().String()},
		fuddle.WithLogger(testutils.Logger()),
	)
	require.NoError(t, err)
	defer observerConn.Close()

	registerConn, err := fuddle.Connect(
		[]string{ln.Addr().String()},
		fuddle.WithLogger(testutils.Logger()),
	)
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
