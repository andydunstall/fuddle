package registry

import (
	"strings"
	"sync"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"go.uber.org/zap"
)

type subHandle struct {
	onUpdate  func(update *rpc.Member2)
	ownerOnly bool
}

type Registry struct {
	// localID is the node ID of this local node.
	localID string

	// members contains each member in the registry and the members version.
	//
	// This contains three types of member:
	// * This nodes local member, which cannot be updated, removed, change
	// ownership or be marked as down
	// * Owned members, which are members registered by clients connected to
	// this node, that this node is responsible for propagating updates and
	// checking member liveness
	// * Non-owned members, which are members owned by other nodes
	members map[string]*rpc.Member2

	// lastSeen contains a map of member ID to timestamp the member was last
	// seen (either updated or received a heartbeat). This only contains members
	// that are owned by the local node.
	lastSeen map[string]int64

	// subs contains a set of subscriptions.
	subs map[*subHandle]interface{}

	// lastVersion is the last version used by this node. This is used to
	// increment the version counter when there are multiple versions in the
	// same millisecond.
	lastVersion *rpc.Version2

	// leftNodes contains a map of nodes that still own members in the registry
	// but are not part of the registry.
	leftNodes map[string]int64

	// mu is a mutex protecting the fields above.
	mu sync.Mutex

	heartbeatTimeout int64
	reconnectTimeout int64
	tombstoneTimeout int64

	logger  *zap.Logger
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

	reg := &Registry{
		localID:          localID,
		members:          make(map[string]*rpc.Member2),
		lastSeen:         make(map[string]int64),
		subs:             make(map[*subHandle]interface{}),
		leftNodes:        make(map[string]int64),
		heartbeatTimeout: options.heartbeatTimeout,
		reconnectTimeout: options.reconnectTimeout,
		tombstoneTimeout: options.tombstoneTimeout,
		metrics:          metrics,
		logger:           options.logger,
	}

	if options.localMember != nil {
		member := &rpc.Member2{
			State:    options.localMember,
			Liveness: rpc.Liveness_UP,
			Version:  reg.nextVersionLocked(options.now),
		}
		reg.setMemberLocked(member)

		reg.logger.Info(
			"added local member",
			zap.Object("member", newMemberLogger(member)),
		)
	}

	return reg
}

func (r *Registry) LocalID() string {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.localID
}

// MemberState returns the member with the given ID, or false if it is not found.
func (r *Registry) MemberState(id string) (*rpc.MemberState, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	m, ok := r.members[id]
	if !ok {
		return nil, false
	}
	return m.State, ok
}

func (r *Registry) MemberStates() []*rpc.MemberState {
	r.mu.Lock()
	defer r.mu.Unlock()

	members := make([]*rpc.MemberState, 0, len(r.members))
	for _, m := range r.members {
		members = append(members, m.State)
	}
	return members
}

func (r *Registry) Member(id string) (*rpc.Member2, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	m, ok := r.members[id]
	if !ok {
		return nil, false
	}
	return m, ok
}

func (r *Registry) Members() []*rpc.Member2 {
	r.mu.Lock()
	defer r.mu.Unlock()

	members := make([]*rpc.Member2, 0, len(r.members))
	for _, m := range r.members {
		members = append(members, m)
	}
	return members
}

func (r *Registry) OwnedMembers() []*rpc.Member2 {
	r.mu.Lock()
	defer r.mu.Unlock()

	members := make([]*rpc.Member2, 0, len(r.members))
	for _, m := range r.members {
		if m.Version.OwnerId == r.localID {
			members = append(members, m)
		}
	}
	return members
}

func (r *Registry) UpMembers() []*rpc.Member2 {
	r.mu.Lock()
	defer r.mu.Unlock()

	members := make([]*rpc.Member2, 0, len(r.members))
	for _, m := range r.members {
		if m.Liveness == rpc.Liveness_UP {
			members = append(members, m)
		}
	}
	return members
}

func (r *Registry) Metrics() *Metrics {
	return r.metrics
}

func (r *Registry) SubscribeLocal(onUpdate func(update *rpc.Member2)) func() {
	r.mu.Lock()
	defer r.mu.Unlock()

	handle := &subHandle{
		onUpdate:  onUpdate,
		ownerOnly: true,
	}
	r.subs[handle] = struct{}{}

	return func() {
		r.mu.Lock()
		defer r.mu.Unlock()

		delete(r.subs, handle)
	}
}

// Subscribe to member updates.
func (r *Registry) Subscribe(req *rpc.SubscribeRequest, onUpdate func(update *rpc.Member2), opts ...Option) func() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if req == nil {
		req = &rpc.SubscribeRequest{
			OwnerOnly: false,
		}
	}

	handle := &subHandle{
		onUpdate:  onUpdate,
		ownerOnly: req.OwnerOnly,
	}
	r.subs[handle] = struct{}{}

	for _, update := range r.updatesLocked(req) {
		onUpdate(update)
	}

	return func() {
		r.mu.Lock()
		defer r.mu.Unlock()

		delete(r.subs, handle)
	}
}

func (r *Registry) Updates(req *rpc.SubscribeRequest) []*rpc.Member2 {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.updatesLocked(req)
}

func (r *Registry) OnNodeJoin(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.logger.Info("on node join", zap.String("id", id))

	delete(r.leftNodes, id)
}

func (r *Registry) OnNodeLeave(id string, opts ...Option) {
	options := defaultOptions()
	for _, o := range opts {
		o.apply(options)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	r.logger.Info(
		"on node leave",
		zap.String("id", id),
		zap.Int64("timestamp", options.now),
	)

	memberCount := r.membersForOwnerLocked(id)
	if memberCount > 0 {
		r.logger.Warn(
			"node left while owning members",
			zap.String("id", id),
			zap.Int("member-count", memberCount),
			zap.Int64("timestamp", options.now),
		)

		r.leftNodes[id] = options.now
	}
}

// AddMember adds a member that is owned by this node.
//
// The member is re-added whenever we receive a heartbeat for the member, which
// will update the members status to UP if it was down.
func (r *Registry) AddMember(member *rpc.MemberState, opts ...Option) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.updateMemberLocked(member, rpc.Liveness_UP, 0, opts...)
}

// MemberHeartbeat updates the last seen timestamp for the member.
//
// If we've lost ownership of the member (such as if we had networking issues
// so another node took ownership, but the client is still connected to this
// node), then update the member version and status to take back ownership.
func (r *Registry) MemberHeartbeat(member *rpc.MemberState, opts ...Option) {
	options := defaultOptions()
	for _, o := range opts {
		o.apply(options)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	// If we own the node and it has a status of UP, theres no need to propagate
	// the update to other nodes, so just update the last seen timestamp.
	// Otherwise update the member with a status of UP to take back ownership
	// and update the members status.
	existing, ok := r.members[member.Id]
	if ok && existing.Version.OwnerId == r.localID && existing.Liveness == rpc.Liveness_UP {
		r.lastSeen[member.Id] = options.now
	} else {
		r.updateMemberLocked(member, rpc.Liveness_UP, 0, opts...)
	}
}

// RemoveMember removes the member with the given ID.
//
// Note this won't actually remove it, but instead mark it as left. This is
// used as a tombstone to ensure the update is propagated before nodes actually
// remove the node.
func (r *Registry) RemoveMember(id string, opts ...Option) {
	options := defaultOptions()
	for _, o := range opts {
		o.apply(options)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	m, ok := r.members[id]
	if !ok {
		r.logger.Warn("remove member; not found", zap.String("id", id))
		return
	}

	r.updateMemberLocked(
		m.State,
		rpc.Liveness_LEFT,
		options.now+r.tombstoneTimeout,
		opts...,
	)
}

// RemoteUpdate applies an updates received from another node.
func (r *Registry) RemoteUpdate(update *rpc.Member2) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if update.State.Id == r.localID {
		r.logger.Error(
			"remote update: discarding update; attempted to update local member",
			zap.Object("member", newMemberLogger(update)),
		)
		return
	}

	if update.Version.OwnerId == r.localID {
		r.logger.Error(
			"remote update: discarding update; same owner as local node",
			zap.Object("update", newMemberLogger(update)),
		)
		return
	}

	ownershipChange := false

	existing, ok := r.members[update.State.Id]
	if ok {
		if compareVersions(existing.Version, update.Version) <= 0 {
			r.logger.Error(
				"discarding remote member update; outdated version",
				zap.Object("update", newMemberLogger(update)),
				zap.Object("update-version", newVersionLogger(update.Version)),
				zap.Object("existing-version", newVersionLogger(existing.Version)),
			)
			return
		}

		if existing.Version.OwnerId == r.localID {
			ownershipChange = true
		}
	}

	r.setMemberLocked(update)

	r.logger.Info(
		"updated member; remote",
		zap.Object("update", newMemberLogger(update)),
	)

	r.notifySubscribersLocked(update, ownershipChange)
}

func (r *Registry) updateMemberLocked(member *rpc.MemberState, liveness rpc.Liveness, expiry int64, opts ...Option) {
	// TODO copy state

	options := defaultOptions()
	for _, o := range opts {
		o.apply(options)
	}

	// Don't allow updating the local member.
	if member.Id == r.localID {
		r.logger.Error(
			"attempted to update local member",
			zap.Object("member", newMemberStateLogger(member)),
		)
		return
	}

	version := r.nextVersionLocked(options.now)

	// If the version is out of date we must ignore the update. This should
	// never happen as it means there is clock skew between nodes or a non-
	// monotonic clock.
	//
	// If it does happen, the owner will get regular heartbeats where it keeps
	// re-adding the member, so will eventually win ownership.
	existing, exists := r.members[member.Id]
	if exists {
		if compareVersions(existing.Version, version) <= 0 {
			r.logger.Error(
				"discarding local member update; outdated version",
				zap.Object("member", newMemberStateLogger(member)),
				zap.Object("update-version", newVersionLogger(version)),
				zap.Object("existing-version", newVersionLogger(existing.Version)),
			)
			return
		}
	}

	versionedMember := &rpc.Member2{
		State:    member,
		Liveness: liveness,
		Version:  version,
		Expiry:   expiry,
	}
	r.setMemberLocked(versionedMember)

	r.logger.Info(
		"updated member; owner",
		zap.Bool("owner", true),
		zap.Object("member", newMemberLogger(versionedMember)),
	)

	r.notifySubscribersLocked(versionedMember, true)
}

func (r *Registry) notifySubscribersLocked(update *rpc.Member2, owner bool) {
	for s := range r.subs {
		if s.ownerOnly && owner {
			s.onUpdate(update)
		} else if !s.ownerOnly {
			s.onUpdate(update)
		}
	}
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

func (r *Registry) membersForOwnerLocked(id string) int {
	count := 0
	for _, m := range r.members {
		if m.Version.OwnerId == id {
			count++
		}
	}
	return count
}

func (r *Registry) updatesLocked(req *rpc.SubscribeRequest) []*rpc.Member2 {
	if req.KnownMembers == nil {
		req.KnownMembers = make(map[string]*rpc.Version2)
	}

	if req.OwnerOnly {
		return r.ownerOnlyUpdatesLocked(req.KnownMembers)
	}
	return r.allUpdatesLocked(req.KnownMembers)
}

func (r *Registry) ownerOnlyUpdatesLocked(knownMembers map[string]*rpc.Version2) []*rpc.Member2 {
	var updates []*rpc.Member2
	for id, m := range r.members {
		knownVersion, ok := knownMembers[id]
		if ok {
			// If either we own the member, or the subscriber thinks we do,
			// and we have a more recent version, send an update.
			if m.Version.OwnerId == r.localID || knownVersion.OwnerId == r.localID {
				if compareVersions(knownVersion, m.Version) > 0 {
					updates = append(updates, m)
				}
			}
		} else {
			// If the subscriber doesn't know abou tthe member, send an update.
			if m.Version.OwnerId == r.localID {
				updates = append(updates, m)
			}
		}
	}

	for id, knownVersion := range knownMembers {
		if _, ok := r.members[id]; ok {
			continue
		}
		if knownVersion.OwnerId != r.localID {
			continue
		}

		// This should never happen, where the subscriber thinks we own a member
		// that we don't. It should be prevented by tombstones.

		r.logger.Error(
			"subscriber knows about a member that is not in the cluster",
			zap.Object("known-version", newVersionLogger(knownVersion)),
		)

		// So the subscriber knows the member has left, send the smallest
		// version that will remove the member, but won't conflict with
		// a more recent version from another owner, by keeping the same
		// timestamp but incrementing the counter.
		updates = append(updates, &rpc.Member2{
			State: &rpc.MemberState{
				Id: id,
			},
			Liveness: rpc.Liveness_LEFT,
			Version: &rpc.Version2{
				OwnerId: knownVersion.OwnerId,
				Timestamp: &rpc.MonotonicTimestamp{
					Timestamp: knownVersion.Timestamp.Timestamp,
					Counter:   knownVersion.Timestamp.Counter + 1,
				},
			},
		})
	}

	return updates
}

func (r *Registry) allUpdatesLocked(knownMembers map[string]*rpc.Version2) []*rpc.Member2 {
	var updates []*rpc.Member2
	for id, m := range r.members {
		knownVersion, ok := knownMembers[id]
		if ok {
			// If either we own the member, or the subscriber thinks we do,
			// and we have a more recent version, send an update.
			if compareVersions(knownVersion, m.Version) > 0 {
				updates = append(updates, m)
			}
		} else {
			// If the subscriber doesn't know abou tthe member, send an update.
			updates = append(updates, m)
		}
	}

	for id, knownVersion := range knownMembers {
		if _, ok := r.members[id]; ok {
			continue
		}

		// This should never happen, where the subscriber thinks we own a member
		// that we don't. It should be prevented by tombstones.

		r.logger.Error(
			"subscriber knows about a member that is not in the cluster",
			zap.Object("known-version", newVersionLogger(knownVersion)),
		)

		// So the subscriber knows the member has left, send the smallest
		// version that will remove the member, but won't conflict with
		// a more recent version from another owner, by keeping the same
		// timestamp but incrementing the counter.
		updates = append(updates, &rpc.Member2{
			State: &rpc.MemberState{
				Id: id,
			},
			Liveness: rpc.Liveness_LEFT,
			Version: &rpc.Version2{
				OwnerId: knownVersion.OwnerId,
				Timestamp: &rpc.MonotonicTimestamp{
					Timestamp: knownVersion.Timestamp.Timestamp,
					Counter:   knownVersion.Timestamp.Counter + 1,
				},
			},
		})
	}

	return updates
}

func (r *Registry) setMemberLocked(m *rpc.Member2) {
	if existing, ok := r.members[m.State.Id]; ok {
		r.metrics.MembersCount.Dec(map[string]string{
			"status": strings.ToLower(existing.Liveness.String()),
			"owner":  existing.Version.OwnerId,
		})
		if existing.Version.OwnerId == r.localID {
			r.metrics.MembersOwned.Dec(map[string]string{
				"status": strings.ToLower(existing.Liveness.String()),
			})
		}
	}

	r.members[m.State.Id] = m

	r.metrics.MembersCount.Inc(map[string]string{
		"status": strings.ToLower(m.Liveness.String()),
		"owner":  m.Version.OwnerId,
	})
	if m.Version.OwnerId == r.localID {
		r.metrics.MembersOwned.Inc(map[string]string{
			"status": strings.ToLower(m.Liveness.String()),
		})

	}

	if m.Version.OwnerId == r.localID {
		r.lastSeen[m.State.Id] = m.Version.Timestamp.Timestamp
	} else {
		delete(r.lastSeen, m.State.Id)
	}
}

func (r *Registry) deleteMemberLocked(id string) {
	if existing, ok := r.members[id]; ok {
		r.metrics.MembersCount.Dec(map[string]string{
			"status": strings.ToLower(existing.Liveness.String()),
			"owner":  existing.Version.OwnerId,
		})

		if existing.Version.OwnerId == r.localID {
			r.metrics.MembersOwned.Dec(map[string]string{
				"status": strings.ToLower(existing.Liveness.String()),
			})
		}
	}

	delete(r.members, id)
	delete(r.lastSeen, id)
}

// compareVersions compares lhs and rhs.
//
// If the owners don't match but the timestamps do, lhs is considered greater.
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
		// If the owners don't match, but the timestamps do, favor lhs.
		return -1
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
