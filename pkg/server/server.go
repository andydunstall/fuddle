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
	conf   *config.Config
	logger *zap.Logger
}

func NewServer(conf *config.Config, logger *zap.Logger) *Server {
	logger = logger.With(zap.String("service", "server"))
	return &Server{
		conf:   conf,
		logger: logger,
	}
}

func (s *Server) Start() error {
	s.logger.Info("starting node", zap.Object("conf", s.conf))

	return nil
}

func (s *Server) GracefulShutdown() {
	s.logger.Info("starting node graceful shutdown")
}
