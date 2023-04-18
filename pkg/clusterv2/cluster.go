package cluster

import (
	"context"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"github.com/fuddle-io/fuddle/pkg/registryv2/client"
	"github.com/fuddle-io/fuddle/pkg/registryv2/registry"
)

// nolint
type ClusterV2 struct {
	registry *registry.Registry
}

func NewClusterV2(registry *registry.Registry) *ClusterV2 {
	c := &ClusterV2{
		registry: registry,
	}
	go c.readRepair()
	return c
}

func (c *ClusterV2) OnJoin(id string, addr string) {
	c.registry.OnNodeJoin(id)

	conn := client.ConnectReplica(addr)
	c.registry.SubscribeLocal(func(m *rpc.Member2) {
		// nolint
		conn.Update(context.Background(), m)
	})
}

func (c *ClusterV2) OnLeave(id string) {
	c.registry.OnNodeLeave(id)
}

func (c *ClusterV2) readRepair() {
	// Start a ticker.
	// Whenever the ticker goes off, select a random node and call
	// client.Sync().
}
