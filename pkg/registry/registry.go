package registry

import (
	rpc "github.com/fuddle-io/fuddle-rpc/go"
)

type Registry struct {
}

func NewRegistry() *Registry {
	return &Registry{}
}

func (r *Registry) LocalUpdate(update *rpc.LocalMemberUpdate) {
}

func (r *Registry) RemoteUpdate(update *rpc.RemoteMemberUpdate) {
}

func (r *Registry) Subscribe(req *rpc.SubscribeRequest, onUpdate func(update *rpc.RemoteMemberUpdate)) {
}
