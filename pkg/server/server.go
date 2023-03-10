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
	"github.com/andydunstall/fuddle/pkg/admin"
	"github.com/andydunstall/fuddle/pkg/config"
	"github.com/andydunstall/fuddle/pkg/registry"
	"github.com/andydunstall/fuddle/pkg/rpc"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"go.uber.org/zap"
)

// Server runs a fuddle node.
type Server struct {
	adminService    *admin.Service
	registryService *registry.Service

	grpcServer *grpcServer

	conf   *config.Config
	logger *zap.Logger
}

func NewServer(conf *config.Config, opts ...Option) *Server {
	options := options{
		logger: zap.NewNop(),
	}
	for _, o := range opts {
		o.apply(&options)
	}

	logger := options.logger.With(zap.String("service", "server"))

	metricsRegistry := prometheus.NewRegistry()
	metricsRegistry.MustRegister(collectors.NewGoCollector())

	registryService := registry.NewService(conf, metricsRegistry, logger)
	adminService := admin.NewService(registryService.Cluster(), conf, metricsRegistry, logger)

	grpcServer := newGRPCServer(conf.BindAddr, logger)
	rpc.RegisterRegistryServer(grpcServer.GRPCServer(), registryService.Server())

	return &Server{
		adminService:    adminService,
		registryService: registryService,
		grpcServer:      grpcServer,
		conf:            conf,
		logger:          logger,
	}
}

// Start starts the node in a background goroutine.
func (s *Server) Start() error {
	s.logger.Info("starting node", zap.Object("conf", s.conf))

	if err := s.adminService.Start(); err != nil {
		return err
	}
	if err := s.registryService.Start(); err != nil {
		return err
	}
	if err := s.grpcServer.Start(); err != nil {
		return err
	}

	return nil
}

func (s *Server) GracefulStop() {
	s.logger.Info("starting node graceful shutdown")
	s.grpcServer.GracefulStop()
	s.registryService.GracefulStop()
	s.adminService.GracefulStop()
}
