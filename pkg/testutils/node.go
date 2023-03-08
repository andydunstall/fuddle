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

package testutils

import (
	"math/rand"
	"time"

	fuddle "github.com/andydunstall/fuddle/pkg/sdk"
	"github.com/google/uuid"
)

func WaitForNodes(registry *fuddle.Registry, count int) ([]fuddle.Node, error) {
	recvCh := make(chan interface{})
	var nodes []fuddle.Node
	unsubscribe := registry.Subscribe(func(n []fuddle.Node) {
		if len(n) == count {
			nodes = n
			close(recvCh)
		}
	})
	defer unsubscribe()

	if err := WaitWithTimeout(recvCh, time.Millisecond*500); err != nil {
		return nil, err
	}
	return nodes, nil
}

func WaitForNode(registry *fuddle.Registry, node fuddle.Node) error {
	recvCh := make(chan interface{})
	unsubscribe := registry.Subscribe(func(nodes []fuddle.Node) {
		for _, n := range nodes {
			if n.Equal(node) {
				close(recvCh)
			}
		}
	})
	defer unsubscribe()

	if err := WaitWithTimeout(recvCh, time.Millisecond*500); err != nil {
		return err
	}
	return nil
}

// RandomNode returns a node with random attributes and state.
func RandomNode() fuddle.Node {
	return fuddle.Node{
		ID:       uuid.New().String(),
		Service:  uuid.New().String(),
		Locality: uuid.New().String(),
		Created:  rand.Int63(),
		Revision: uuid.New().String(),
		State:    RandomState(),
	}
}

func RandomState() map[string]string {
	return map[string]string{
		uuid.New().String(): uuid.New().String(),
		uuid.New().String(): uuid.New().String(),
		uuid.New().String(): uuid.New().String(),
		uuid.New().String(): uuid.New().String(),
		uuid.New().String(): uuid.New().String(),
	}
}
