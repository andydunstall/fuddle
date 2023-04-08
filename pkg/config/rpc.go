package config

import (
	"fmt"

	"go.uber.org/zap/zapcore"
)

type RPC struct {
	// Address to bind to and listen on. Used for both UDP and TCP gossip.
	BindAddr string
	BindPort int

	// Address to advertise to other cluster members.
	AdvAddr string
	AdvPort int
}

func (c *RPC) MarshalLogObject(e zapcore.ObjectEncoder) error {
	e.AddString("bind-addr", c.BindAddr)
	e.AddInt("bind-port", c.BindPort)
	e.AddString("adv-addr", c.AdvAddr)
	e.AddInt("adv-port", c.AdvPort)
	return nil
}

func (c *RPC) JoinBindAddr() string {
	return fmt.Sprintf("%s:%d", c.BindAddr, c.BindPort)
}

func (c *RPC) JoinAdvAddr() string {
	return fmt.Sprintf("%s:%d", c.AdvAddr, c.AdvPort)
}

func DefaultRPCConfig() *RPC {
	return &RPC{
		BindAddr: "0.0.0.0",
		BindPort: 8110,
		AdvAddr:  "",
		AdvPort:  8110,
	}
}
