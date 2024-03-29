package clock

import (
	"context"
	"fmt"
	"net"
	"time"

	fuddle "github.com/fuddle-io/fuddle-go"
	"go.uber.org/zap"
	grpc "google.golang.org/grpc"
)

type Service struct {
	ID   string
	Addr string

	fuddleClient *fuddle.Fuddle

	grpcServer *grpc.Server
}

func NewService(id string, ln *net.TCPListener, fuddleAddrs []string, logger *zap.Logger) (*Service, error) {
	member := fuddle.Member{
		ID:       id,
		Service:  "clock",
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
		return nil, fmt.Errorf("clock service: %w", err)
	}

	grpcServer := grpc.NewServer()
	RegisterClockServer(grpcServer, newServer(logger))

	go func() {
		if err := grpcServer.Serve(ln); err != nil {
			logger.Error("grpc server error", zap.Error(err))
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
