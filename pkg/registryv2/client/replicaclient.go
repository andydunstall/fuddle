package client

import (
	"context"
	"fmt"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"github.com/fuddle-io/fuddle/pkg/registryv2/registry"
)

const (
	// maxDigestSize is the maximum number of member versions to include in the
	// digest.
	maxDigestSize = 10000
)

type ReplicaClient struct {
	registry *registry.Registry

	client rpc.ReplicaRegistry2Client
}

func ConnectReplica(addr string) *ReplicaClient {
	return &ReplicaClient{}
}

func (c *ReplicaClient) Update(ctx context.Context, member *rpc.Member2) error {
	// TODO(AD) don't block
	return nil
}

func (c *ReplicaClient) Sync(ctx context.Context) error {
	resp, err := c.client.Sync(ctx, &rpc.ReplicaSyncRequest{
		Digest: c.registry.MembersDigest(maxDigestSize),
	})
	if err != nil {
		return fmt.Errorf("replica client: sync: %w", err)
	}
	for _, m := range resp.Members {
		c.registry.RemoteUpsertMember(m)
	}
	return nil
}
