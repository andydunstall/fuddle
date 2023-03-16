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

package admin

import (
	"github.com/fuddle-io/fuddle/pkg/config"
	"github.com/fuddle-io/fuddle/pkg/registry/cluster"
	"go.uber.org/zap"
)

type Service struct {
	server *server

	logger *zap.Logger
}

func NewService(cluster *cluster.Cluster, conf *config.Config, opts ...Option) *Service {
	options := options{
		logger:       zap.NewNop(),
		listener:     nil,
		promRegistry: nil,
	}
	for _, o := range opts {
		o.apply(&options)
	}

	logger := options.logger.With(zap.String("service", "admin"))
	options.logger = logger

	server := newServer(conf.BindAdminAddr, cluster, options)
	return &Service{
		server: server,
		logger: logger,
	}
}

func (s *Service) Start() error {
	s.logger.Info("starting admin service")
	return s.server.Start()
}

func (s *Service) GracefulStop() {
	s.logger.Info("starting admin service graceful shutdown")
	s.server.GracefulStop()
}

func (s *Service) Stop() {
	s.logger.Info("starting admin service hard shutdown")
	s.server.Stop()
}
