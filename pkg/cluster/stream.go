package cluster

import (
	"context"
	"time"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"github.com/fuddle-io/fuddle/pkg/registry"
	"google.golang.org/grpc"
	"google.golang.org/grpc/credentials/insecure"
)

// TODO(AD) not yet handling reconnect etc.
type stream struct {
	registry *registry.Registry
	conn     *grpc.ClientConn
}

func connect(addr string, registry *registry.Registry) (*stream, error) {
	ctx, cancel := context.WithTimeout(context.Background(), time.Second)
	defer cancel()

	conn, err := grpc.DialContext(
		ctx,
		addr,
		grpc.WithTransportCredentials(insecure.NewCredentials()),
	)
	if err != nil {
		return nil, err
	}

	client := rpc.NewRegistryClient(conn)
	grpcStream, err := client.Subscribe(context.Background(), &rpc.SubscribeRequest{
		OwnerOnly: true,
	})
	if err != nil {
		return nil, err
	}

	s := &stream{
		registry: registry,
		conn:     conn,
	}

	go s.recvLoop(grpcStream)

	return s, nil
}

func (s *stream) Close() {
	s.conn.Close()
}

func (s *stream) recvLoop(grpcStream rpc.Registry_SubscribeClient) {
	for {
		update, err := grpcStream.Recv()
		if err != nil {
			return
		}

		switch update.UpdateType {
		case rpc.MemberUpdateType_REGISTER:
			s.registry.RegisterRemote(update.Id, "")
			// s.registry.RegisterRemote(update.Member.Id, update.Member.Owner)
		}
	}
}
