package config

import (
	"time"

	"github.com/google/uuid"
	"go.uber.org/zap/zapcore"
)

type Config struct {
	NodeID   string
	RPC      *RPC
	Gossip   *Gossip
	Registry *Registry
}

func DefaultConfig() *Config {
	return &Config{
		NodeID: "fuddle-" + randomID(),
		RPC: &RPC{
			BindAddr: "0.0.0.0",
			BindPort: 8110,
			AdvAddr:  "",
			AdvPort:  8110,
		},
		Gossip: &Gossip{
			BindAddr: "0.0.0.0",
			BindPort: 8111,
			AdvAddr:  "",
			AdvPort:  8111,
		},
		Registry: &Registry{
			HeartbeatTimeout: time.Second * 20,
			ReconnectTimeout: time.Minute * 5,
			TombstoneTimeout: time.Minute * 30,
		},
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
