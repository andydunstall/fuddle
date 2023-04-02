package gossip

// delegate is the memberlist delegate which simply returns the advertise RPC
// address of this node.
type delegate struct {
	state []byte
}

func newDelegate(state []byte) *delegate {
	return &delegate{
		state: state,
	}
}

func (d *delegate) NodeMeta(limit int) []byte {
	return d.state
}

func (d *delegate) NotifyMsg([]byte) {
}

func (d *delegate) GetBroadcasts(overhead, limit int) [][]byte {
	return nil
}

func (d *delegate) LocalState(join bool) []byte {
	return nil
}

func (d *delegate) MergeRemoteState(buf []byte, join bool) {
}
