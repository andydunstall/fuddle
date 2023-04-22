package registry

import (
	"math/rand"
	"sync"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
)

// Registry manages the set of registered members in the cluster.
type Registry struct {
	localID string

	members map[string]*rpc.Member2

	// priorityMembers contains the member IDs that are known to be out of date
	// so must be included as a priority in the next digest.
	priorityMembers map[string]interface{}

	// lastVersion is the last version used by this node. This is used to
	// increment the version counter when there are multiple versions in the
	// same millisecond.
	lastVersion *rpc.Version2

	// mu is a mutex protecting the fields above.
	mu sync.Mutex

	tombstoneTimeout int64

	metrics *Metrics
}

func NewRegistry(localID string, timestamp int64, opts ...Option) *Registry {
	options := defaultOptions()
	for _, o := range opts {
		o.apply(options)
	}

	metrics := NewMetrics()
	if options.collector != nil {
		metrics.Register(options.collector)
	}

	r := &Registry{
		localID:          localID,
		members:          make(map[string]*rpc.Member2),
		priorityMembers:  make(map[string]interface{}),
		tombstoneTimeout: options.tombstoneTimeout,
		metrics:          metrics,
	}

	if options.localMember != nil {
		member := &rpc.Member2{
			State:    copyMemberState(options.localMember),
			Liveness: rpc.Liveness_UP,
			Version:  r.nextVersionLocked(timestamp),
		}
		r.members[localID] = member
		r.incMembersCount(member)
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
	if m.Liveness != rpc.Liveness_UP {
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

// OwnedMemberUpsert takes ownership of the given member and adds or updates the
// members state.
func (r *Registry) OwnedMemberUpsert(memberState *rpc.MemberState, timestamp int64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Discard any update to the local member.
	if memberState.Id == r.localID {
		return
	}

	version := r.nextVersionLocked(timestamp)

	existing, ok := r.members[memberState.Id]
	if ok {
		// If the local update is before the existing member version, this
		// likely means there is clock skew between nodes, so this should never
		// happen.
		//
		// If it does, the local node will try again whenever it gets a
		// heartbeat and eventually take back ownership.
		if compareVersions(existing.Version, version) <= 0 {
			return
		}

		r.decMembersCount(existing)
	}

	member := &rpc.Member2{
		State:    copyMemberState(memberState),
		Liveness: rpc.Liveness_UP,
		Version:  version,
	}
	r.members[memberState.Id] = member

	r.incMembersCount(member)
}

// OwnedMemberLeave takes ownership of the member with the given ID and marks
// it as left with an expiry for when it should be removed.
func (r *Registry) OwnedMemberLeave(id string, timestamp int64) {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Discard any update to the local member.
	if id == r.localID {
		return
	}

	version := r.nextVersionLocked(timestamp)

	existing, ok := r.members[id]
	if ok {
		// If the local update is before the existing member version, this
		// likely means there is clock skew between nodes, so this should never
		// happen.
		//
		// If it does happen, it means we arn't the owner, so we have to just
		// discard the leave update and the current owner will mark the member
		// as down.
		if compareVersions(existing.Version, version) <= 0 {
			return
		}

		r.decMembersCount(existing)
	}

	// Mark the member as left with an expiry to be removed after the tombstone
	// timeout.
	member := &rpc.Member2{
		State:    existing.State,
		Liveness: rpc.Liveness_LEFT,
		Version:  version,
		Expiry:   timestamp + r.tombstoneTimeout,
	}
	r.members[id] = member

	r.incMembersCount(member)
}

func (r *Registry) OwnedMemberHeartbeat(id string, timestamp int64) {
	// if not owned by the local node, return an error so the client
	// re-registers

	// update last contact
}

func (r *Registry) RemoteUpsertMember(member *rpc.Member2) {
	r.mu.Lock()
	defer r.mu.Unlock()

	existing, ok := r.members[member.State.Id]
	if ok {
		// Ignore out of date remote updates.
		if compareVersions(existing.Version, member.Version) <= 0 {
			return
		}

		r.decMembersCount(existing)
	}

	r.members[member.State.Id] = copyMember(member)

	r.incMembersCount(member)

	// notify SubscribeFromDigest subscribers
	// if we are owner or lost ownership
	//    notify SubscribeLocal subscribers
}

func (r *Registry) MembersDigest(maxMembers int) map[string]*rpc.MonotonicTimestamp {
	r.mu.Lock()
	defer r.mu.Unlock()

	digest := make(map[string]*rpc.MonotonicTimestamp)

	for id := range r.priorityMembers {
		// The member may not exist, so if not use a timestamp of 0.
		if m, ok := r.members[id]; ok {
			digest[id] = &rpc.MonotonicTimestamp{
				Timestamp: m.Version.Timestamp.Timestamp,
				Counter:   m.Version.Timestamp.Counter,
			}
		} else {
			digest[id] = &rpc.MonotonicTimestamp{
				Timestamp: 0,
			}
		}
		delete(r.priorityMembers, id)

		if len(digest) >= maxMembers {
			return digest
		}
	}

	var memberIDs []string
	for id := range r.members {
		memberIDs = append(memberIDs, id)
	}
	memberIDs = shuffleStrings(memberIDs)

	for _, id := range memberIDs {
		digest[id] = &rpc.MonotonicTimestamp{
			Timestamp: r.members[id].Version.Timestamp.Timestamp,
			Counter:   r.members[id].Version.Timestamp.Counter,
		}
		if len(digest) >= maxMembers {
			return digest
		}
	}

	return digest
}

func (r *Registry) MembersDelta(digest map[string]*rpc.MonotonicTimestamp) []*rpc.Member2 {
	r.mu.Lock()
	defer r.mu.Unlock()

	// Compare our local state with the requested version in the digest.
	// If our version is greater, then we include the full member in the
	// response. Otherwise if either we don't know about the member, or the
	// sender is more up to date, then we save the member ID as a 'priority
	// member' to be included in the next digest as a priority.
	//
	// Note only compare timestamps, not owner IDs. If there is a conflict where
	// multiple nodes took ownership of a member in the same millisecond, it
	// will be resolved later so it doesn't matter which we pick.

	var members []*rpc.Member2
	for id, timestamp := range digest {
		if m, ok := r.members[id]; ok {
			if compareTimestamps(timestamp, m.Version.Timestamp) > 0 {
				members = append(members, copyMember(m))
			} else {
				r.priorityMembers[id] = struct{}{}
			}
		} else {
			r.priorityMembers[id] = struct{}{}
		}
	}

	return members
}

func (r *Registry) OnNodeJoin(id string) {
	// if the node was considered down, mark it alive again
}

func (r *Registry) OnNodeLeave(id string) {
	// if the node was considered up, mark it as down
}

func (r *Registry) nextVersionLocked(now int64) *rpc.Version2 {
	v := &rpc.Version2{
		OwnerId: r.localID,
		Timestamp: &rpc.MonotonicTimestamp{
			Timestamp: now,
			Counter:   0,
		},
	}
	// If the version has the same count as the previous version, increment
	// the counter.
	if r.lastVersion != nil && r.lastVersion.Timestamp.Timestamp == v.Timestamp.Timestamp {
		v.Timestamp.Counter = r.lastVersion.Timestamp.Counter + 1
	}
	r.lastVersion = v
	return v
}

func (r *Registry) incMembersCount(m *rpc.Member2) {
	r.metrics.MembersCount.Inc(map[string]string{
		"liveness": livenessToString(m.Liveness),
		"service":  m.State.Service,
		"owner":    m.Version.OwnerId,
	})
	if m.Version.OwnerId == r.localID {
		r.metrics.MembersOwned.Inc(map[string]string{
			"liveness": livenessToString(m.Liveness),
			"service":  m.State.Service,
		})
	}
}

func (r *Registry) decMembersCount(m *rpc.Member2) {
	r.metrics.MembersCount.Dec(map[string]string{
		"liveness": livenessToString(m.Liveness),
		"service":  m.State.Service,
		"owner":    m.Version.OwnerId,
	})
	if m.Version.OwnerId == r.localID {
		r.metrics.MembersOwned.Dec(map[string]string{
			"liveness": livenessToString(m.Liveness),
			"service":  m.State.Service,
		})
	}
}

func copyMember(m *rpc.Member2) *rpc.Member2 {
	return &rpc.Member2{
		State:    copyMemberState(m.State),
		Liveness: m.Liveness,
		Version:  copyVersion(m.Version),
	}
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

func copyVersion(m *rpc.Version2) *rpc.Version2 {
	return &rpc.Version2{
		OwnerId: m.OwnerId,
		Timestamp: &rpc.MonotonicTimestamp{
			Timestamp: m.Timestamp.Timestamp,
			Counter:   m.Timestamp.Counter,
		},
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

func compareVersions(lhs *rpc.Version2, rhs *rpc.Version2) int {
	if lhs.OwnerId != rhs.OwnerId {
		// Ignore the counter when owners don't match, as the counter only
		// applies locally.
		if lhs.Timestamp.Timestamp < rhs.Timestamp.Timestamp {
			return 1
		}
		if lhs.Timestamp.Timestamp > rhs.Timestamp.Timestamp {
			return -1
		}
		return 0
	}

	if lhs.Timestamp.Timestamp < rhs.Timestamp.Timestamp {
		return 1
	}
	if lhs.Timestamp.Timestamp > rhs.Timestamp.Timestamp {
		return -1
	}
	if lhs.Timestamp.Counter < rhs.Timestamp.Counter {
		return 1
	}
	if lhs.Timestamp.Counter > rhs.Timestamp.Counter {
		return -1
	}
	return 0
}

func compareTimestamps(lhs *rpc.MonotonicTimestamp, rhs *rpc.MonotonicTimestamp) int {
	if lhs.Timestamp < rhs.Timestamp {
		return 1
	}
	if lhs.Timestamp > rhs.Timestamp {
		return -1
	}
	if lhs.Counter < rhs.Counter {
		return 1
	}
	if lhs.Counter > rhs.Counter {
		return -1
	}
	return 0
}

func shuffleStrings(src []string) []string {
	res := make([]string, len(src))
	perm := rand.Perm(len(src))
	for i, v := range perm {
		res[v] = src[i]
	}
	return res
}
