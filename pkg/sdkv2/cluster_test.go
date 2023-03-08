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
	"math/rand"
	"sort"
	"testing"

	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

// Tests adding a local node to the cluster and looking it up.
func TestCluster_LookupNode(t *testing.T) {
	node := NodeState{
		ID:       "123",
		Service:  "foo",
		Locality: "aws.us-east-1-b",
		Created:  123456,
		Revision: "v1.2.3",
		State: map[string]string{
			"foo": "bar",
		},
	}
	cluster := newCluster(node)

	nodes := cluster.Nodes()
	assert.Equal(t, []NodeState{node}, nodes)

	// Verify Nodes returns a copy of each node.
	nodes[0].ID = "456"
	nodes[0].State["foo"] = "abc"

	assert.Equal(t, []NodeState{node}, cluster.Nodes())
}

func TestCluster_NodesLookupWithFilter(t *testing.T) {
	tests := []struct {
		Filter        Filter
		AddedNodes    []NodeState
		FilteredNodes []NodeState
	}{
		// Filter no nodes.
		{
			Filter: Filter{
				"service-1": {
					Locality: []string{"eu-west-1-*", "eu-west-2-*"},
					State: StateFilter{
						"foo": []string{"bar", "car", "boo"},
					},
				},
				"service-2": {
					Locality: []string{"us-east-1-*", "eu-west-2-*"},
					State: StateFilter{
						"bar": []string{"bar", "car", "boo"},
					},
				},
			},
			AddedNodes: []NodeState{
				NodeState{
					ID:       "1",
					Service:  "service-1",
					Locality: "eu-west-2-c",
					State: map[string]string{
						"foo": "car",
					},
				},
				NodeState{
					ID:       "2",
					Service:  "service-2",
					Locality: "us-east-1-c",
					State: map[string]string{
						"bar": "boo",
					},
				},
			},
			FilteredNodes: []NodeState{
				NodeState{
					ID:       "1",
					Service:  "service-1",
					Locality: "eu-west-2-c",
					State: map[string]string{
						"foo": "car",
					},
				},
				NodeState{
					ID:       "2",
					Service:  "service-2",
					Locality: "us-east-1-c",
					State: map[string]string{
						"bar": "boo",
					},
				},
			},
		},

		// Filter partial nodes.
		{
			Filter: Filter{
				"service-1": {
					Locality: []string{"eu-west-1-*", "eu-west-2-*"},
					State: StateFilter{
						"foo": []string{"bar", "car", "boo"},
					},
				},
			},
			AddedNodes: []NodeState{
				NodeState{
					ID:       "1",
					Service:  "service-1",
					Locality: "eu-west-2-c",
					State: map[string]string{
						"foo": "car",
					},
				},
				NodeState{
					ID:       "2",
					Service:  "service-2",
					Locality: "us-east-1-c",
					State: map[string]string{
						"bar": "boo",
					},
				},
			},
			FilteredNodes: []NodeState{
				NodeState{
					ID:       "1",
					Service:  "service-1",
					Locality: "eu-west-2-c",
					State: map[string]string{
						"foo": "car",
					},
				},
			},
		},

		// Filter all nodes.
		{
			Filter: Filter{},
			AddedNodes: []NodeState{
				NodeState{
					ID:       "1",
					Service:  "service-1",
					Locality: "eu-west-2-c",
					State: map[string]string{
						"foo": "car",
					},
				},
				NodeState{
					ID:       "2",
					Service:  "service-2",
					Locality: "us-east-1-c",
					State: map[string]string{
						"bar": "boo",
					},
				},
			},
			FilteredNodes: []NodeState{},
		},
	}

	for _, tt := range tests {
		cluster := newCluster(tt.AddedNodes[0])
		for i := 1; i != len(tt.AddedNodes); i++ {
			assert.Nil(t, cluster.AddNode(tt.AddedNodes[i]))
		}

		nodes := cluster.Nodes(WithFilter(tt.Filter))
		assert.Equal(t, tt.FilteredNodes, nodes)
	}
}

func TestCluster_UpdateNodeLocalState(t *testing.T) {
	node := NodeState{
		ID:       "123",
		Service:  "foo",
		Locality: "aws.us-east-1-b",
		Created:  123456,
		Revision: "v1.2.3",
		State: map[string]string{
			"foo": "bar",
		},
	}
	cluster := newCluster(node)

	assert.Nil(t, cluster.UpdateLocalState(map[string]string{
		"foo": "car",
		"bar": "boo",
	}))

	node.State["foo"] = "car"
	node.State["bar"] = "boo"
	nodes := cluster.Nodes()
	assert.Equal(t, []NodeState{node}, nodes)
}

func TestCluster_AddNode(t *testing.T) {
	localNode := randomNode()
	cluster := newCluster(localNode)

	var addedNodes []NodeState
	for i := 0; i != 10; i++ {
		node := randomNode()
		assert.Nil(t, cluster.AddNode(node))
		addedNodes = append(addedNodes, node)
	}

	expectedNodes := append(addedNodes, localNode)
	// Sort to make comparison easier.
	sort.Slice(expectedNodes, func(i, j int) bool {
		return expectedNodes[i].ID < expectedNodes[j].ID
	})

	nodes := cluster.Nodes()
	// Sort to make comparison easier.
	sort.Slice(nodes, func(i, j int) bool {
		return nodes[i].ID < nodes[j].ID
	})

	assert.Equal(t, nodes, expectedNodes)
}

func TestCluster_AddNodeMissingID(t *testing.T) {
	cluster := newCluster(randomNode())

	assert.NotNil(t, cluster.AddNode(NodeState{}))
}

func TestCluster_RemoveNode(t *testing.T) {
	localNode := randomNode()
	cluster := newCluster(localNode)

	var addedNodes []NodeState
	for i := 0; i != 10; i++ {
		node := randomNode()
		assert.Nil(t, cluster.AddNode(node))
		addedNodes = append(addedNodes, node)
	}

	for _, node := range addedNodes {
		cluster.RemoveNode(node.ID)
	}

	assert.Equal(t, []NodeState{localNode}, cluster.Nodes())
}

func TestCluster_Subscribe(t *testing.T) {
	localNode := randomNode()
	cluster := newCluster(localNode)

	count := 0
	unsubscribe := cluster.Subscribe(func(nodes []NodeState) {
		count++
	})
	defer unsubscribe()

	node := randomNode()
	assert.Nil(t, cluster.AddNode(node))
	assert.Nil(t, cluster.UpdateState(node.ID, map[string]string{"foo": "bar"}))
	cluster.RemoveNode(node.ID)

	assert.Equal(t, 4, count)
}

// randomNode returns a node with random attributes and state.
func randomNode() NodeState {
	return NodeState{
		ID:       uuid.New().String(),
		Service:  uuid.New().String(),
		Locality: uuid.New().String(),
		Created:  rand.Int63(),
		Revision: uuid.New().String(),
		State: map[string]string{
			uuid.New().String(): uuid.New().String(),
			uuid.New().String(): uuid.New().String(),
			uuid.New().String(): uuid.New().String(),
			uuid.New().String(): uuid.New().String(),
			uuid.New().String(): uuid.New().String(),
		},
	}
}
