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

package frontend

import (
	"sync"
)

// loadBalancer returns addresses in a round robin strategy.
type loadBalancer struct {
	addrs []string
	idx   int

	mu sync.RWMutex
}

func newLoadBalancer() *loadBalancer {
	return &loadBalancer{
		addrs: nil,
		idx:   0,
	}
}

// Addr returns the next address in a round robin strategy.
func (l *loadBalancer) Addr() (string, bool) {
	l.mu.RLock()
	defer l.mu.RUnlock()

	if len(l.addrs) == 0 {
		return "", false
	}

	addr := l.addrs[l.idx]
	l.idx = (l.idx + 1) % len(l.addrs)
	return addr, true
}

// SetAddrs sets the addresses for the load balancer.
func (l *loadBalancer) SetAddrs(addrs []string) {
	l.mu.Lock()
	defer l.mu.Unlock()

	l.addrs = addrs
	l.idx = 0
}
