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
	"github.com/fuddle-io/fuddle/pkg/config"
	"github.com/fuddle-io/fuddle/pkg/registry"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/collectors"
	"go.uber.org/zap"
)

// Fuddle runs a Fuddle node.
type Fuddle struct {
	registry *registry.Service

	conf   *config.Config
	logger *zap.Logger
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

	promRegistry := prometheus.NewRegistry()
	promRegistry.MustRegister(collectors.NewGoCollector())

	registry := registry.NewService(
		conf,
		registry.WithListener(options.listener),
		registry.WithPromRegistry(promRegistry),
		registry.WithLogger(logger),
	)
	return &Fuddle{
		registry: registry,
		conf:     conf,
		logger:   logger,
	}
}

// Start starts the Fuddle node in a background goroutine.
func (s *Fuddle) Start() error {
	s.logger.Info("starting fuddle node", zap.Object("conf", s.conf))

	if err := s.registry.Start(); err != nil {
		return err
	}

	return nil
}

func (s *Fuddle) GracefulStop() {
	s.logger.Info("node graceful stop")
	s.registry.GracefulStop()
}

func (s *Fuddle) Stop() {
	s.logger.Info("node hard stop")
	s.registry.Stop()
}
