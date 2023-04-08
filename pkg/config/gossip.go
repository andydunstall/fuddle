package config

import (
	"go.uber.org/zap/zapcore"
)

type Gossip struct {
	// Address to bind to and listen on. Used for both UDP and TCP gossip.
	BindAddr string
	BindPort int

	// Address to advertise to other cluster members.
	AdvAddr string
	AdvPort int

	// Seeds contains a list of gossip addresses of nodes in the target cluster
	// to join.
	Seeds []string
}

func (c *Gossip) MarshalLogObject(e zapcore.ObjectEncoder) error {
	e.AddString("bind-addr", c.BindAddr)
	e.AddInt("bind-port", c.BindPort)
	e.AddString("adv-addr", c.AdvAddr)
	e.AddInt("adv-port", c.AdvPort)
	if err := e.AddArray("seeds", stringArray(c.Seeds)); err != nil {
		return err
	}
	return nil
}

func DefaultGossipConfig() *Gossip {
	return &Gossip{
		BindAddr: "0.0.0.0",
		BindPort: 8111,
		AdvAddr:  "",
		AdvPort:  8111,
		Seeds:    nil,
	}
}
