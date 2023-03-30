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
	"net"
	"time"

	fuddle "github.com/fuddle-io/fuddle-go"
	"github.com/fuddle-io/fuddle/demos/counter/pkg/rpc"
	"go.uber.org/zap"
)

// Service implements the counter service.
type Service struct {
	grpcServer *grpcServer

	rpcListener net.Listener

	fuddleClient *fuddle.Fuddle

	conf *Config

	logger *zap.Logger
}

func NewService(conf *Config, opts ...Option) *Service {
	options := options{
		logger:      zap.NewNop(),
		rpcListener: nil,
	}
	for _, o := range opts {
		o.apply(&options)
	}

	logger := options.logger.With(zap.String("service", "counter"))

	server := newServer(logger)

	grpcServer := newGRPCServer(conf.RPCAddr, logger)
	rpc.RegisterCounterServer(grpcServer.GRPCServer(), server)

	return &Service{
		grpcServer:  grpcServer,
		rpcListener: options.rpcListener,
		conf:        conf,
		logger:      logger,
	}
}

func (s *Service) Start() error {
	s.logger.Info("starting node", zap.Object("conf", s.conf))

	fuddleClient, err := fuddle.Connect(context.Background(), s.conf.FuddleAddrs)
	if err != nil {
		return fmt.Errorf("frontend service: start: %w", err)
	}

	if err = fuddleClient.Register(context.Background(), fuddle.Member{
		ID:       s.conf.ID,
		Service:  "counter",
		Locality: s.conf.Locality,
		Created:  time.Now().UnixMilli(),
		Revision: s.conf.Revision,
		Metadata: map[string]string{
			"addr.rpc": s.conf.RPCAddr,
		},
	}); err != nil {
		return fmt.Errorf("counter service: start: %w", err)
	}
	s.fuddleClient = fuddleClient

	return s.grpcServer.Start(s.rpcListener)
}

func (s *Service) GracefulStop() {
	s.logger.Info("starting node graceful shutdown")

	s.fuddleClient.Close()
	s.grpcServer.GracefulStop()
}

func (s *Service) Stop() {
	s.logger.Info("starting node hard shutdown")

	s.fuddleClient.Close()
	s.grpcServer.Stop()
}
