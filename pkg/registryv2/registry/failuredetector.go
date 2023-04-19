package registry

// FailureDetector detects when members are no longer responding and marks them
// as down.
type FailureDetector struct {
}

func NewFailureDetector(registry *Registry) *FailureDetector {
	return &FailureDetector{}
}

func (fd *FailureDetector) Check() {
	// mark any members missing heartbeats as 'down' with an expiry of
	// now + reconnect timeout

	// mark any members registered to a missing node, where the node has been
	// down for the heartbeat timeout, as 'down' with same expiry as above
	// we'll try to take ownership (and may lose which is fine)
}
