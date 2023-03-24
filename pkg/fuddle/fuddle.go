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

package fuddle

import (
	"net"
	"time"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"github.com/fuddle-io/fuddle/pkg/config"
	"github.com/fuddle-io/fuddle/pkg/registry"
	"go.uber.org/zap"
	"google.golang.org/grpc"
)

// Fuddle runs a Fuddle node.
type Fuddle struct {
	registryServer *registry.Server

	grpcServer *grpc.Server
	grpcLn     net.Listener

	conf   *config.Config
	logger *zap.Logger

	done chan interface{}
}

func New(conf *config.Config, opts ...Option) *Fuddle {
	options := options{
		logger:   zap.NewNop(),
		listener: nil,
	}
	for _, o := range opts {
		o.apply(&options)
	}

	logger := options.logger.With(zap.String("service", "server"))

	registryServer := registry.NewServer(
		registry.NewRegistry(),
		registry.WithLogger(logger),
	)

	grpcServer := grpc.NewServer()
	rpc.RegisterRegistryServer(grpcServer, registryServer)

	return &Fuddle{
		registryServer: registryServer,
		grpcServer:     grpcServer,
		grpcLn:         options.listener,
		conf:           conf,
		logger:         logger,
		done:           make(chan interface{}),
	}
}

// Start starts the Fuddle node in a background goroutine.
func (s *Fuddle) Start() error {
	s.logger.Info("starting fuddle node", zap.Object("conf", s.conf))

	ln := s.grpcLn
	if ln == nil {
		var err error
		ln, err = net.Listen("tcp", s.conf.BindRegistryAddr)
		if err != nil {
			return err
		}
	}

	go func() {
		if err := s.grpcServer.Serve(ln); err != nil {
			s.logger.Error("grpc serve", zap.Error(err))
		}
	}()

	go s.livenessChecks()

	return nil
}

func (s *Fuddle) GracefulStop() {
	s.logger.Info("node graceful stop")
	close(s.done)
	s.grpcServer.GracefulStop()
}

func (s *Fuddle) Stop() {
	s.logger.Info("node hard stop")
	close(s.done)
	s.grpcServer.Stop()
}

func (s *Fuddle) livenessChecks() {
	ticker := time.NewTicker(time.Second)
	defer ticker.Stop()
	for {
		select {
		case <-ticker.C:
			s.registryServer.UnregisterDownNodes()
		case <-s.done:
			return
		}
	}
}
