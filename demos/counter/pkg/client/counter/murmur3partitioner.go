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
	"fmt"
	"sort"
	"sync"

	"github.com/spaolacci/murmur3"
)

type relocateHandle struct {
	ID          string
	Callback    func(addr string, ok bool)
	CurrentAddr string
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

func (p *Murmur3Partitioner) Locate(id string, onRelocate func(addr string, ok bool)) (string, func(), bool) {
	p.mu.RLock()
	defer p.mu.RUnlock()

	if len(p.nodes) == 0 {
		return "", nil, false
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
			handle.Callback("", ok)
		}
	}
}

func (p *Murmur3Partitioner) locateLocked(id string) (string, bool) {
	if len(p.nodes) == 0 {
		return "", false
	}

	hash := murmur3.Sum64([]byte(id))
	for i := len(p.nodes) - 1; i >= 0; i-- {
		if p.nodes[i].Hash >= hash {
			return p.nodes[i].Addr, true
		}
	}
	return p.nodes[len(p.nodes)-1].Addr, true
}

var _ Partitioner = &Murmur3Partitioner{}
