package fuddle

import (
	"fmt"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"github.com/fuddle-io/fuddle/pkg/cluster"
	"github.com/fuddle-io/fuddle/pkg/config"
	"github.com/fuddle-io/fuddle/pkg/gossip"
	"github.com/fuddle-io/fuddle/pkg/registry"
	"github.com/fuddle-io/fuddle/pkg/server"
	"go.uber.org/zap"
)

// Fuddle implements a single Fuddle node.
type Fuddle struct {
	Config *config.Config

	gossip   *gossip.Gossip
	registry *registry.Registry
	server   *server.Server
	logger   *zap.Logger
}

func NewFuddle(conf *config.Config, opts ...Option) (*Fuddle, error) {
	options := defaultOptions()
	for _, o := range opts {
		o.apply(&options)
	}

	logger := options.logger.With(
		zap.String("stream", "fuddle"),
		zap.String("node-id", conf.NodeID),
	)

	logger.Info("starting fuddle", zap.Object("conf", conf))

	r := registry.NewRegistry(
		conf.NodeID,
		registry.WithRegistryLocalMember(&rpc.Member{
			Id: conf.NodeID,
		}),
		registry.WithRegistryLogger(
			logger.With(zap.String("stream", "registry")),
		),
	)

	c := cluster.NewCluster(r, logger.With(zap.String("stream", "cluster")))

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
	gossipOpts = append(gossipOpts, gossip.WithLogger(
		options.logger.With(zap.String("stream", "gossip")),
	))

	g, err := gossip.NewGossip(conf, gossipOpts...)
	if err != nil {
		return nil, fmt.Errorf("fuddle: %w", err)
	}

	var serverOpts []server.Option
	if options.rpcListener != nil {
		serverOpts = append(serverOpts, server.WithListener(options.rpcListener))
	}
	serverOpts = append(serverOpts, server.WithLogger(
		logger.With(zap.String("stream", "server")),
	))
	s := server.NewServer(conf, serverOpts...)

	registryServer := registry.NewServer(r, registry.WithServerLogger(
		logger.With(zap.String("stream", "registry")),
	))
	rpc.RegisterRegistryServer(s.GRPCServer(), registryServer)

	if err := s.Serve(); err != nil {
		return nil, fmt.Errorf("fuddle: %w", err)
	}

	return &Fuddle{
		Config:   conf,
		registry: r,
		gossip:   g,
		server:   s,
		logger:   logger,
	}, nil
}

func (f *Fuddle) Registry() *registry.Registry {
	return f.registry
}

func (f *Fuddle) Nodes() map[string]interface{} {
	return f.gossip.Nodes()
}

func (f *Fuddle) Shutdown() {
	f.logger.Info("shutting down fuddle")

	f.server.Shutdown()
	f.gossip.Shutdown()
}
