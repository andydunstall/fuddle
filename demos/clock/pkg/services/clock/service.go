package clock

import (
	"context"
	"fmt"
	"net"
	"time"

	fuddle "github.com/fuddle-io/fuddle-go"
	"github.com/google/uuid"
	grpc "google.golang.org/grpc"
)

type Service struct {
	ID   string
	Addr string

	fuddleClient *fuddle.Fuddle

	grpcServer *grpc.Server
}

func NewService(ln *net.TCPListener, fuddleAddrs []string) (*Service, error) {
	member := fuddle.Member{
		ID:       "clock-" + uuid.New().String()[:8],
		Service:  "clock",
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
		return nil, fmt.Errorf("clock service: %w", err)
	}

	grpcServer := grpc.NewServer()
	RegisterClockServer(grpcServer, &server{})

	go func() {
		if err := grpcServer.Serve(ln); err != nil {
			fmt.Println("grpc server", err)
		}
	}()

	return &Service{
		ID:           member.ID,
		Addr:         ln.Addr().String(),
		fuddleClient: fuddleClient,
		grpcServer:   grpcServer,
	}, nil
}

func (s *Service) Shutdown() {
	s.grpcServer.GracefulStop()
	s.fuddleClient.Close()
}
