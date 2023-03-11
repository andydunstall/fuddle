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

	"github.com/stretchr/testify/assert"
)

type countUpdate struct {
	ID    string
	Count uint64
}

type clientContribution struct {
	Contributor interface{}
	Counts      map[string]uint64
}

func TestCounter_AggregateContributions(t *testing.T) {
	tests := []struct {
		Contributions []clientContribution
		Aggregates    map[string]uint64
	}{
		// Two contributors with one update each.
		{
			Contributions: []clientContribution{
				{
					Contributor: 1,
					Counts: map[string]uint64{
						"a": 1,
						"b": 2,
					},
				},
				{
					Contributor: 2,
					Counts: map[string]uint64{
						"b": 2,
						"c": 3,
					},
				},
			},
			Aggregates: map[string]uint64{
				"a": 1,
				"b": 4,
				"c": 3,
			},
		},
		// Contributors with multiple updates.
		{
			Contributions: []clientContribution{
				{
					Contributor: 1,
					Counts: map[string]uint64{
						"a": 1,
						"b": 2,
					},
				},
				{
					Contributor: 1,
					Counts: map[string]uint64{
						"b": 2,
						"c": 3,
					},
				},
				{
					Contributor: 2,
					Counts: map[string]uint64{
						"b": 4,
						"c": 5,
					},
				},
			},
			Aggregates: map[string]uint64{
				"b": 6,
				"c": 8,
			},
		},
		// Contributors unregister.
		{
			Contributions: []clientContribution{
				{
					Contributor: 1,
					Counts: map[string]uint64{
						"a": 1,
						"b": 2,
					},
				},
				{
					// Unregister.
					Contributor: 1,
				},
				{
					Contributor: 2,
					Counts: map[string]uint64{
						"b": 4,
						"c": 5,
					},
				},
			},
			Aggregates: map[string]uint64{
				"b": 4,
				"c": 5,
			},
		},
	}

	for _, tt := range tests {
		counter := newCounter()
		for _, c := range tt.Contributions {
			if c.Counts != nil {
				counter.Register(c.Contributor, c.Counts)
			} else {
				counter.Unregister(c.Contributor)
			}
		}
		assert.Equal(t, tt.Aggregates, counter.Aggregates())
	}
}

func TestCounter_Subscribe(t *testing.T) {
	counter := newCounter()

	var updates []countUpdate
	counter.Subscribe(func(id string, count uint64) {
		updates = append(updates, countUpdate{
			ID:    id,
			Count: count,
		})
	})

	counter.Register(1, map[string]uint64{
		"a": 1,
	})
	counter.Register(1, map[string]uint64{
		"a": 1,
		"b": 2,
	})
	counter.Register(1, map[string]uint64{
		"b": 2,
	})
	counter.Register(1, map[string]uint64{
		"b": 2,
		"c": 3,
	})
	counter.Register(1, map[string]uint64{
		"b": 1,
		"c": 3,
	})
	counter.Register(2, map[string]uint64{
		"b": 1,
	})
	counter.Register(1, map[string]uint64{
		"c": 3,
	})

	expected := []countUpdate{
		{"a", 1},
		{"b", 2},
		{"a", 0},
		{"c", 3},
		{"b", 1},
		{"b", 2},
		{"b", 1},
	}
	assert.Equal(t, expected, updates)
}
