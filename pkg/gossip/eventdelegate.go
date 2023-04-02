package gossip

import (
	"github.com/hashicorp/memberlist"
)

type eventDelegate struct {
	onJoin  func(id string, addr string)
	onLeave func(id string)
}

func newEventDelegate(onJoin func(id string, addr string), onLeave func(id string)) *eventDelegate {
	return &eventDelegate{
		onJoin:  onJoin,
		onLeave: onLeave,
	}
}

func (d *eventDelegate) NotifyJoin(n *memberlist.Node) {
	if d.onJoin != nil {
		d.onJoin(n.Name, string(n.Meta))
	}
}

func (d *eventDelegate) NotifyLeave(n *memberlist.Node) {
	if d.onLeave != nil {
		d.onLeave(n.Name)
	}
}

func (d *eventDelegate) NotifyUpdate(n *memberlist.Node) {
}
