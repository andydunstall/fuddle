package server

import (
	"context"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"github.com/fuddle-io/fuddle/pkg/registry"
	"go.uber.org/zap"
)

// ClientReadRegistryServer serves updates to the registry to the external
// clients.
type ClientReadRegistryServer struct {
	registry *registry.Registry

	logger *zap.Logger

	rpc.UnimplementedClientReadRegistryServer
}

func NewClientReadRegistryServer(reg *registry.Registry, opts ...Option) *ClientReadRegistryServer {
	options := defaultOptions()
	for _, o := range opts {
		o.apply(options)
	}

	return &ClientReadRegistryServer{
		registry: reg,
		logger:   options.logger,
	}
}

// Updates streams updates to the local registry. This includes sending any
// updates the client missed given their known members in the subscribe request.
func (s *ClientReadRegistryServer) Updates(req *rpc.SubscribeRequest, stream rpc.ClientReadRegistry_UpdatesServer) error {
	logger := s.logger.With(zap.String("rpc", "ClientReadRegistryServer.Updates"))
	logger.Debug("updates stream")

	unsubscribe := s.registry.Subscribe(req, func(update *rpc.RemoteMemberUpdate) {
		logger.Debug(
			"send update",
			zap.String("id", update.Member.Id),
			zap.String("type", update.UpdateType.String()),
		)

		// Ignore return error, if the client closes the stream the context
		// will be cancelled.
		// nolint
		stream.Send(update)
	})
	defer unsubscribe()

	<-stream.Context().Done()
	logger.Debug("subscribe stream closed")

	return nil
}

// Member looks up the requested member.
func (s *ClientReadRegistryServer) Member(ctx context.Context, req *rpc.MemberRequest) (*rpc.MemberResponse, error) {
	logger := s.logger.With(zap.String("rpc", "ClientReadRegistryServer.Member"))

	m, ok := s.registry.Member(req.Id)
	if !ok {
		logger.Debug("member request; not found", zap.String("id", req.Id))
		return &rpc.MemberResponse{}, nil
	}

	logger.Debug("member request; found", zap.String("id", req.Id))

	return &rpc.MemberResponse{
		Member: m,
	}, nil
}

// Members lists the members in the registry.
func (s *ClientReadRegistryServer) Members(context.Context, *rpc.MembersRequest) (*rpc.MembersResponse, error) {
	logger := s.logger.With(zap.String("rpc", "ClientReadRegistryServer.Members"))

	members := s.registry.Members()
	logger.Debug("members request", zap.Int("num-members", len(members)))

	return &rpc.MembersResponse{
		Members: members,
	}, nil
}
