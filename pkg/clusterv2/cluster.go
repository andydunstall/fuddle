package cluster

import (
	"context"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"github.com/fuddle-io/fuddle/pkg/registryv2/client"
	"github.com/fuddle-io/fuddle/pkg/registryv2/registry"
)

type Cluster struct {
	registry *registry.Registry
}

func NewCluster(registry *registry.Registry) *Cluster {
	c := &Cluster{
		registry: registry,
	}
	return c
}

func (c *Cluster) OnJoin(id string, addr string) {
	c.registry.OnNodeJoin(id)

	conn := client.ConnectReplica(addr)
	c.registry.SubscribeLocal(func(m *rpc.Member2) {
		// nolint
		conn.Update(context.Background(), m)
	})
}

func (c *Cluster) OnLeave(id string) {
	c.registry.OnNodeLeave(id)
}

func (c *Cluster) ReplicaRepair() {
	// Select a random node and call client.Sync().
}
