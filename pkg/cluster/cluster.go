package cluster

import (
	"sync"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
	registryClient "github.com/fuddle-io/fuddle/pkg/registry/client"
	"github.com/fuddle-io/fuddle/pkg/registry/registry"
	"go.uber.org/zap"
)

type Cluster struct {
	nodes   map[string]interface{}
	clients map[string]*registryClient.ReplicaClient

	// mu is a mutex protecting the fields above.
	mu sync.Mutex

	registry *registry.Registry

	logger  *zap.Logger
	metrics *Metrics
}

func NewCluster(reg *registry.Registry, opts ...Option) *Cluster {
	options := defaultOptions()
	for _, o := range opts {
		o.apply(options)
	}

	metrics := NewMetrics()
	if options.collector != nil {
		metrics.Register(options.collector)
	}

	return &Cluster{
		nodes:    make(map[string]interface{}),
		clients:  make(map[string]*registryClient.ReplicaClient),
		registry: reg,
		logger:   options.logger,
		metrics:  metrics,
	}
}

func (c *Cluster) OnJoin(id string, addr string) {
	c.logger.Info(
		"cluster on join",
		zap.String("id", id),
		zap.String("addr", addr),
	)

	client, err := registryClient.ReplicaConnect(
		addr, c.registry, registryClient.WithLogger(c.logger),
	)
	if err != nil {
		c.logger.Error("client connect", zap.Error(err))
		return
	}

	c.mu.Lock()
	c.nodes[id] = struct{}{}
	c.clients[id] = client

	nodesCount := len(c.nodes)
	c.mu.Unlock()

	// Add 1 to include this node.
	c.metrics.NodesCount.Set(float64(nodesCount+1), make(map[string]string))

	c.registry.OnNodeJoin(id)

	// To bootstrap the node send the members we own.
	for _, m := range c.registry.OwnedMembers() {
		client.Update(m)
	}
}

func (c *Cluster) OnLeave(id string) {
	c.logger.Info("cluster on leave", zap.String("id", id))

	c.mu.Lock()
	delete(c.nodes, id)
	if client, ok := c.clients[id]; ok {
		client.Close()
		delete(c.clients, id)
	}

	nodesCount := len(c.nodes)
	c.mu.Unlock()

	// Add 1 to include this node.
	c.metrics.NodesCount.Set(float64(nodesCount+1), make(map[string]string))

	c.registry.OnNodeLeave(id)
}

func (c *Cluster) OnUpdate(m *rpc.Member2) {
	c.mu.Lock()
	defer c.mu.Unlock()

	for _, client := range c.clients {
		client.Update(m)
	}
}
