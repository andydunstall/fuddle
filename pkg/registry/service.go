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
	"github.com/andydunstall/fuddle/pkg/config"
	"github.com/prometheus/client_golang/prometheus"
	"go.uber.org/zap"
)

type Service struct {
	nodeMap *NodeMap
	server  *Server

	logger *zap.Logger
}

func NewService(conf *config.Config, metricsRegistry *prometheus.Registry, logger *zap.Logger) *Service {
	logger = logger.With(zap.String("service", "registry"))

	nodeCountGauge := prometheus.NewGauge(prometheus.GaugeOpts{
		Name: "fuddle_registry_node_count",
		Help: "Number of nodes registered with Fuddle",
	})
	metricsRegistry.MustRegister(nodeCountGauge)

	nodeMap := NewNodeMap()
	nodeMap.Subscribe("registry.metrics", func() {
		nodeCountGauge.Set(float64(len(nodeMap.NodeIDs())))
	})

	server := NewServer(nodeMap)
	return &Service{
		nodeMap: nodeMap,
		server:  server,
		logger:  logger,
	}
}

func (s *Service) Start() error {
	s.logger.Info("starting registry service")
	return nil
}

func (s *Service) GracefulStop() {
	s.logger.Info("starting registry service graceful shutdown")
}

func (s *Service) Server() *Server {
	return s.server
}

func (s *Service) NodeMap() *NodeMap {
	return s.nodeMap
}
