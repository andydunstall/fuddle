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
	"fmt"
	"time"

	fuddle "github.com/andydunstall/fuddle/pkg/sdk"
	"go.uber.org/zap"
)

type Service struct {
	conf *Config

	registry *fuddle.Registry

	logger *zap.Logger
}

func NewService(conf *Config, opts ...Option) *Service {
	options := options{
		logger: zap.NewNop(),
	}
	for _, o := range opts {
		o.apply(&options)
	}

	logger := options.logger.With(zap.String("service", "counter"))
	return &Service{
		conf:   conf,
		logger: logger,
	}
}

func (s *Service) Start() error {
	registry, err := fuddle.Register(
		s.conf.FuddleAddrs,
		fuddle.Node{
			ID:       s.conf.ID,
			Service:  "frontend",
			Locality: s.conf.Locality,
			Created:  time.Now().UnixMilli(),
			Revision: s.conf.Revision,
			State: map[string]string{
				"addr.rpc": s.conf.RPCAddr,
			},
		},
	)
	if err != nil {
		return fmt.Errorf("frontend service: start: %w", err)
	}
	s.registry = registry

	return nil
}

func (s *Service) GracefulStop() {
	if err := s.registry.Unregister(); err != nil {
		s.logger.Error("failed to unregister", zap.Error(err))
	}
}
