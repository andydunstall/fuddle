package config

import (
	"github.com/google/uuid"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	NodeID   string
	RPC      *RPC
	Gossip   *Gossip
	Admin    *Admin
	Registry *Registry
}

func DefaultConfig() *Config {
	return &Config{
		NodeID:   "fuddle-" + randomID(),
		RPC:      DefaultRPCConfig(),
		Gossip:   DefaultGossipConfig(),
		Admin:    DefaultAdminConfig(),
		Registry: DefaultRegistryConfig(),
	}
}

func (c *Config) MarshalLogObject(e zapcore.ObjectEncoder) error {
	e.AddString("node-id", c.NodeID)
	if err := e.AddObject("rpc", c.RPC); err != nil {
		return err
	}
	if err := e.AddObject("gossip", c.Gossip); err != nil {
		return err
	}
	if err := e.AddObject("admin", c.Admin); err != nil {
		return err
	}
	if err := e.AddObject("registry", c.Registry); err != nil {
		return err
	}
	return nil
}

func randomID() string {
	return uuid.New().String()[:8]
}

type stringArray []string

func (ss stringArray) MarshalLogArray(arr zapcore.ArrayEncoder) error {
	for i := range ss {
		arr.AppendString(ss[i])
	}
	return nil
}
