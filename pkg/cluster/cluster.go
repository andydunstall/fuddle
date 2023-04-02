package cluster

import (
	"sync"

	"github.com/fuddle-io/fuddle/pkg/registry"
	"go.uber.org/zap"
)

type Cluster struct {
	streams map[string]*stream

	// mu is a mutex protecting the fields above.
	mu sync.Mutex

	registry *registry.Registry
	logger   *zap.Logger
}

func NewCluster(registry *registry.Registry, logger *zap.Logger) *Cluster {
	return &Cluster{
		streams:  make(map[string]*stream),
		registry: registry,
		logger:   logger,
	}
}

func (c *Cluster) OnJoin(id string, addr string) {
	c.logger.Info(
		"cluster on join",
		zap.String("id", id),
		zap.String("addr", addr),
	)

	stream, err := connect(addr, c.registry)
	if err != nil {
		c.logger.Error("stream connect", zap.Error(err))
		return
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	c.streams[id] = stream
}

func (c *Cluster) OnLeave(id string) {
	c.logger.Info("cluster on leave", zap.String("id", id))

	c.mu.Lock()
	defer c.mu.Unlock()

	if s, ok := c.streams[id]; ok {
		s.Close()
		delete(c.streams, id)
	}
}
