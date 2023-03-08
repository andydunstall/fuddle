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

package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestQuery_MatchingState(t *testing.T) {
	tests := []struct {
		Node          Node
		Query         Query
		ExpectedMatch bool
		ExpectedState map[string]string
	}{
		// Matching service with state filtered.
		{
			Node: Node{
				Service: "xyz",
				State: map[string]string{
					"foo": "a",
					"bar": "b",
					"car": "c",
				},
			},
			Query: Query{
				"xyz": &ServiceQuery{
					State: []string{"bar", "car"},
				},
			},
			ExpectedMatch: true,
			ExpectedState: map[string]string{
				"bar": "b",
				"car": "c",
			},
		},
		// Matching service with no matching state.
		{
			Node: Node{
				Service: "xyz",
				State: map[string]string{
					"foo": "a",
					"bar": "b",
					"car": "c",
				},
			},
			Query: Query{
				"xyz": &ServiceQuery{
					State: []string{"xyz", "qrs"},
				},
			},
			ExpectedMatch: false,
		},
		// Matching service with no state query.
		{
			Node: Node{
				Service: "xyz",
				State: map[string]string{
					"foo": "a",
					"bar": "b",
					"car": "c",
				},
			},
			Query: Query{
				"xyz": nil,
			},
			ExpectedMatch: true,
			ExpectedState: map[string]string{
				"foo": "a",
				"bar": "b",
				"car": "c",
			},
		},
		// Matching service with no state query fields.
		{
			Node: Node{
				Service: "xyz",
				State: map[string]string{
					"foo": "a",
					"bar": "b",
					"car": "c",
				},
			},
			Query: Query{
				"xyz": &ServiceQuery{},
			},
			ExpectedMatch: true,
			ExpectedState: nil,
		},
		// Matching service with no state query fields.
		{
			Node: Node{
				Service: "xyz",
				State: map[string]string{
					"foo": "a",
					"bar": "b",
					"car": "c",
				},
			},
			Query: Query{
				"xyz": &ServiceQuery{
					State: []string{},
				},
			},
			ExpectedMatch: true,
			ExpectedState: nil,
		},
		// No matching service.
		{
			Node: Node{
				Service: "abc",
				State: map[string]string{
					"foo": "a",
					"bar": "b",
					"car": "c",
				},
			},
			Query: Query{
				"xyz": nil,
			},
			ExpectedMatch: false,
		},
	}
	for _, tt := range tests {
		state, match := tt.Query.MatchingState(tt.Node)
		assert.Equal(t, tt.ExpectedMatch, match)
		if match {
			assert.Equal(t, tt.ExpectedState, state)
		}
	}
}
