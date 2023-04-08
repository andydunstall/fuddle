package config

import (
	"time"
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
