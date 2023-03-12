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
	"net/http"
	"testing"

	"github.com/andydunstall/fuddle/demos/counter/pkg/testutils/cluster"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestFrontend_Connect(t *testing.T) {
	c, err := cluster.NewCluster(
		cluster.WithFuddleNodes(1),
		cluster.WithCounterNodes(1),
		cluster.WithFrontendNodes(1),
	)
	require.NoError(t, err)
	defer c.Shutdown()

	resp, err := http.Get("http://" + c.FrontendAddrs()[0] + "/foo")
	require.NoError(t, err)
	defer resp.Body.Close()

	assert.Equal(t, 200, resp.StatusCode)
}