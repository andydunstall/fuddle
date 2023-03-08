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
	"time"

	"github.com/andydunstall/fuddle/pkg/build"
	fuddle "github.com/andydunstall/fuddle/pkg/sdkv2"
	"go.uber.org/zap"
)

type Service struct {
	server *server

	registry     *fuddle.Registry
	loadBalancer *loadBalancer

	conf   *Config
	logger *zap.Logger
}

func NewService(conf *Config, logger *zap.Logger) *Service {
	logger = logger.With(zap.String("service", "frontend"))

	loadBalancer := newLoadBalancer()
	server := newServer(conf.Addr, loadBalancer, logger)
	return &Service{
		server:       server,
		loadBalancer: loadBalancer,
		registry:     nil,
		conf:         conf,
		logger:       logger,
	}
}

func (s *Service) Start() error {
	registry, err := fuddle.Register([]string{"localhost:8220"},
		fuddle.NodeState{
			ID:       s.conf.ID,
			Service:  "frontend",
			Locality: "aws.us-east-1.us-east-1-a",
			Created:  time.Now().UnixMilli(),
			Revision: build.Revision,
			State: map[string]string{
				"addr": s.conf.Addr,
			},
		},
	)
	if err != nil {
		return err
	}

	// Subscribe to only random nodes.
	filter := fuddle.Filter{
		"random": {},
	}
	registry.Subscribe(func(nodes []fuddle.NodeState) {
		addrs := []string{}
		for _, node := range nodes {
			addrs = append(addrs, node.State["addr"])
		}
		s.loadBalancer.SetAddrs(addrs)
	}, fuddle.WithFilter(filter))

	s.registry = registry

	return s.server.Start()
}

func (s *Service) GracefulStop() {
	s.server.GracefulStop()
	if s.registry != nil {
		if err := s.registry.Unregister(); err != nil {
			s.logger.Error("failed to unregister", zap.Error(err))
		}
	}
}
