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

	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"go.uber.org/zap"
)

type Server struct {
	registry *Registry

	logger *zap.Logger

	rpc.UnimplementedRegistryServer
}

func NewServer(registry *Registry, opts ...Option) *Server {
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

func (s *Server) RegisterMember(ctx context.Context, req *rpc.RegisterMemberRequest) (*rpc.RegisterMemberResponse, error) {
	logger := s.logger.With(zap.String("rpc", "registry.RegisterMember"))

	err := s.registry.Register(req.Member)
	if err == ErrAlreadyRegistered {
		logger.Warn("member already registered", zap.String("id", req.Member.Id))
		return &rpc.RegisterMemberResponse{
			Error: errorResponse(rpc.ErrorStatus_ALREADY_REGISTERED, err.Error()),
		}, nil
	}
	if err == ErrInvalidUpdate {
		logger.Warn("invalid member", zap.String("id", req.Member.Id))
		return &rpc.RegisterMemberResponse{
			Error: errorResponse(rpc.ErrorStatus_INVALID_MEMBER, err.Error()),
		}, nil
	}
	if err != nil {
		logger.Error("unknown error", zap.String("id", req.Member.Id))
		return &rpc.RegisterMemberResponse{
			Error: errorResponse(rpc.ErrorStatus_UNKNOWN, err.Error()),
		}, nil
	}

	logger.Debug("member registered", zap.String("id", req.Member.Id))

	return &rpc.RegisterMemberResponse{}, nil
}

func (s *Server) UnregisterMember(ctx context.Context, req *rpc.UnregisterMemberRequest) (*rpc.UnregisterMemberResponse, error) {
	logger := s.logger.With(zap.String("rpc", "registry.UnregisterMember"))

	if s.registry.Unregister(req.Id) {
		logger.Debug("member unregistered", zap.String("id", req.Id))
	} else {
		logger.Warn("node already unregistered", zap.String("id", req.Id))
	}

	return &rpc.UnregisterMemberResponse{}, nil
}

func (s *Server) UpdateMemberMetadata(ctx context.Context, req *rpc.UpdateMemberMetadataRequest) (*rpc.UpdateMemberMetadataResponse, error) {
	logger := s.logger.With(zap.String("rpc", "registry.UpdateMemberMetadata"))

	err := s.registry.UpdateMemberMetadata(req.Id, req.Metadata)
	if err == ErrNotRegistered {
		logger.Warn("member not registered", zap.String("id", req.Id))
		return &rpc.UpdateMemberMetadataResponse{
			Error: errorResponse(rpc.ErrorStatus_NOT_REGISTERED, err.Error()),
		}, nil
	}
	if err == ErrInvalidUpdate {
		logger.Warn("invalid updatea", zap.String("id", req.Id))
		return &rpc.UpdateMemberMetadataResponse{
			Error: errorResponse(rpc.ErrorStatus_INVALID_MEMBER, err.Error()),
		}, nil
	}
	if err != nil {
		logger.Error("unknown error", zap.String("id", req.Id))
		return &rpc.UpdateMemberMetadataResponse{
			Error: errorResponse(rpc.ErrorStatus_UNKNOWN, err.Error()),
		}, nil
	}

	logger.Debug("member metadata updated", zap.String("id", req.Id))

	return &rpc.UpdateMemberMetadataResponse{}, nil
}

func (s *Server) Subscribe(req *rpc.SubscribeRequest, stream rpc.Registry_SubscribeServer) error {
	logger := s.logger.With(zap.String("rpc", "registry.Subscribe"))

	done := make(chan interface{})
	unsubscribe := s.registry.Subscribe(req.Versions, func(update *rpc.MemberUpdate) {
		logger.Debug(
			"send update",
			zap.String("id", update.Id),
			zap.String("update-type", update.UpdateType.String()),
		)
		if err := stream.Send(update); err != nil {
			logger.Debug("stream closed", zap.Error(err))

			close(done)
		}
	})
	defer unsubscribe()

	select {
	case <-stream.Context().Done():
		logger.Debug("stream context cancelled")
	case <-done:
	}

	return nil
}

func (s *Server) Heartbeat(ctx context.Context, req *rpc.HeartbeatRequest) (*rpc.HeartbeatResponse, error) {
	logger := s.logger.With(zap.String("rpc", "registry.Heartbeat"))

	s.registry.Heartbeat(req.ClientId)

	logger.Debug("heartbeat", zap.String("client-id", req.ClientId))

	return &rpc.HeartbeatResponse{}, nil
}

func (s *Server) Member(ctx context.Context, req *rpc.MemberRequest) (*rpc.MemberResponse, error) {
	logger := s.logger.With(zap.String("rpc", "registry.Member"))

	m, ok := s.registry.Member(req.Id)
	if !ok {
		logger.Debug("member not found", zap.String("id", req.Id))
		return &rpc.MemberResponse{
			Error: errorResponse(rpc.ErrorStatus_NOT_FOUND, "not found"),
		}, nil
	}

	logger.Debug("member found", zap.String("id", req.Id))

	return &rpc.MemberResponse{
		Member: m,
	}, nil
}

func (s *Server) Members(context.Context, *rpc.MembersRequest) (*rpc.MembersResponse, error) {
	logger := s.logger.With(zap.String("rpc", "registry.Members"))

	members := s.registry.Members()

	logger.Debug("members found", zap.Int("num", len(members)))

	return &rpc.MembersResponse{
		Members: members,
	}, nil
}

func errorResponse(status rpc.ErrorStatus, description string) *rpc.Error {
	return &rpc.Error{
		Status:      status,
		Description: description,
	}
}
