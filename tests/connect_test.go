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

package tests

import (
	"context"
	"testing"
	"time"

	fuddle "github.com/fuddle-io/fuddle-go"
	"github.com/fuddle-io/fuddle/pkg/testutils"
	"github.com/fuddle-io/fuddle/pkg/testutils/cluster"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConnect_ConnectOK(t *testing.T) {
	cluster, err := cluster.NewCluster()
	require.NoError(t, err)
	defer cluster.Close()

	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	c, err := fuddle.Connect(
		ctx,
		cluster.Addrs(),
		fuddle.WithLogger(testutils.Logger()),
	)
	require.NoError(t, err)
	defer c.Close()
}

// Tests the client connection will succeed even if some of the seed addresses
// are wrong.
func TestConnect_ConnectIgnoreBadAddrs(t *testing.T) {
	cluster, err := cluster.NewCluster()
	require.NoError(t, err)
	defer cluster.Close()

	addrs := []string{
		// Blocked port.
		"fuddle.io:12345",
		// Bad protocol.
		"fuddle.io:443",
		// No host.
		"notfound.fuddle.io:12345",
	}
	addrs = append(addrs, cluster.Addrs()...)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	c, err := fuddle.Connect(
		ctx,
		addrs,
		fuddle.WithLogger(testutils.Logger()),
		fuddle.WithConnectAttemptTimeout(time.Millisecond*100),
	)
	require.NoError(t, err)
	defer c.Close()
}

// Tests connecting to an unreachable address fails.
func TestConnect_ConnectUnreachable(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	// Attempt to connect to a blocked port.
	_, err := fuddle.Connect(
		ctx,
		[]string{"fuddle.io:12345"},
		fuddle.WithLogger(testutils.Logger()),
	)
	assert.Error(t, err)
}

// Tests connecting with no seed addresses fails.
func TestConnect_ConnectNoSeeds(t *testing.T) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Millisecond*100)
	defer cancel()

	_, err := fuddle.Connect(
		ctx,
		[]string{},
		fuddle.WithLogger(testutils.Logger()),
	)
	assert.Error(t, err)
}
