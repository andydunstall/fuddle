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

type relocateHandle struct {
	ID          string
	Callback    func(addr NodeAddr, ok bool)
	CurrentAddr NodeAddr
}

type hashedNode struct {
	ID   string
	Addr string
	Hash uint64
}

type Murmur3Partitioner struct {
	nodes   []hashedNode
	handles map[*relocateHandle]interface{}

	// mu is a mutex protecting the above fields.
	mu sync.RWMutex
}

func NewMurmur3Partitioner() *Murmur3Partitioner {
	return &Murmur3Partitioner{
		handles: make(map[*relocateHandle]interface{}),
	}
}

func (p *Murmur3Partitioner) Locate(id string, onRelocate func(addr NodeAddr, ok bool)) (NodeAddr, func(), bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(p.nodes) == 0 {
		return NodeAddr{}, nil, false
	}

	handle := &relocateHandle{
		ID:       id,
		Callback: onRelocate,
	}
	if onRelocate != nil {
		p.handles[handle] = nil
	}

	addr, _ := p.locateLocked(id)
	handle.CurrentAddr = addr
	return addr, func() {
		p.mu.RLock()
		defer p.mu.RUnlock()

		delete(p.handles, handle)
	}, true
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

	for handle := range p.handles {
		addr, ok := p.locateLocked(handle.ID)
		if ok {
			if addr != handle.CurrentAddr {
				handle.CurrentAddr = addr
				handle.Callback(addr, true)
			}
		} else {
			handle.Callback(NodeAddr{}, false)
		}
	}
}

func (p *Murmur3Partitioner) locateLocked(id string) (NodeAddr, bool) {
	if len(p.nodes) == 0 {
		return NodeAddr{}, false
	}

	hash := murmur3.Sum64([]byte(id))
	idx := sort.Search(len(p.nodes), func(i int) bool {
		return p.nodes[i].Hash >= hash
	})
	// Cycled back to first node.
	if idx == len(p.nodes) {
		idx = 0
	}

	n := p.nodes[idx]
	return NodeAddr{
		ID:   n.ID,
		Addr: n.Addr,
	}, true
}

var _ Partitioner = &Murmur3Partitioner{}
