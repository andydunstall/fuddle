// Copyright (C) 2023 Andrew Dunstall
//
// Registry is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Registry is distributed in the hope that it will be useful,
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

func TestFilter(t *testing.T) {
	tests := []struct {
		Filter Filter
		Node   Node
		Match  bool
	}{
		// Empty service filter.
		{
			Filter: Filter{
				"service-2": {},
			},
			Node: Node{
				Service:  "service-2",
				Locality: "us-east-1-c",
				State: map[string]string{
					"bar": "boo",
				},
			},
			Match: true,
		},

		// Match.
		{
			Filter: Filter{
				"myservice": ServiceFilter{
					Locality: []string{"eu-west-1-*", "eu-west-2-*"},
					State: StateFilter{
						"foo": []string{"bar", "car", "boo"},
					},
				},
			},
			Node: Node{
				Service:  "myservice",
				Locality: "eu-west-2-c",
				State: map[string]string{
					"foo": "car",
				},
			},
			Match: true,
		},

		// Match multiple matching glob services.
		{
			Filter: Filter{
				"my*": ServiceFilter{
					Locality: []string{"eu-west-1-*", "eu-west-2-*"},
					State: StateFilter{
						"foo": []string{"bar", "car", "boo"},
					},
				},
				"*service": ServiceFilter{
					Locality: []string{"eu-west-1-*", "eu-west-2-*"},
					State: StateFilter{
						"foo": []string{"bar", "car", "boo"},
					},
				},
			},
			Node: Node{
				Service:  "myservice",
				Locality: "eu-west-2-c",
				State: map[string]string{
					"foo": "car",
				},
			},
			Match: true,
		},

		// No match with multiple matching glob services.
		{
			Filter: Filter{
				"my*": ServiceFilter{
					Locality: []string{"us-west-1-*", "us-west-2-*"},
					State: StateFilter{
						"foo": []string{"bar", "car", "boo"},
					},
				},
				"*service": ServiceFilter{
					Locality: []string{"eu-west-1-*", "eu-west-2-*"},
					State: StateFilter{
						"foo": []string{"bar", "car", "boo"},
					},
				},
			},
			Node: Node{
				Service:  "myservice",
				Locality: "eu-west-2-c",
				State: map[string]string{
					"foo": "car",
				},
			},
			Match: false,
		},

		// No matching services.
		{
			Filter: Filter{},
			Node: Node{
				Service:  "myservice",
				Locality: "eu-west-2-c",
				State: map[string]string{
					"foo": "car",
				},
			},
			Match: false,
		},
	}

	for _, tt := range tests {
		match := tt.Filter.Match(tt.Node)
		assert.Equal(t, tt.Match, match)
	}
}

func TestServiceFilter(t *testing.T) {
	tests := []struct {
		Filter   ServiceFilter
		Locality string
		State    map[string]string
		Match    bool
	}{
		// Node is a match.
		{
			Filter: ServiceFilter{
				Locality: []string{"eu-west-1-*", "eu-west-2-*"},
				State: StateFilter{
					"foo": []string{"bar", "car", "boo"},
				},
			},
			Locality: "eu-west-2-a",
			State: map[string]string{
				"foo": "car",
			},
			Match: true,
		},

		// Node doesn't match locality.
		{
			Filter: ServiceFilter{
				Locality: []string{"eu-west-1-*", "eu-west-2-*"},
				State: StateFilter{
					"foo": []string{"bar", "car", "boo"},
				},
			},
			Locality: "us-east-1-a",
			State: map[string]string{
				"foo": "car",
			},
			Match: false,
		},

		// Node doesn't match state.
		{
			Filter: ServiceFilter{
				Locality: []string{"eu-west-1-*", "eu-west-2-*"},
				State: StateFilter{
					"foo": []string{"bar", "car", "boo"},
				},
			},
			Locality: "eu-west-2-a",
			State: map[string]string{
				"foo": "xyz",
			},
			Match: false,
		},
	}

	for _, tt := range tests {
		match := tt.Filter.Match(Node{
			Locality: tt.Locality,
			State:    tt.State,
		})
		assert.Equal(t, tt.Match, match)
	}
}

func TestStateFilter(t *testing.T) {
	tests := []struct {
		Filter StateFilter
		State  map[string]string
		Match  bool
	}{
		// Value is a match.
		{
			Filter: StateFilter{
				"foo": []string{"bar", "car", "boo"},
			},
			State: map[string]string{
				"foo": "car",
			},
			Match: true,
		},

		// Value not a match.
		{
			Filter: StateFilter{
				"foo": []string{"bar", "car", "boo"},
			},
			State: map[string]string{
				"foo": "xyz",
			},
			Match: false,
		},

		// Value is a wildcard match.
		{
			Filter: StateFilter{
				"foo": []string{"bar", "car.*", "boo"},
			},
			State: map[string]string{
				"foo": "car.123",
			},
			Match: true,
		},

		// Node doesn't match all keys.
		{
			Filter: StateFilter{
				"foo": []string{"bar", "car", "boo"},
				"xyz": []string{"a", "b", "c"},
			},
			State: map[string]string{
				"foo": "car",
				"xyz": "d",
			},
			Match: false,
		},

		// Filter key not in node.
		{
			Filter: StateFilter{
				"foo": []string{"bar"},
			},
			State: map[string]string{},
			Match: false,
		},
	}

	for _, tt := range tests {
		match := tt.Filter.Match(Node{
			State: tt.State,
		})
		assert.Equal(t, tt.Match, match)
	}
}
