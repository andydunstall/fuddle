package node

import (
	"fmt"
	"time"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
	adminServer "github.com/fuddle-io/fuddle/pkg/admin/server"
	"github.com/fuddle-io/fuddle/pkg/cluster"
	"github.com/fuddle-io/fuddle/pkg/config"
	"github.com/fuddle-io/fuddle/pkg/gossip"
	"github.com/fuddle-io/fuddle/pkg/logger"
	"github.com/fuddle-io/fuddle/pkg/metrics"
	"github.com/fuddle-io/fuddle/pkg/registry/registry"
	registryServer "github.com/fuddle-io/fuddle/pkg/registry/server"
	rpcServer "github.com/fuddle-io/fuddle/pkg/server"
	"go.uber.org/zap"
)

// Node sets up and manages a Fuddle node.
type Node struct {
	Config *config.Config

	gossip      *gossip.Gossip
	registry    *registry.Registry
	rpcServer   *rpcServer.Server
	adminServer *adminServer.Server

	done chan interface{}

	logger *zap.Logger
}

// NewNode creates and starts a Fuddle node with the given config and options.
func NewNode(conf *config.Config, opts ...Option) (*Node, error) {
	options := defaultOptions()
	for _, o := range opts {
		o.apply(&options)
	}

	collector := metrics.NewPromCollector()

	logger, err := logger.NewLogger(
		logger.WithLevel(options.logLevel),
		logger.WithPath(options.logPath),
		logger.WithCollector(collector),
	)
	if err != nil {
		return nil, err
	}

	logger.Logger("fuddle").Info("starting fuddle", zap.Object("conf", conf))

	r := registry.NewRegistry(
		conf.NodeID,
		registry.WithLocalMember(&rpc.MemberState{
			Id:       conf.NodeID,
			Status:   "active",
			Service:  "fuddle",
			Started:  time.Now().UnixMilli(),
			Revision: "unknown",
		}),
		registry.WithHeartbeatTimeout(conf.Registry.HeartbeatTimeout.Milliseconds()),
		registry.WithReconnectTimeout(conf.Registry.ReconnectTimeout.Milliseconds()),
		registry.WithTombstoneTimeout(conf.Registry.TombstoneTimeout.Milliseconds()),
		registry.WithCollector(collector),
		registry.WithLogger(logger.Logger("registry")),
	)

	c := cluster.NewCluster(
		r,
		cluster.WithLogger(logger.Logger("cluster")),
		cluster.WithCollector(collector),
	)

	r.SubscribeLocal(func(update *rpc.Member2) {
		c.OnUpdate(update)
	})

	var adminServerOpts []adminServer.Option
	if options.adminListener != nil {
		adminServerOpts = append(adminServerOpts, adminServer.WithListener(options.adminListener))
	}
	adminServerOpts = append(adminServerOpts, adminServer.WithCollector(collector))
	adminServerOpts = append(adminServerOpts, adminServer.WithLogger(logger.Logger("admin")))
	adminServer, err := adminServer.NewServer(conf, adminServerOpts...)
	if err != nil {
		return nil, fmt.Errorf("fuddle: %w", err)
	}

	var gossipOpts []gossip.Option
	if options.gossipTCPListener != nil {
		gossipOpts = append(gossipOpts, gossip.WithTCPListener(
			options.gossipTCPListener,
		))
	}
	if options.gossipUDPListener != nil {
		gossipOpts = append(gossipOpts, gossip.WithUDPListener(
			options.gossipUDPListener,
		))
	}
	gossipOpts = append(gossipOpts, gossip.WithOnJoin(func(node gossip.Node) {
		if node.ID != conf.NodeID {
			c.OnJoin(node.ID, node.RPCAddr)
		}
	}))
	gossipOpts = append(gossipOpts, gossip.WithOnLeave(func(node gossip.Node) {
		if node.ID != conf.NodeID {
			c.OnLeave(node.ID)
		}
	}))
	gossipOpts = append(gossipOpts, gossip.WithLogger(logger.Logger("gossip")))

	g, err := gossip.NewGossip(conf, gossipOpts...)
	if err != nil {
		return nil, fmt.Errorf("fuddle: %w", err)
	}

	var rpcServerOpts []rpcServer.Option
	if options.rpcListener != nil {
		rpcServerOpts = append(rpcServerOpts, rpcServer.WithListener(options.rpcListener))
	}
	rpcServerOpts = append(rpcServerOpts, rpcServer.WithLogger(logger.Logger("server")))
	s := rpcServer.NewServer(conf, rpcServerOpts...)

	clientReadRegistryServer := registryServer.NewClientReadRegistryServer(
		r,
		registryServer.WithLogger(logger.Logger("registry")),
		registryServer.WithCollector(collector),
	)
	clientWriteRegistryServer := registryServer.NewClientWriteRegistryServer(
		r,
		registryServer.WithLogger(logger.Logger("registry")),
		registryServer.WithCollector(collector),
	)
	replicaReadRegistryServer := registryServer.NewReplicaReadRegistryServer(
		r,
		registryServer.WithLogger(logger.Logger("registry")),
		registryServer.WithCollector(collector),
	)
	rpc.RegisterClientReadRegistryServer(s.GRPCServer(), clientReadRegistryServer)
	rpc.RegisterClientWriteRegistryServer(s.GRPCServer(), clientWriteRegistryServer)
	rpc.RegisterReplicaReadRegistryServer(s.GRPCServer(), replicaReadRegistryServer)

	if err := s.Serve(); err != nil {
		return nil, fmt.Errorf("fuddle: %w", err)
	}

	n := &Node{
		Config:      conf,
		registry:    r,
		gossip:      g,
		rpcServer:   s,
		adminServer: adminServer,
		logger:      logger.Logger("fuddle"),
		done:        make(chan interface{}),
	}

	go n.failureDetector()

	return n, nil
}

func (n *Node) Registry() *registry.Registry {
	return n.registry
}

func (n *Node) Nodes() map[string]interface{} {
	return n.gossip.Nodes()
}

func (n *Node) Shutdown() {
	n.logger.Info("shutting down fuddle")

	close(n.done)

	n.rpcServer.Shutdown()
	n.gossip.Shutdown()
	n.adminServer.Shutdown()
}

func (n *Node) failureDetector() {
	ticker := time.NewTicker(time.Millisecond * 100)
	defer ticker.Stop()

	for {
		select {
		case <-n.done:
			return
		case <-ticker.C:
			n.registry.CheckMembersLiveness()
		}
	}
}
