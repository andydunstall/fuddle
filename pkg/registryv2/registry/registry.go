package registry

import (
	rpc "github.com/fuddle-io/fuddle-rpc/go"
)

// Registry manages the set of registered members in the cluster.
type Registry struct {
}

func (r *Registry) SubscribeLocal(onUpdate func(member *rpc.Member2)) {
}

func (r *Registry) SubscribeFromDigest(digest map[string]*rpc.MonotonicTimestamp, filter *rpc.ClientFilter, onUpdate func(member *rpc.Member2)) {
}

func (r *Registry) UpsertMember(member *rpc.Member2) {
}

func (r *Registry) MemberLeave(id string) {
}

func (r *Registry) MemberHeartbeat(id string) {
}

func (r *Registry) MembersDigest(maxMembers int) map[string]*rpc.MonotonicTimestamp {
	return nil
}

func (r *Registry) MembersDelta(digest map[string]*rpc.MonotonicTimestamp) []*rpc.Member2 {
	return nil
}

func (r *Registry) OnNodeJoin(id string) {
}

func (r *Registry) OnNodeLeave(id string) {
}
