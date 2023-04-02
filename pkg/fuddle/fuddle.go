package fuddle

import (
	"fmt"

	"github.com/fuddle-io/fuddle/pkg/cluster"
	"github.com/fuddle-io/fuddle/pkg/config"
	"github.com/fuddle-io/fuddle/pkg/gossip"
	"go.uber.org/zap"
)

// Fuddle implements a single Fuddle node.
type Fuddle struct {
	gossip *gossip.Gossip
	logger *zap.Logger
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

	c := cluster.NewCluster(logger.With(zap.String("stream", "cluster")))

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
		c.OnJoin(id, addr)
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

	return &Fuddle{
		gossip: g,
		logger: logger,
	}, nil
}

func (f *Fuddle) Nodes() map[string]interface{} {
	return f.gossip.Nodes()
}

func (f *Fuddle) Shutdown() {
	f.logger.Info("shutting down fuddle")

	f.gossip.Shutdown()
}
