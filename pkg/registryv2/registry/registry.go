package registry

import (
	"sync"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
)

// Registry manages the set of registered members in the cluster.
type Registry struct {
	localID string

	members map[string]*rpc.Member2

	// mu is a mutex protecting the fields above.
	mu sync.Mutex

	metrics *Metrics
}

func NewRegistry(localID string, opts ...Option) *Registry {
	options := defaultOptions()
	for _, o := range opts {
		o.apply(options)
	}

	metrics := NewMetrics()
	if options.collector != nil {
		metrics.Register(options.collector)
	}

	r := &Registry{
		localID: localID,
		members: make(map[string]*rpc.Member2),
		metrics: metrics,
	}

	if options.localMember != nil {
		r.LocalMemberAdd(options.localMember)
	}

	return r
}

// Member returns the member state with the given ID, or false if it is not
// found.
func (r *Registry) Member(id string) (*rpc.MemberState, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	m, ok := r.members[id]
	if !ok {
		return nil, false
	}
	return copyMemberState(m.State), ok
}

func (r *Registry) Metrics() *Metrics {
	return r.metrics
}

func (r *Registry) SubscribeLocal(onUpdate func(member *rpc.Member2)) {
	// register local subscriber (onUpdate must not block)
}

func (r *Registry) SubscribeFromDigest(digest map[string]*rpc.MonotonicTimestamp, filter *rpc.ClientFilter, onUpdate func(member *rpc.Member2)) {
	// register subscriber (onUpdate must not block)

	// get a delta similar to MembersDelta and send to onUpdate
}

func (r *Registry) LocalMemberAdd(memberState *rpc.MemberState) {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.members[memberState.Id]
	if ok {
		r.metrics.MembersCount.Dec(map[string]string{
			"liveness": livenessToString(existing.Liveness),
			"service":  existing.State.Service,
			"owner":    existing.Version.OwnerId,
		})
		if existing.Version.OwnerId == r.localID {
			r.metrics.MembersOwned.Dec(map[string]string{
				"liveness": livenessToString(existing.Liveness),
				"service":  existing.State.Service,
			})
		}
	}

	member := &rpc.Member2{
		State:    copyMemberState(memberState),
		Liveness: rpc.Liveness_UP,
		Version: &rpc.Version2{
			OwnerId: r.localID,
		},
	}
	r.members[memberState.Id] = member

	r.metrics.MembersCount.Inc(map[string]string{
		"liveness": livenessToString(member.Liveness),
		"service":  member.State.Service,
		"owner":    member.Version.OwnerId,
	})
	if member.Version.OwnerId == r.localID {
		r.metrics.MembersOwned.Inc(map[string]string{
			"liveness": livenessToString(member.Liveness),
			"service":  member.State.Service,
		})
	}
}

func (r *Registry) LocalMemberLeave(id string) {
	// if we own the member set its status to 'left' with an expiry of
	// 'now + tombstone timeout'

	// notify all subscribers
}

func (r *Registry) LocalMemberHeartbeat(id string) {
	// if not owned by the local node, take ownership and set liveness=up
	// else we already own, set liveness=up

	// update last contact
}

func (r *Registry) RemoteUpsertMember(member *rpc.Member2) {
	// update the member

	// notify SubscribeFromDigest subscribers
	// if we are owner or lost ownership
	//    notify SubscribeLocal subscribers
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

func copyMemberState(m *rpc.MemberState) *rpc.MemberState {
	metadata := make(map[string]string)
	for k, v := range m.Metadata {
		metadata[k] = v
	}
	return &rpc.MemberState{
		Id:       m.Id,
		Status:   m.Status,
		Service:  m.Service,
		Locality: m.Locality,
		Started:  m.Started,
		Revision: m.Revision,
		Metadata: metadata,
	}
}

func livenessToString(liveness rpc.Liveness) string {
	switch liveness {
	case rpc.Liveness_UP:
		return "up"
	case rpc.Liveness_DOWN:
		return "down"
	case rpc.Liveness_LEFT:
		return "left"
	default:
		return "unknown"
	}
}
