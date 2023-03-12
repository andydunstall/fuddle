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

	"go.uber.org/atomic"
)

type counter struct {
	id string

	// localCount is the number of users registered for the ID on this node.
	localCount *atomic.Uint64

	conn *connection
}

func newCounter(id string, addr string) (*counter, error) {
	conn, err := connect(addr)
	if err != nil {
		return nil, fmt.Errorf("counter: %w", err)
	}

	return &counter{
		id:         id,
		localCount: atomic.NewUint64(0),
		conn:       conn,
	}, nil
}

func (c *counter) Register(onUpdate func(c uint64)) (func() error, error) {
	c.localCount.Inc()

	handle := c.conn.Subscribe(c.id, onUpdate)

	if err := c.conn.Send(c.id, c.localCount.Load()); err != nil {
		return nil, fmt.Errorf("counter: %w", err)
	}
	return func() error {
		c.conn.Unsubscribe(handle)
		c.localCount.Dec()

		if err := c.conn.Send(c.id, c.localCount.Load()); err != nil {
			return fmt.Errorf("counter: %w", err)
		}
		return nil
	}, nil
}

func (c *counter) Close() {
	c.conn.Close()
}
