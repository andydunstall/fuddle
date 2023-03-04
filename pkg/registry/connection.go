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

package registry

import (
	"sync"

	"github.com/andydunstall/fuddle/pkg/rpc"
)

// pendingUpdates stores a list of updates waiting to be sent to the connected
// client.
type pendingUpdates struct {
	updates []*rpc.NodeUpdate

	mu *sync.Mutex

	cv     *sync.Cond
	wg     sync.WaitGroup
	closed bool
}

func newPendingUpdates() *pendingUpdates {
	mu := &sync.Mutex{}
	return &pendingUpdates{
		updates: nil,
		mu:      mu,
		cv:      sync.NewCond(mu),
		wg:      sync.WaitGroup{},
		closed:  false,
	}
}

// Wait blocks until the next update is available of the connection has been
// closed. If the returned bool is false the connection has been closed.
func (p *pendingUpdates) Wait() ([]*rpc.NodeUpdate, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil, false
	}

	// Since we can miss signals when processing updates, must only block
	// if the updates are empty.
	if len(p.updates) == 0 {
		p.cv.Wait()
	}

	updates := p.updates
	p.updates = nil
	return updates, true
}

// Push an update to be sent.
func (p *pendingUpdates) Push(update *rpc.NodeUpdate) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return
	}

	p.updates = append(p.updates, update)
	p.cv.Signal()
}

func (p *pendingUpdates) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.closed = true
	// Signal to Wait returns.
	p.cv.Signal()
}

type connection struct {
	stream  rpc.Registry_RegisterServer
	pending *pendingUpdates

	wg sync.WaitGroup
}

func newConnection(stream rpc.Registry_RegisterServer) *connection {
	conn := &connection{
		stream:  stream,
		pending: newPendingUpdates(),
	}

	conn.wg.Add(1)
	go conn.sendPending()

	return conn
}

func (c *connection) AddUpdate(update *rpc.NodeUpdate) {
	c.pending.Push(update)
}

func (c *connection) RecvUpdate() (*rpc.NodeUpdate, error) {
	return c.stream.Recv()
}

func (c *connection) Close() {
	c.pending.Close()
	c.wg.Wait()
}

func (c *connection) sendPending() {
	defer c.wg.Done()

	for {
		updates, ok := c.pending.Wait()
		if !ok {
			return
		}

		for _, update := range updates {
			if err := c.stream.Send(update); err != nil {
				return
			}
		}
	}
}
