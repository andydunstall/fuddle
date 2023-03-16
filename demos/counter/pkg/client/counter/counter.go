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
	"sync"

	"go.uber.org/atomic"
)

type regHandle struct {
	OnUpdate   func(c uint64)
	OnError    func(e error)
	connHandle *subHandle
}

type counter struct {
	registered map[*regHandle]interface{}

	// mu is a mutex that protects the fields above.
	mu sync.Mutex

	id string

	// localCount is the number of users registered for the ID on this node.
	localCount *atomic.Uint64

	closePartitioner func()
	conn             *connection
}

func newCounter(id string, partitioner Partitioner) (*counter, error) {
	counter := &counter{
		registered: make(map[*regHandle]interface{}),
		id:         id,
		localCount: atomic.NewUint64(0),
	}

	addr, closePartitioner, ok := partitioner.Locate(id, counter.onRelocate)
	if !ok {
		return nil, fmt.Errorf("no available backends")
	}
	counter.closePartitioner = closePartitioner

	conn, err := connect(addr.Addr)
	if err != nil {
		return nil, fmt.Errorf("counter: %w", err)
	}
	counter.conn = conn

	return counter, nil
}

func (c *counter) Register(onUpdate func(c uint64), onError func(e error)) (func() error, error) {
	c.localCount.Inc()

	connHandle := c.conn.Subscribe(c.id, onUpdate)
	handle := &regHandle{
		OnUpdate:   onUpdate,
		OnError:    onError,
		connHandle: connHandle,
	}
	c.mu.Lock()
	c.registered[handle] = struct{}{}
	c.mu.Unlock()

	if err := c.conn.Send(c.id, c.localCount.Load()); err != nil {
		return nil, fmt.Errorf("counter: %w", err)
	}
	return func() error {
		c.mu.Lock()
		c.conn.Unsubscribe(handle.connHandle)
		delete(c.registered, handle)
		c.mu.Unlock()

		c.localCount.Dec()

		if err := c.conn.Send(c.id, c.localCount.Load()); err != nil {
			return fmt.Errorf("counter: %w", err)
		}
		return nil
	}, nil
}

func (c *counter) Close() {
	c.closePartitioner()
	c.conn.Close()
}

func (c *counter) onRelocate(addr NodeAddr, ok bool) {
	var handles []*regHandle

	c.mu.Lock()
	for handle := range c.registered {
		c.conn.Unsubscribe(handle.connHandle)
		handle.connHandle = nil
		handles = append(handles, handle)
	}
	c.mu.Unlock()

	if !ok {
		for _, handle := range handles {
			handle.OnError(fmt.Errorf("no nodes available"))
		}
		return
	}

	c.conn.Close()

	conn, err := connect(addr.Addr)
	if err != nil {
		for _, handle := range handles {
			handle.OnError(fmt.Errorf("failed to connect"))
		}
	}
	c.conn = conn

	if err := c.conn.Send(c.id, c.localCount.Load()); err != nil {
		for _, handle := range handles {
			handle.OnError(fmt.Errorf("failed to update count"))
		}
	}
}
