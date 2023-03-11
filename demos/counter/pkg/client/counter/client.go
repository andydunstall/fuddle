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
	"sync"
)

// Client is a client to the counter service.
//
// Note starting with simply routing all streams to the first node.
type Client struct {
	counters map[string]*counter

	// mu is a mutex that protects the fields above.
	mu sync.Mutex

	addr string
}

// NewClient connects to the counter service and streams the count for
// subscribed IDs.
func NewClient(addr string) *Client {
	return &Client{
		counters: make(map[string]*counter),
		addr:     addr,
	}
}

// Register registers the user for the given ID, which will call onUpdate
// whenever the count changes. Returns a function to unregister.
func (c *Client) Register(id string, onUpdate func(c uint64)) (func(), error) {
	counter, err := c.counter(id)
	if err != nil {
		return nil, err
	}
	return counter.Register(onUpdate)
}

func (c *Client) Close() {
	for _, counter := range c.counters {
		counter.Close()
	}
}

func (c *Client) counter(id string) (*counter, error) {
	c.mu.Lock()
	defer c.mu.Unlock()

	counter, ok := c.counters[id]
	if !ok {
		var err error
		counter, err = newCounter(id, c.addr)
		if err != nil {
			return nil, err
		}
		c.counters[id] = counter
	}
	return counter, nil
}
