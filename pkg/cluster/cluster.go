package cluster

import (
	"sync"

	"github.com/fuddle-io/fuddle/pkg/registry"
	"go.uber.org/zap"
)

type Cluster struct {
	clients map[string]*registry.Client

	// mu is a mutex protecting the fields above.
	mu sync.Mutex

	registry *registry.Registry

	logger *zap.Logger
}

func NewCluster(reg *registry.Registry, logger *zap.Logger) *Cluster {
	return &Cluster{
		clients:  make(map[string]*registry.Client),
		registry: reg,
		logger:   logger,
	}
}

func (c *Cluster) OnJoin(id string, addr string) {
	c.logger.Info(
		"cluster on join",
		zap.String("id", id),
		zap.String("addr", addr),
	)

	client, err := registry.Connect(
		addr, c.registry, registry.WithClientLogger(c.logger),
	)
	if err != nil {
		c.logger.Error("client connect", zap.Error(err))
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.clients[id] = client
}

func (c *Cluster) OnLeave(id string) {
	c.logger.Info("cluster on leave", zap.String("id", id))

	c.mu.Lock()
	defer c.mu.Unlock()

	if client, ok := c.clients[id]; ok {
		client.Close()
		delete(c.clients, id)
	}
}
