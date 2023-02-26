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
	"github.com/andydunstall/fuddle/pkg/config"
	"go.uber.org/zap"
)

// Server runs a fuddle node.
type Server struct {
	grpcServer *grpcServer

	conf   *config.Config
	logger *zap.Logger
}

func NewServer(conf *config.Config, logger *zap.Logger) *Server {
	logger = logger.With(zap.String("service", "server"))

	grpcServer := newGRPCServer(conf.BindAddr, logger)
	return &Server{
		grpcServer: grpcServer,
		conf:       conf,
		logger:     logger,
	}
}

// Start starts the node in a background goroutine.
func (s *Server) Start() error {
	s.logger.Info("starting node", zap.Object("conf", s.conf))

	if err := s.grpcServer.Start(); err != nil {
		return err
	}

	return nil
}

func (s *Server) GracefulStop() {
	s.logger.Info("starting node graceful shutdown")
	s.grpcServer.GracefulStop()
}
