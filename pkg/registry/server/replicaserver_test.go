package server_test

import (
	"context"
	"testing"
	"time"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"github.com/fuddle-io/fuddle/pkg/registry/registry"
	"github.com/fuddle-io/fuddle/pkg/registry/server"
	"github.com/fuddle-io/fuddle/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestReplicaServer_Update(t *testing.T) {
	reg := registry.NewRegistry("local")
	server := server.NewReplicaServer(reg)

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
	_, err := server.Update(context.Background(), &rpc.UpdateRequest{
		Member:       member,
		SourceNodeId: "local",
	})
	assert.NoError(t, err)

	// Check the registry was updated.
	m, ok := reg.MemberState("new-member")
	assert.True(t, ok)
	assert.True(t, proto.Equal(member.State, m))
}

func TestReplicaServer_UpdateMetrics(t *testing.T) {
	reg := registry.NewRegistry("local")
	server := server.NewReplicaServer(reg)

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
	_, err := server.Update(context.Background(), &rpc.UpdateRequest{
		Member:       member,
		SourceNodeId: "local",
	})
	assert.NoError(t, err)

	assert.Equal(t, 1.0, server.Metrics().ReplicaUpdatesInbound.Value(map[string]string{
		"source": "local",
	}))

	_, err = server.Update(context.Background(), &rpc.UpdateRequest{
		Member:       member,
		SourceNodeId: "local",
	})
	assert.NoError(t, err)

	assert.Equal(t, 2.0, server.Metrics().ReplicaUpdatesInbound.Value(map[string]string{
		"source": "local",
	}))
}
