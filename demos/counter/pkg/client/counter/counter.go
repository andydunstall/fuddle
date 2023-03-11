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

type regHandle struct {
	Callback func(c uint64)
}

type counter struct {
	registered map[*regHandle]interface{}
	localCount uint64

	// mu is a mutex that protects the fields above.
	mu sync.Mutex

	id string

	// updateCh is a channel to tell the write loop to send an updated count to
	// the server.
	updateCh chan interface{}

	conn   *grpc.ClientConn
	stream rpc.Counter_StreamClient

	wg sync.WaitGroup
}

func newCounter(id string, addr string) (*counter, error) {
	conn, err := grpc.Dial(
		addr, grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, fmt.Errorf("counter: connect: %w", err)
	}

	client := rpc.NewCounterClient(conn)
	stream, err := client.Stream(context.Background())
	if err != nil {
		return nil, fmt.Errorf("counter: stream: %w", err)
	}

	c := &counter{
		registered: make(map[*regHandle]interface{}),
		localCount: 0,
		id:         id,
		updateCh:   make(chan interface{}),
		conn:       conn,
		stream:     stream,
	}

	c.wg.Add(2)
	go func() {
		defer c.wg.Done()
		c.readLoop()
	}()
	go func() {
		defer c.wg.Done()
		c.writeLoop()
	}()

	return c, nil
}

func (c *counter) Register(onUpdate func(c uint64)) (func(), error) {
	c.mu.Lock()

	c.localCount++

	handle := &regHandle{Callback: onUpdate}
	c.registered[handle] = struct{}{}

	c.mu.Unlock()

	c.updateCh <- struct{}{}

	return func() {
		c.unregister(handle)
	}, nil
}

func (c *counter) Close() {
	close(c.updateCh)
	c.conn.Close()
	c.wg.Wait()
}

func (c *counter) unregister(handle *regHandle) {
	c.mu.Lock()
	defer c.mu.Unlock()

	c.localCount--
	delete(c.registered, handle)

	c.updateCh <- struct{}{}
}

func (c *counter) writeLoop() {
	for range c.updateCh {
		var count uint64
		c.mu.Lock()
		count = c.localCount
		c.mu.Unlock()

		if err := c.stream.Send(&rpc.CountUpdate{
			Id:    c.id,
			Count: count,
		}); err != nil {
			return
		}
	}
}

func (c *counter) readLoop() {
	for {
		update, err := c.stream.Recv()
		if err != nil {
			return
		}

		c.mu.Lock()
		for handle := range c.registered {
			handle.Callback(update.Count)
		}
		c.mu.Unlock()
	}
}
