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
	"context"
	"fmt"
	"sync"

	"github.com/andydunstall/fuddle/demos/counter/pkg/rpc"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

type subHandle struct {
	ID       string
	Callback func(c uint64)
}

// connection is a connection to a counter server.
type connection struct {
	subscribers map[*subHandle]interface{}

	// mu is a mutex that protects the fields above.
	mu sync.Mutex

	conn   *grpc.ClientConn
	stream rpc.Counter_StreamClient

	wg sync.WaitGroup
}

func connect(addr string) (*connection, error) {
	conn, err := grpc.Dial(
		addr, grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	client := rpc.NewCounterClient(conn)
	stream, err := client.Stream(context.Background())
	if err != nil {
		return nil, fmt.Errorf("connect: %w", err)
	}

	c := &connection{
		subscribers: make(map[*subHandle]interface{}),
		conn:        conn,
		stream:      stream,
	}

	c.wg.Add(1)
	go func() {
		defer c.wg.Done()
		c.readLoop()
	}()

	return c, nil
}

func (c *connection) Send(id string, count uint64) error {
	if err := c.stream.Send(&rpc.CountUpdate{
		Id:    id,
		Count: count,
	}); err != nil {
		return fmt.Errorf("connection send: %w", err)
	}
	return nil
}

func (c *connection) Subscribe(id string, cb func(c uint64)) *subHandle {
	c.mu.Lock()
	defer c.mu.Unlock()

	handle := &subHandle{ID: id, Callback: cb}
	c.subscribers[handle] = struct{}{}
	return handle
}

func (c *connection) Unsubscribe(handle *subHandle) {
	c.mu.Lock()
	defer c.mu.Unlock()

	delete(c.subscribers, handle)
}

func (c *connection) Close() {
	c.conn.Close()
	c.wg.Wait()
}

func (c *connection) readLoop() {
	for {
		update, err := c.stream.Recv()
		if err != nil {
			return
		}

		c.mu.Lock()
		for handle := range c.subscribers {
			if update.Id == handle.ID {
				handle.Callback(update.Count)
			}
		}
		c.mu.Unlock()
	}

}
