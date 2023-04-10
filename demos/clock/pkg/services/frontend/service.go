package frontend

import (
	"context"
	"fmt"
	"net"
	"time"

	fuddle "github.com/fuddle-io/fuddle-go"
	"github.com/google/uuid"
)

type Service struct {
	ID   string
	Addr string

	fuddleClient *fuddle.Fuddle
	server       *server
}

func NewService(ln *net.TCPListener, fuddleAddrs []string) (*Service, error) {
	member := fuddle.Member{
		ID:       "frontend-" + uuid.New().String()[:8],
		Service:  "frontend",
		Created:  time.Now().UnixMilli(),
		Revision: "v0.1.0",
		Metadata: map[string]string{
			"rpc-addr": ln.Addr().String(),
		},
	}

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	fuddleClient, err := fuddle.Register(
		ctx,
		fuddleAddrs,
		member,
	)
	if err != nil {
		return nil, fmt.Errorf("frontend service: %w", err)
	}

	server := newServer(ln)
	return &Service{
		ID:           member.ID,
		Addr:         ln.Addr().String(),
		fuddleClient: fuddleClient,
		server:       server,
	}, nil
}

func (s *Service) Shutdown() {
	s.fuddleClient.Close()
	s.server.Shutdown()
}
