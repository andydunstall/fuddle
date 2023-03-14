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

package registry

import (
	"time"

	"github.com/fuddle-io/fuddle/pkg/config"
	"github.com/fuddle-io/fuddle/pkg/registry/cluster"
	"github.com/fuddle-io/fuddle/pkg/registry/server"
	"go.uber.org/zap"
)

type Service struct {
	clusterState *cluster.Cluster
	server       *server.Server

	logger *zap.Logger
}

func NewService(conf *config.Config, logger *zap.Logger) *Service {
	logger = logger.With(zap.String("service", "registry"))

	clusterState := cluster.NewCluster(cluster.Node{
		ID:       conf.ID,
		Service:  "fuddle",
		Locality: conf.Locality,
		Created:  time.Now().UnixMilli(),
		Revision: conf.Revision,
		Metadata: map[string]string{
			"addr.rpc":   conf.BindAddr,
			"addr.admin": conf.BindAdminAddr,
		},
	})

	server := server.NewServer(conf.AdvAddr, clusterState, server.WithLogger(logger))
	return &Service{
		clusterState: clusterState,
		server:       server,
		logger:       logger,
	}
}

func (s *Service) Start() error {
	s.logger.Info("starting registry service")
	return s.server.Start()
}

func (s *Service) GracefulStop() {
	s.logger.Info("starting registry service graceful shutdown")
	s.server.GracefulStop()
}

func (s *Service) Cluster() *cluster.Cluster {
	return s.clusterState
}
