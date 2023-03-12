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

	"github.com/fuddle-io/fuddle/demos/counter/pkg/rpc"
	"go.uber.org/zap"
)

type server struct {
	counter *counter

	logger *zap.Logger

	rpc.UnimplementedCounterServer
}

func newServer(logger *zap.Logger) *server {
	return &server{
		counter: newCounter(),
		logger:  logger,
	}
}

// Stream sends and receives updates containing the count for different IDs.
//
// Clients will send the local count to the server, which is the number of
// users registered with the ID on the client node. Then the server
// aggregates the counts and broadcasts the global count to each client.
func (s *server) Stream(stream rpc.Counter_StreamServer) error {
	s.logger.Debug("stream connected")

	defer s.counter.Unregister(stream)

	counts := make(map[string]uint64)
	var mu sync.Mutex

	unsubscribe := s.counter.Subscribe(func(id string, count uint64) {
		// Only send updates for IDs that the client has contributed to.
		mu.Lock()
		if _, ok := counts[id]; !ok {
			mu.Unlock()
			return
		}
		mu.Unlock()

		s.logger.Debug(
			"send count",
			zap.String("id", id),
			zap.Uint64("count", count),
		)

		update := &rpc.CountUpdate{
			Id:    id,
			Count: count,
		}
		// nolint:errcheck
		// Ignore send errors. If send fails the stream will be aborted so
		// theres nothing to do.
		stream.Send(update)
	})
	defer unsubscribe()

	for {
		update, err := stream.Recv()
		if err != nil {
			return err
		}

		s.logger.Debug(
			"recv count",
			zap.String("id", update.Id),
			zap.Uint64("count", update.Count),
		)

		mu.Lock()
		counts[update.Id] = update.Count
		if counts[update.Id] == 0 {
			delete(counts, update.Id)
		}

		countsCopy := make(map[string]uint64)
		for id, count := range counts {
			countsCopy[id] = count
		}
		mu.Unlock()

		s.counter.Register(stream, countsCopy)
	}
}
