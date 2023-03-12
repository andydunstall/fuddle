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

package counter

import (
	"fmt"
	"math/rand"
	"strconv"
	"testing"

	"github.com/andydunstall/fuddle/demos/counter/pkg/testutils/cluster"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

type connection struct {
	ws *websocket.Conn
}

func newConnection(ws *websocket.Conn) *connection {
	return &connection{
		ws: ws,
	}
}

func (c *connection) Recv() (uint64, error) {
	_, m, err := c.ws.ReadMessage()
	if err != nil {
		return 0, err
	}
	n, err := strconv.ParseUint(string(m), 10, 64)
	if err != nil {
		return 0, err
	}
	return n, nil
}

func (c *connection) Close() error {
	c.ws.Close()
	return nil
}

type client struct {
	addrs []string
}

func newClient(addrs []string) *client {
	return &client{
		addrs: addrs,
	}
}

func (c *client) Register(id string) (*connection, error) {
	if len(c.addrs) == 0 {
		return nil, fmt.Errorf("no nodes available")
	}

	addr := c.addrs[rand.Intn(len(c.addrs))]
	url := "ws://" + addr + "/foo"
	ws, _, err := websocket.DefaultDialer.Dial(url, nil)
	if err != nil {
		return nil, err
	}
	return newConnection(ws), nil
}

// Tests registering increments the count and receives update when the
// count changes.
func TestService_Register(t *testing.T) {
	c, err := cluster.NewCluster(
		cluster.WithFuddleNodes(1),
		cluster.WithCounterNodes(3),
		cluster.WithFrontendNodes(3),
	)
	require.NoError(t, err)
	defer c.Shutdown()

	client := newClient(c.FrontendAddrs())

	conn, err := client.Register("foo")

	require.Nil(t, err)
	defer conn.Close()

	m, err := conn.Recv()
	require.Nil(t, err)
	assert.Equal(t, uint64(1), m)

	for i := 0; i != 10; i++ {
		subConn, err := client.Register("foo")
		require.Nil(t, err)
		defer subConn.Close()

		m, err := conn.Recv()
		require.Nil(t, err)
		assert.Equal(t, uint64(i+2), m)
	}
}
