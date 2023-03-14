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

	"github.com/fuddle-io/fuddle/pkg/registry/cluster"
	"github.com/gorilla/websocket"
	"go.uber.org/zap"
)

// pendingUpdates stores a list of updates waiting to be sent to the connected
// client.
type pendingUpdates struct {
	updates []*cluster.NodeUpdate

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
func (p *pendingUpdates) Wait() ([]*cluster.NodeUpdate, bool) {
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
func (p *pendingUpdates) Push(update *cluster.NodeUpdate) {
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

type conn struct {
	ws      *websocket.Conn
	pending *pendingUpdates

	logger *zap.Logger

	wg sync.WaitGroup
}

func newConn(ws *websocket.Conn, logger *zap.Logger) *conn {
	conn := &conn{
		ws:      ws,
		pending: newPendingUpdates(),
		logger:  logger,
	}

	conn.wg.Add(1)
	go conn.sendPending()

	return conn
}

func (c *conn) AddUpdate(update *cluster.NodeUpdate) {
	c.pending.Push(update)
}

func (c *conn) RecvUpdate() (*cluster.NodeUpdate, error) {
	_, b, err := c.ws.ReadMessage()
	if err != nil {
		return nil, fmt.Errorf("conn recv: %w", err)
	}
	var update cluster.NodeUpdate
	if err := json.Unmarshal(b, &update); err != nil {
		return nil, fmt.Errorf("conn recv: decode error: %w", err)
	}
	return &update, nil
}

func (c *conn) Close() {
	c.ws.Close()
	c.pending.Close()
	c.wg.Wait()
}

func (c *conn) sendPending() {
	defer c.wg.Done()

	for {
		updates, ok := c.pending.Wait()
		if !ok {
			return
		}

		for _, update := range updates {
			b, err := json.Marshal(update)
			if err != nil {
				c.logger.Error("encode update", zap.Error(err))
				continue
			}

			if err := c.ws.WriteMessage(websocket.BinaryMessage, b); err != nil {
				return
			}
		}
	}
}
