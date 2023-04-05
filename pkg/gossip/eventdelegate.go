package gossip

import (
	"github.com/hashicorp/memberlist"
)

type eventDelegate struct {
	onJoin  func(node Node)
	onLeave func(node Node)
}

func newEventDelegate(onJoin func(node Node), onLeave func(node Node)) *eventDelegate {
	return &eventDelegate{
		onJoin:  onJoin,
		onLeave: onLeave,
	}
}

func (d *eventDelegate) NotifyJoin(n *memberlist.Node) {
	if d.onJoin != nil {
		d.onJoin(Node{
			ID:      n.Name,
			RPCAddr: string(n.Meta),
		})
	}
}

func (d *eventDelegate) NotifyLeave(n *memberlist.Node) {
	if d.onLeave != nil {
		d.onLeave(Node{
			ID:      n.Name,
			RPCAddr: string(n.Meta),
		})
	}
}

func (d *eventDelegate) NotifyUpdate(n *memberlist.Node) {
}
