package config

import (
	"time"

	"go.uber.org/zap/zapcore"
)

type Registry struct {
	// HeartbeatTimeout is the time a member has to send a heartbeat after
	// their last update before they are considered down.
	HeartbeatTimeout time.Duration

	// ReconnectTimeout is the time a member has to reconnect after it was
	// marked down before they are removed from the cluster.
	ReconnectTimeout time.Duration

	// TombstoneTimeout is the time members that have left the cluster are
	// removed. This must be at least as long has the sum of the heartbeat
	// and reconnect timeouts.
	TombstoneTimeout time.Duration
}

func (c *Registry) MarshalLogObject(e zapcore.ObjectEncoder) error {
	e.AddDuration("heartbeat-timeout", c.HeartbeatTimeout)
	e.AddDuration("reconnect-timeout", c.ReconnectTimeout)
	e.AddDuration("tombstone-timeout", c.TombstoneTimeout)
	return nil
}

func DefaultRegistryConfig() *Registry {
	return &Registry{
		HeartbeatTimeout: time.Second * 20,
		ReconnectTimeout: time.Minute * 5,
		TombstoneTimeout: time.Minute * 30,
	}
}
