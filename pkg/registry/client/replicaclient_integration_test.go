//go:build all || integration

package client_test

import (
	"context"
	"fmt"
	"net"
	"testing"
	"time"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"github.com/fuddle-io/fuddle/pkg/registry/client"
	"github.com/fuddle-io/fuddle/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"google.golang.org/grpc"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
	"google.golang.org/protobuf/proto"
)

type fakeReplicaServer struct {
	Ch            chan *rpc.Member2
	numUnavilable int

	rpc.UnimplementedReplicaRegistry2Server
}

func newFakeReplicaServer(numUnavilable int) *fakeReplicaServer {
	return &fakeReplicaServer{
		Ch:            make(chan *rpc.Member2, 64),
		numUnavilable: numUnavilable,
	}
}

func (s *fakeReplicaServer) Update(ctx context.Context, req *rpc.UpdateRequest) (*rpc.UpdateResponse, error) {
	if s.numUnavilable > 0 {
		s.numUnavilable--
		return nil, status.Error(
			codes.Unavailable, "Service is currently unavailable",
		)
	}

	s.Ch <- req.Member
	return &rpc.UpdateResponse{}, nil
}

func (s *fakeReplicaServer) Serve() (*grpc.Server, string, error) {
	ln, err := net.Listen("tcp", "127.0.0.1:0")
	if err != nil {
		return nil, "", fmt.Errorf("fake replica server: listen: %w", err)
	}

	grpcServer := grpc.NewServer()
	rpc.RegisterReplicaRegistry2Server(grpcServer, s)

	go func() {
		if err := grpcServer.Serve(ln); err != nil {
			panic(err)
		}
	}()

	return grpcServer, ln.Addr().String(), nil
}

func TestClient_ForwardUpdate(t *testing.T) {
	server := newFakeReplicaServer(0)
	grpcServer, addr, err := server.Serve()
	require.NoError(t, err)
	defer grpcServer.Stop()

	client, err := client.ReplicaConnect(
		addr, "local", "target", client.NewReplicaClientMetrics(),
	)
	require.NoError(t, err)
	defer client.Close()

	member := &rpc.Member2{
		State:    testutils.RandomMemberState("new-member", ""),
		Liveness: rpc.Liveness_UP,
		Version: &rpc.Version2{
			OwnerId: "foo-123",
			Timestamp: &rpc.MonotonicTimestamp{
				Timestamp: time.Now().UnixMilli() + 10000,
			},
		},
	}
	client.Update(member)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	update, ok := waitWithContext(ctx, server.Ch)
	assert.True(t, ok, "timeout")
	assert.True(t, proto.Equal(member, update))
}

// Tests the replica client retries updates when the server is unavailable.
func TestClient_ForwardUpdateRetryUnavailable(t *testing.T) {
	// Configure the server to return UNAVAILABLE the first 4 attempts.
	server := newFakeReplicaServer(4)
	grpcServer, addr, err := server.Serve()
	require.NoError(t, err)
	defer grpcServer.Stop()

	client, err := client.ReplicaConnect(
		addr, "local", "target", client.NewReplicaClientMetrics(),
	)
	require.NoError(t, err)
	defer client.Close()

	member := &rpc.Member2{
		State:    testutils.RandomMemberState("new-member", ""),
		Liveness: rpc.Liveness_UP,
		Version: &rpc.Version2{
			OwnerId: "foo-123",
			Timestamp: &rpc.MonotonicTimestamp{
				Timestamp: time.Now().UnixMilli() + 10000,
			},
		},
	}
	client.Update(member)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()
	update, ok := waitWithContext(ctx, server.Ch)
	assert.True(t, ok, "timeout")
	assert.True(t, proto.Equal(member, update))
}

func waitWithContext(ctx context.Context, ch chan *rpc.Member2) (*rpc.Member2, bool) {
	select {
	case m := <-ch:
		return m, true
	case <-ctx.Done():
		return nil, false
	}
}
