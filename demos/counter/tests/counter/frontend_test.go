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
	"testing"

	"github.com/andydunstall/fuddle/demos/counter/pkg/testutils/cluster"
	"github.com/gorilla/websocket"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFrontend_Register(t *testing.T) {
	c, err := cluster.NewCluster(
		cluster.WithFuddleNodes(1),
		cluster.WithCounterNodes(3),
		cluster.WithFrontendNodes(3),
	)
	require.NoError(t, err)
	defer c.Shutdown()

	url := "ws://" + c.FrontendAddrs()[0] + "/foo"
	conn, _, err := websocket.DefaultDialer.Dial(url, nil)
	require.Nil(t, err)
	defer conn.Close()

	_, m, err := conn.ReadMessage()
	require.Nil(t, err)

	assert.Equal(t, "1", string(m))

	url2 := "ws://" + c.FrontendAddrs()[1] + "/foo"
	conn2, _, err := websocket.DefaultDialer.Dial(url2, nil)
	require.Nil(t, err)
	defer conn2.Close()

	_, m, err = conn.ReadMessage()
	require.Nil(t, err)

	assert.Equal(t, "2", string(m))
}
