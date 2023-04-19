package registry

import (
	rpc "github.com/fuddle-io/fuddle-rpc/go"
)

// Registry manages the set of registered members in the cluster.
type Registry struct {
}

func NewRegistry() *Registry {
	return &Registry{}
}

func (r *Registry) SubscribeLocal(onUpdate func(member *rpc.Member2)) {
	// register local subscriber (onUpdate must not block)
}

func (r *Registry) SubscribeFromDigest(digest map[string]*rpc.MonotonicTimestamp, filter *rpc.ClientFilter, onUpdate func(member *rpc.Member2)) {
	// register subscriber (onUpdate must not block)

	// get a delta similar to MembersDelta and send to onUpdate
}

func (r *Registry) UpsertMember(member *rpc.Member2) {
	// update the member

	// notify SubscribeFromDigest subscribers
	// if we are owner or lost ownership
	//    notify SubscribeLocal subscribers
}

func (r *Registry) MemberLeave(id string) {
	// if we own the member set its status to 'left' with an expiry of
	// 'now + tombstone timeout'

	// notify all subscribers
}

func (r *Registry) MemberHeartbeat(id string) {
	// if not owned by the local node, take ownership and set liveness=up
	// else we already own, set liveness=up

	// update last contact
}

func (r *Registry) MembersDigest(maxMembers int) map[string]*rpc.MonotonicTimestamp {
	// select maxMembers random members and return a mapping from
	// id => version.Timestamp
	// TODO(AD) does this need to contain the owner too? probably not given if
	// theres a confict it will be resolved by heartbeats later
	return nil
}

func (r *Registry) MembersDelta(digest map[string]*rpc.MonotonicTimestamp) []*rpc.Member2 {
	// check the digest for any members we don't know about or we are out of
	// date on
	// return our own digest containing a digest of these members we need so
	// we can use push/pull

	// check the digest for any members we are more up to date on and return
	// (note as digest only partital we don't know what members the sender is
	// missing)

	return nil
}

func (r *Registry) OnNodeJoin(id string) {
	// if the node was considered down, mark it alive again
}

func (r *Registry) OnNodeLeave(id string) {
	// if the node was considered up, mark it as down
}
