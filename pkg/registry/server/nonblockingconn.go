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

package server

import (
	"fmt"
	"sync"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
)

// pendingMessages stores a list of messages waiting to be sent to the connected
// client.
type pendingMessages struct {
	messages []*rpc.Message

	mu *sync.Mutex

	cv     *sync.Cond
	wg     sync.WaitGroup
	closed bool
}

func newPendingMessages() *pendingMessages {
	mu := &sync.Mutex{}
	return &pendingMessages{
		messages: nil,
		mu:       mu,
		cv:       sync.NewCond(mu),
		wg:       sync.WaitGroup{},
		closed:   false,
	}
}

// Wait blocks until the next message is available of the connection has been
// closed. If the returned bool is false the connection has been closed.
func (p *pendingMessages) Wait() ([]*rpc.Message, bool) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return nil, false
	}

	// Since we can miss signals when processing messages, must only block
	// if the messages are empty.
	if len(p.messages) == 0 {
		p.cv.Wait()
	}

	// Check if closed since blocking.
	if p.closed {
		return nil, false
	}

	messages := p.messages
	p.messages = nil
	return messages, true
}

// Push an message to be sent.
func (p *pendingMessages) Push(message *rpc.Message) {
	p.mu.Lock()
	defer p.mu.Unlock()

	if p.closed {
		return
	}

	p.messages = append(p.messages, message)
	p.cv.Signal()
}

func (p *pendingMessages) Close() {
	p.mu.Lock()
	defer p.mu.Unlock()

	p.closed = true
	// Signal to Wait returns.
	p.cv.Signal()
}

// nonBlockingConn is a wrapper around the underlying connection that ensures
// send does not block.
type nonBlockingConn struct {
	conn    rpc.Registry_RegisterServer
	pending *pendingMessages

	wg sync.WaitGroup
}

func newNonBlockingConn(conn rpc.Registry_RegisterServer) *nonBlockingConn {
	nonBlockingConn := &nonBlockingConn{
		conn:    conn,
		pending: newPendingMessages(),
	}

	nonBlockingConn.wg.Add(1)
	go nonBlockingConn.sendPending()

	return nonBlockingConn
}

// Send sends the given message to the client without blocking.
func (c *nonBlockingConn) Send(m *rpc.Message) error {
	c.pending.Push(m)
	return nil
}

func (c *nonBlockingConn) Recv() (*rpc.Message, error) {
	m, err := c.conn.Recv()
	if err != nil {
		return nil, fmt.Errorf("conn recv: %w", err)
	}
	return m, nil
}

func (c *nonBlockingConn) Close() {
	c.pending.Close()
	c.wg.Wait()
}

func (c *nonBlockingConn) sendPending() {
	defer c.wg.Done()

	for {
		messages, ok := c.pending.Wait()
		if !ok {
			return
		}

		for _, m := range messages {
			if err := c.conn.Send(m); err != nil {
				return
			}
		}
	}
}
