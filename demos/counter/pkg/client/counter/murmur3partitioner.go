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
	"sort"
	"sync"

	"github.com/spaolacci/murmur3"
)

type hashedNode struct {
	ID   string
	Addr string
	Hash uint64
}

type Murmur3Partitioner struct {
	nodes []hashedNode

	// mu is a mutex protecting the above fields.
	mu sync.RWMutex
}

func NewMurmur3Partitioner() *Murmur3Partitioner {
	return &Murmur3Partitioner{}
}

func (p *Murmur3Partitioner) Locate(id string, onRelocate func(addr string)) (string, func(), bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(p.nodes) == 0 {
		return "", nil, false
	}

	hash := murmur3.Sum64([]byte(id))
	for i := len(p.nodes) - 1; i >= 0; i-- {
		if p.nodes[i].Hash >= hash {
			return p.nodes[i].Addr, nil, true
		}
	}
	return p.nodes[len(p.nodes)-1].Addr, nil, true
}

func (p *Murmur3Partitioner) SetNodes(nodes map[string]string) {
	var hashedNodes []hashedNode
	for id, addr := range nodes {
		hashedNodes = append(hashedNodes, hashedNode{
			ID:   id,
			Addr: addr,
			Hash: murmur3.Sum64([]byte(id)),
		})
	}
	sort.Slice(hashedNodes, func(i, j int) bool {
		return hashedNodes[i].Hash < hashedNodes[j].Hash
	})

	p.mu.Lock()
	defer p.mu.Unlock()

	p.nodes = hashedNodes
}

var _ Partitioner = &Murmur3Partitioner{}
