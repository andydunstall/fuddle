package gossip

import (
	"fmt"
	"time"

	"github.com/fuddle-io/fuddle/pkg/config"
	"github.com/hashicorp/memberlist"
	"go.uber.org/zap"
)

type Node struct {
	ID           string
	RegistryAddr string
}

type Gossip struct {
	memberlist *memberlist.Memberlist
	logger     *zap.Logger
}

func NewGossip(conf *config.Config, opts ...Option) (*Gossip, error) {
	options := defaultOptions()
	for _, o := range opts {
		o.apply(&options)
	}

	memberlistConf := memberlist.DefaultLANConfig()
	memberlistConf.Name = conf.NodeID
	memberlistConf.BindAddr = conf.Gossip.BindAddr
	memberlistConf.BindPort = conf.Gossip.BindPort
	memberlistConf.AdvertiseAddr = conf.Gossip.AdvAddr
	memberlistConf.AdvertisePort = conf.Gossip.AdvPort
	memberlistConf.SuspicionMult = 3
	transport, err := newTransport(conf.Gossip, options)
	if err != nil {
		return nil, fmt.Errorf("gossip: transport: %w", err)
	}
	memberlistConf.Transport = transport
	memberlistConf.Delegate = newDelegate([]byte(
		fmt.Sprintf("%s:%d", conf.Registry.AdvAddr, conf.Registry.AdvPort),
	))
	memberlistConf.Events = newEventDelegate(
		options.onJoin,
		options.onLeave,
	)
	memberlistConf.LogOutput = newLoggerWriter(options.logger)

	memberlist, err := memberlist.Create(memberlistConf)
	if err != nil {
		return nil, fmt.Errorf("gossip: memberlist: %w", err)
	}

	if _, err = memberlist.Join(conf.Gossip.Seeds); err != nil {
		options.logger.Error(
			"failed to join cluster",
			zap.Strings("seeds", conf.Gossip.Seeds),
			zap.Error(err),
		)
		return nil, fmt.Errorf("gossip: memberlist: %w", err)
	}

	options.logger.Info(
		"joined cluster",
		zap.Strings("seeds", conf.Gossip.Seeds),
	)

	return &Gossip{
		memberlist: memberlist,
		logger:     options.logger,
	}, nil
}

func (g *Gossip) Nodes() []Node {
	var nodes []Node
	for _, m := range g.memberlist.Members() {
		nodes = append(nodes, Node{
			ID:           m.Name,
			RegistryAddr: string(m.Meta),
		})
	}
	return nodes
}

func (g *Gossip) Shutdown() {
	g.logger.Info("gossip shutdown")
	if err := g.memberlist.Leave(time.Second); err != nil {
		g.logger.Error("failed to leave gossip", zap.Error(err))
	}
	g.memberlist.Shutdown()
}
