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
	"context"

	"github.com/andydunstall/fuddle/pkg/rpc"
	"go.uber.org/zap"
)

// Server exposes gRPC endpoints for the registry.
type Server struct {
	nodeMap *NodeMap

	logger *zap.Logger

	rpc.UnimplementedRegistryServer
}

func NewServer(nodeMap *NodeMap, logger *zap.Logger) *Server {
	return &Server{
		nodeMap: nodeMap,
		logger:  logger,
	}
}

func (s *Server) Register(ctx context.Context, req *rpc.RegisterRequest) (*rpc.RegisterResponse, error) {
	s.logger.Debug("register", zap.String("node-id", req.NodeId))

	s.nodeMap.Register(req.NodeId)
	return &rpc.RegisterResponse{}, nil
}

func (s *Server) Unregister(ctx context.Context, req *rpc.RegisterRequest) (*rpc.RegisterResponse, error) {
	s.logger.Debug("unregister", zap.String("node-id", req.NodeId))

	s.nodeMap.Unregister(req.NodeId)
	return &rpc.RegisterResponse{}, nil
}

func (s *Server) Nodes(ctx context.Context, req *rpc.NodesRequest) (*rpc.NodesResponse, error) {
	return &rpc.NodesResponse{
		Ids: s.nodeMap.NodeIDs(),
	}, nil
}
