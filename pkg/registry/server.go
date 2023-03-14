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
	"fmt"

	"github.com/fuddle-io/fuddle/pkg/registry/cluster"
	"github.com/fuddle-io/fuddle/pkg/rpc"
	"go.uber.org/zap"
)

// Server exposes gRPC endpoints for the registry.
type Server struct {
	cluster *cluster.Cluster

	logger *zap.Logger

	rpc.UnimplementedRegistryServer
}

func NewServer(cluster *cluster.Cluster, logger *zap.Logger) *Server {
	return &Server{
		cluster: cluster,
		logger:  logger,
	}
}

func (s *Server) Register(stream rpc.Registry_RegisterServer) error {
	return fmt.Errorf("unimplemented")
}
