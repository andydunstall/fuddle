package server

import (
	"context"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"github.com/fuddle-io/fuddle/pkg/metrics"
	"github.com/fuddle-io/fuddle/pkg/registry/registry"
	"go.uber.org/zap"
)

// ClientReadServer serves updates to the registry to the external
// clients.
type ClientReadServer struct {
	registry *registry.Registry

	outboundUpdates *metrics.Counter
	logger          *zap.Logger

	rpc.UnimplementedClientReadRegistryServer
}

func NewClientReadServer(reg *registry.Registry, opts ...Option) *ClientReadServer {
	options := defaultOptions()
	for _, o := range opts {
		o.apply(options)
	}

	outboundUpdates := metrics.NewCounter(
		"registry",
		"updates.client.outbound",
		[]string{},
		"Number of outbound updates sent to a client",
	)
	if options.collector != nil {
		options.collector.AddCounter(outboundUpdates)
	}

	return &ClientReadServer{
		registry:        reg,
		outboundUpdates: outboundUpdates,
		logger:          options.logger,
	}
}

// Updates streams updates to the local registry. This includes sending any
// updates the client missed given their known members in the subscribe request.
func (s *ClientReadServer) Updates(req *rpc.SubscribeRequest, stream rpc.ClientReadRegistry_UpdatesServer) error {
	logger := s.logger.With(zap.String("rpc", "ClientReadServer.Updates"))
	logger.Debug("updates stream")

	unsubscribe := s.registry.Subscribe(req, func(update *rpc.Member2) {
		logger.Debug(
			"send update",
			zap.String("id", update.State.Id),
		)

		s.outboundUpdates.Inc(map[string]string{})

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
func (s *ClientReadServer) Member(ctx context.Context, req *rpc.MemberRequest) (*rpc.MemberResponse, error) {
	logger := s.logger.With(zap.String("rpc", "ClientReadServer.Member"))

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
func (s *ClientReadServer) Members(context.Context, *rpc.MembersRequest) (*rpc.MembersResponse, error) {
	logger := s.logger.With(zap.String("rpc", "ClientReadServer.Members"))

	members := s.registry.Members()
	logger.Debug("members request", zap.Int("num-members", len(members)))

	return &rpc.MembersResponse{
		Members: members,
	}, nil
}
