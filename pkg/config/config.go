package config

import (
	"github.com/google/uuid"
)

type Config struct {
	NodeID   string
	Registry Registry
	Gossip   Gossip
}

func DefaultConfig() *Config {
	return &Config{
		NodeID: "fuddle-" + randomID(),
		Registry: Registry{
			BindAddr: "0.0.0.0",
			BindPort: 8110,
			AdvAddr:  "",
			AdvPort:  8110,
		},
		Gossip: Gossip{
			BindAddr: "0.0.0.0",
			BindPort: 8111,
			AdvAddr:  "",
			AdvPort:  8111,
		},
	}
}

func randomID() string {
	return uuid.New().String()[:8]
}
