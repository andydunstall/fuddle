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
	"context"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"github.com/fuddle-io/fuddle/pkg/registryv2/registry"
	"go.uber.org/zap"
)

type Server struct {
	registry *registry.Registry

	logger *zap.Logger
}

func NewServer(registry *registry.Registry, opts ...Option) *Server {
	options := options{
		logger: zap.NewNop(),
	}
	for _, o := range opts {
		o.apply(&options)
	}

	return &Server{
		registry: registry,
		logger:   options.logger,
	}
}

func (s *Server) RegisterV2(ctx context.Context, req *rpc.RegisterRequest) (*rpc.RegisterResponse, error) {
	logger := s.logger.With(zap.String("rpc", "registry.RegisterV2"))

	err := s.registry.Register(req.Node)
	if err == registry.ErrAlreadyRegistered {
		logger.Warn(
			"node already registered",
			zap.String("id", req.Node.Id),
		)

		return &rpc.RegisterResponse{
			Error: &rpc.Error{
				Status:      rpc.ErrorStatus_ALREADY_REGISTERED,
				Description: err.Error(),
			},
		}, nil
	}
	if err == registry.ErrInvalidUpdate {
		logger.Warn(
			"invalid node",
			zap.String("id", req.Node.Id),
		)

		return &rpc.RegisterResponse{
			Error: &rpc.Error{
				Status:      rpc.ErrorStatus_PROTOCOL,
				Description: err.Error(),
			},
		}, nil
	}
	if err != nil {
		logger.Error(
			"unknown error",
			zap.String("id", req.Node.Id),
		)

		return &rpc.RegisterResponse{
			Error: &rpc.Error{
				Status:      rpc.ErrorStatus_UNKNOWN,
				Description: err.Error(),
			},
		}, nil
	}

	logger.Debug(
		"node registered",
		zap.String("id", req.Node.Id),
	)

	return &rpc.RegisterResponse{}, nil
}

func (s *Server) Unregister(ctx context.Context, req *rpc.UnregisterRequest) (*rpc.UnregisterResponse, error) {
	logger := s.logger.With(zap.String("rpc", "registry.Unregister"))

	if err := s.registry.Unregister(req.NodeId); err != nil {
		logger.Error(
			"unknown error",
			zap.String("id", req.NodeId),
		)

		return &rpc.UnregisterResponse{
			Error: &rpc.Error{
				Status:      rpc.ErrorStatus_UNKNOWN,
				Description: err.Error(),
			},
		}, nil
	}
	return &rpc.UnregisterResponse{}, nil
}

func (s *Server) UpdateNode(ctx context.Context, req *rpc.UpdateNodeRequest) (*rpc.UpdateNodeResponse, error) {
	logger := s.logger.With(zap.String("rpc", "registry.Unregister"))

	err := s.registry.UpdateNode(req.NodeId, req.Metadata)

	if err == registry.ErrNotFound {
		logger.Warn(
			"node not found",
			zap.String("id", req.NodeId),
		)

		return &rpc.UpdateNodeResponse{
			Error: &rpc.Error{
				Status:      rpc.ErrorStatus_NOT_FOUND,
				Description: err.Error(),
			},
		}, nil
	}
	if err == registry.ErrInvalidUpdate {
		logger.Warn(
			"invalid update",
			zap.String("id", req.NodeId),
		)

		return &rpc.UpdateNodeResponse{
			Error: &rpc.Error{
				Status:      rpc.ErrorStatus_PROTOCOL,
				Description: err.Error(),
			},
		}, nil
	}
	if err != nil {
		logger.Error(
			"unknown error",
			zap.String("id", req.NodeId),
		)

		return &rpc.UpdateNodeResponse{
			Error: &rpc.Error{
				Status:      rpc.ErrorStatus_UNKNOWN,
				Description: err.Error(),
			},
		}, nil
	}

	return &rpc.UpdateNodeResponse{}, nil
}

func (s *Server) Updates(req *rpc.UpdatesRequest, stream rpc.Registry_UpdatesServer) error {
	return nil
}

func (s *Server) Node(ctx context.Context, req *rpc.NodeRequest) (*rpc.NodeResponse, error) {
	logger := s.logger.With(zap.String("rpc", "registry.Node"))

	n, err := s.registry.Node(req.NodeId)
	if err == registry.ErrNotFound {
		logger.Debug(
			"node not found",
			zap.String("id", req.NodeId),
		)

		return &rpc.NodeResponse{
			Error: &rpc.Error{
				Status:      rpc.ErrorStatus_NOT_FOUND,
				Description: err.Error(),
			},
		}, nil
	}
	if err != nil {
		logger.Error(
			"unknown error",
			zap.String("id", req.NodeId),
		)

		return &rpc.NodeResponse{
			Error: &rpc.Error{
				Status:      rpc.ErrorStatus_UNKNOWN,
				Description: err.Error(),
			},
		}, nil
	}

	logger.Debug(
		"node found",
		zap.String("id", req.NodeId),
	)

	return &rpc.NodeResponse{
		Node: n,
	}, nil
}

func (s *Server) Nodes(ctx context.Context, req *rpc.NodesRequest) (*rpc.NodesResponse, error) {
	nodes := s.registry.Nodes(req.IncludeMetadata)
	return &rpc.NodesResponse{
		Nodes: nodes,
	}, nil
}
