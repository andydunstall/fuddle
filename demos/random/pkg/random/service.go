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

package random

import (
	"context"
	"fmt"

	"github.com/andydunstall/fuddle/pkg/build"
	"github.com/andydunstall/fuddle/pkg/client"
	"github.com/andydunstall/fuddle/pkg/rpc"
	"go.uber.org/zap"
)

type Service struct {
	registry *client.Registry

	conf   *Config
	logger *zap.Logger
}

func NewService(conf *Config, logger *zap.Logger) *Service {
	logger = logger.With(zap.String("service", "random"))

	return &Service{
		registry: nil,
		conf:     conf,
		logger:   logger,
	}
}

func (s *Service) Start() error {
	registry, err := client.ConnectRegistry("127.0.0.1:8220")
	if err != nil {
		return fmt.Errorf("random service: %w", err)
	}

	state := make(map[string]string)
	state["addr"] = s.conf.Addr
	node := &rpc.NodeState{
		Id:       s.conf.ID,
		Service:  "random",
		Revision: build.Revision,
		State:    state,
	}
	if err = registry.Register(context.Background(), node); err != nil {
		return fmt.Errorf("random service: %w", err)
	}

	s.registry = registry

	return nil
}

func (s *Service) GracefulStop() {
	if s.registry != nil {
		if err := s.registry.Unregister(context.Background(), s.conf.ID); err != nil {
			s.logger.Error("failed to unregister", zap.Error(err))
		}
	}
}
