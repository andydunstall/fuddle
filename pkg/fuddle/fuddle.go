package fuddle

import (
	"fmt"

	"github.com/fuddle-io/fuddle/pkg/cluster"
	"github.com/fuddle-io/fuddle/pkg/config"
	"github.com/fuddle-io/fuddle/pkg/gossip"
	"github.com/fuddle-io/fuddle/pkg/registry"
	"github.com/fuddle-io/fuddle/pkg/server"
	"go.uber.org/zap"
)

// Fuddle implements a single Fuddle node.
type Fuddle struct {
	server   *server.Server
	registry *registry.Registry
	gossip   *gossip.Gossip
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
		conf.NodeID, logger.With(zap.String("stream", "registry")),
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
	gossipOpts = append(gossipOpts, gossip.WithOnJoin(func(id string, addr string) {
		if id != conf.NodeID {
			c.OnJoin(id, addr)
		}
	}))
	gossipOpts = append(gossipOpts, gossip.WithOnLeave(func(id string) {
		c.OnLeave(id)
	}))
	gossipOpts = append(gossipOpts, gossip.WithLogger(
		logger.With(zap.String("stream", "gossip")),
	))

	g, err := gossip.NewGossip(conf, gossipOpts...)
	if err != nil {
		return nil, fmt.Errorf("fuddle: %w", err)
	}

	var serverOpts []server.Option
	if options.registryListener != nil {
		serverOpts = append(serverOpts, server.WithListener(options.registryListener))
	}
	serverOpts = append(serverOpts, server.WithLogger(
		logger.With(zap.String("stream", "server")),
	))
	s, err := server.NewServer(conf, r, serverOpts...)
	if err != nil {
		return nil, fmt.Errorf("fuddle: %w", err)
	}

	return &Fuddle{
		server:   s,
		registry: r,
		gossip:   g,
		logger:   logger,
	}, nil
}

func (f *Fuddle) Nodes() map[string]interface{} {
	return f.gossip.Nodes()
}

func (f *Fuddle) Registry() *registry.Registry {
	return f.registry
}

func (f *Fuddle) Shutdown() {
	f.logger.Info("shutting down fuddle")

	f.server.Shutdown()
	f.gossip.Shutdown()
}
