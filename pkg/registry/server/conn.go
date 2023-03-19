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
	"encoding/json"
	"fmt"
	"sync"

	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// pendingMessages stores a list of messages waiting to be sent to the connected
// client.
type pendingMessages struct {
	messages []*message

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
func (p *pendingMessages) Wait() ([]*message, bool) {
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

	messages := p.messages
	p.messages = nil
	return messages, true
}

// Push an message to be sent.
func (p *pendingMessages) Push(message *message) {
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

type conn struct {
	ws      *websocket.Conn
	pending *pendingMessages

	logger *zap.Logger

	wg sync.WaitGroup
}

func newConn(ws *websocket.Conn, logger *zap.Logger) *conn {
	conn := &conn{
		ws:      ws,
		pending: newPendingMessages(),
		logger:  logger,
	}

	conn.wg.Add(1)
	go conn.sendPending()

	return conn
}

func (c *conn) AddMessage(m *message) {
	c.pending.Push(m)
}

func (c *conn) RecvMessage() (*message, error) {
	_, b, err := c.ws.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("conn recv: %w", err)
	}
	var m message
	if err := json.Unmarshal(b, &m); err != nil {
		return nil, fmt.Errorf("conn recv: decode error: %w", err)
	}
	return &m, nil
}

func (c *conn) Close() {
	c.ws.Close()
	c.pending.Close()
	c.wg.Wait()
}

func (c *conn) sendPending() {
	defer c.wg.Done()

	for {
		messages, ok := c.pending.Wait()
		if !ok {
			return
		}

		for _, m := range messages {
			b, err := json.Marshal(m)
			if err != nil {
				c.logger.Error("encode message", zap.Error(err))
				continue
			}

			if err := c.ws.WriteMessage(websocket.BinaryMessage, b); err != nil {
				return
			}
		}
	}
}
