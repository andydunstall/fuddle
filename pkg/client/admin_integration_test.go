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

	"github.com/fuddle-io/fuddle/pkg/config"
	"github.com/fuddle-io/fuddle/pkg/fuddle"
	"github.com/fuddle-io/fuddle/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestAdmin_Cluster(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	conf := config.DefaultConfig()
	server := fuddle.New(
		conf, fuddle.WithListener(ln), fuddle.WithLogger(testutils.Logger()),
	)
	require.NoError(t, server.Start())
	defer server.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	admin := NewAdmin(ln.Addr().String())
	nodes, err := admin.Cluster(ctx)
	assert.NoError(t, err)

	assert.Equal(t, 1, len(nodes))
	assert.Equal(t, conf.ID, nodes[0].ID)
}

func TestAdmin_Node(t *testing.T) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	require.NoError(t, err)

	conf := config.DefaultConfig()
	server := fuddle.New(
		conf, fuddle.WithListener(ln), fuddle.WithLogger(testutils.Logger()),
	)
	require.NoError(t, server.Start())
	defer server.Stop()

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	admin := NewAdmin(ln.Addr().String())
	nodes, err := admin.Cluster(ctx)
	assert.NoError(t, err)
	node, err := admin.Node(ctx, nodes[0].ID)
	assert.NoError(t, err)

	assert.Equal(t, conf.ID, node.ID)
}
