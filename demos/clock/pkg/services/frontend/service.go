package frontend

import (
	"context"
	"fmt"
	"net"
	"time"

	fuddle "github.com/fuddle-io/fuddle-go"
	"github.com/fuddle-io/fuddle/demos/clock/pkg/services/clock"
	"go.uber.org/zap"
)

type Service struct {
	ID   string
	Addr string

	fuddleClient *fuddle.Fuddle
	clockClient  *clock.Client
	server       *server
}

func NewService(id string, ln *net.TCPListener, fuddleAddrs []string, logger *zap.Logger) (*Service, error) {
	member := fuddle.Member{
		ID:       id,
		Service:  "frontend",
		Started:  time.Now().UnixMilli(),
		Revision: "v0.1.0",
		Metadata: map[string]string{
			"rpc-addr": ln.Addr().String(),
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	fuddleClient, err := fuddle.Connect(
		ctx,
		member,
		fuddleAddrs,
		fuddle.WithLogger(logger.With(zap.String("stream", "fuddle"))),
	)
	if err != nil {
		return nil, fmt.Errorf("frontend service: %w", err)
	}

	clockClient, err := clock.NewClient(fuddleClient)
	if err != nil {
		return nil, fmt.Errorf("frontend service: %w", err)
	}

	server := newServer(
		ln,
		clockClient,
		logger.With(zap.String("stream", "server")),
	)
	return &Service{
		ID:           id,
		Addr:         ln.Addr().String(),
		fuddleClient: fuddleClient,
		clockClient:  clockClient,
		server:       server,
	}, nil
}

func (s *Service) Shutdown() {
	s.clockClient.Close()
	s.fuddleClient.Close()
	s.server.Shutdown()
}
