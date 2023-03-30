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

package frontend

import (
	"context"
	"fmt"
	"net"
	"time"

	fuddle "github.com/fuddle-io/fuddle-go"
	"github.com/fuddle-io/fuddle/demos/counter/pkg/client/counter"
	"go.uber.org/zap"
)

type Service struct {
	server     *server
	wsListener net.Listener

	partitioner counter.Partitioner
	counter     *counter.Client

	conf *Config

	fuddleClient *fuddle.Fuddle

	logger *zap.Logger
}

func NewService(conf *Config, opts ...Option) *Service {
	options := options{
		logger:     zap.NewNop(),
		wsListener: nil,
	}
	for _, o := range opts {
		o.apply(&options)
	}

	logger := options.logger.With(zap.String("service", "counter"))

	partitioner := counter.NewMurmur3Partitioner()
	counter := counter.NewClient(partitioner)
	server := newServer(conf.WSAddr, counter, logger)
	return &Service{
		server:      server,
		wsListener:  options.wsListener,
		partitioner: partitioner,
		counter:     counter,
		conf:        conf,
		logger:      logger,
	}
}

func (s *Service) Start() error {
	fuddleClient, err := fuddle.Connect(context.Background(), s.conf.FuddleAddrs)
	if err != nil {
		return fmt.Errorf("frontend service: start: %w", err)
	}

	if err = fuddleClient.Register(context.Background(), fuddle.Member{
		ID:       s.conf.ID,
		Service:  "frontend",
		Locality: s.conf.Locality,
		Created:  time.Now().UnixMilli(),
		Revision: s.conf.Revision,
		Metadata: map[string]string{
			"addr.ws": s.conf.WSAddr,
		},
	}); err != nil {
		return fmt.Errorf("frontend service: start: %w", err)
	}
	s.fuddleClient = fuddleClient

	// Subscribe to counter members updates and add to the partitioner.
	fuddleClient.Subscribe(func() {
		filter := fuddle.Filter{
			"counter": {},
		}

		counterMembers := make(map[string]string)
		for _, member := range fuddleClient.Members(fuddle.WithFilter(&filter)) {
			counterMembers[member.ID] = member.Metadata["addr.rpc"]
		}
		s.partitioner.SetNodes(counterMembers)
	})

	return s.server.Start(s.wsListener)
}

func (s *Service) GracefulStop() {
	s.fuddleClient.Close()
	s.server.GracefulStop()
	s.counter.Close()
}

func (s *Service) Stop() {
	s.fuddleClient.Close()
	s.server.Stop()
	s.counter.Close()
}
