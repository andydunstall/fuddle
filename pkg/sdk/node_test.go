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

package sdk

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

// Tests Node.Copy returns a copy that is equal to the original, and
// changing state in one won't affect the other.
func TestNode_Copy(t *testing.T) {
	node := Node{
		ID:       "local-123",
		Service:  "foo",
		Locality: "us-east-1-a",
		Revision: "v0.1.0",
		State: map[string]string{
			"a": "1",
			"b": "2",
			"c": "3",
		},
	}

	// Verify the copy is the same as the original.
	nodeCopy := node.Copy()
	assert.Equal(t, node, nodeCopy)

	// Verify changing the state of the original doesn't affect the copy.
	node.State["a"] = "5"
	assert.Equal(t, "1", nodeCopy.State["a"])
}
