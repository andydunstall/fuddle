package registry

import (
	"sync"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"go.uber.org/zap"
	"google.golang.org/protobuf/proto"
)

type subHandle struct {
	onUpdate  func(update *rpc.RemoteMemberUpdate)
	ownerOnly bool
}

type VersionedMember struct {
	Member  *rpc.Member
	Version *rpc.Version
}

func (m *VersionedMember) Equal(o *VersionedMember) bool {
	return proto.Equal(m.Member, o.Member) && proto.Equal(m.Version, o.Version)
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
	members map[string]*VersionedMember

	// lastSeen contains a map of member ID to timestamp the member was last
	// seen (either updated or received a heartbeat). This only contains members
	// that are owned by the local node.
	lastSeen map[string]int64

	// subs contains a set of subscriptions.
	subs map[*subHandle]interface{}

	// lastVersion is the last version used by this node. This is used to
	// increment the version counter when there are multiple versions in the
	// same millisecond.
	lastVersion *rpc.Version

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
		members:          make(map[string]*VersionedMember),
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
		member := memberWithStatus(options.localMember, rpc.MemberStatus_UP)
		versionedMember := &VersionedMember{
			Member:  member,
			Version: reg.nextVersionLocked(options.now),
		}
		reg.setMemberLocked(versionedMember)

		reg.logger.Info(
			"added local member",
			zap.Object("member", newMemberLogger(versionedMember.Member)),
			zap.Object("version", newVersionLogger(versionedMember.Version)),
		)
	}

	return reg
}

// Member returns the member with the given ID, or false if it is not found.
func (r *Registry) Member(id string) (*rpc.Member, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	m, ok := r.members[id]
	if !ok {
		return nil, false
	}
	return m.Member, ok
}

func (r *Registry) Members() []*rpc.Member {
	r.mu.Lock()
	defer r.mu.Unlock()

	members := make([]*rpc.Member, 0, len(r.members))
	for _, m := range r.members {
		members = append(members, m.Member)
	}
	return members
}

func (r *Registry) VersionedMembers() []*VersionedMember {
	r.mu.Lock()
	defer r.mu.Unlock()

	members := make([]*VersionedMember, 0, len(r.members))
	for _, m := range r.members {
		members = append(members, m)
	}
	return members
}

func (r *Registry) UpMembers() []*VersionedMember {
	r.mu.Lock()
	defer r.mu.Unlock()

	members := make([]*VersionedMember, 0, len(r.members))
	for _, m := range r.members {
		if m.Member.Status == rpc.MemberStatus_UP {
			members = append(members, m)
		}
	}
	return members
}

func (r *Registry) Metrics() *Metrics {
	return r.metrics
}

// Subscribe to member updates.
func (r *Registry) Subscribe(req *rpc.SubscribeRequest, onUpdate func(update *rpc.RemoteMemberUpdate), opts ...Option) func() {
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

func (r *Registry) Updates(req *rpc.SubscribeRequest) []*rpc.RemoteMemberUpdate {
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
func (r *Registry) AddMember(member *rpc.Member, opts ...Option) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.updateMemberLocked(memberWithStatus(member, rpc.MemberStatus_UP), opts...)
}

// MemberHeartbeat updates the last seen timestamp for the member.
//
// If we've lost ownership of the member (such as if we had networking issues
// so another node took ownership, but the client is still connected to this
// node), then update the member version and status to take back ownership.
func (r *Registry) MemberHeartbeat(member *rpc.Member, opts ...Option) {
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
	if ok && existing.Version.Owner == r.localID && existing.Member.Status == rpc.MemberStatus_UP {
		r.lastSeen[member.Id] = options.now
	} else {
		r.updateMemberLocked(memberWithStatus(member, rpc.MemberStatus_UP), opts...)
	}
}

// RemoveMember removes the member with the given ID.
//
// Note this won't actually remove it, but instead mark it as left. This is
// used as a tombstone to ensure the update is propagated before nodes actually
// remove the node.
func (r *Registry) RemoveMember(id string, opts ...Option) {
	r.mu.Lock()
	defer r.mu.Unlock()

	m, ok := r.members[id]
	if !ok {
		r.logger.Warn("remove member; not found", zap.String("id", id))
		return
	}

	member := m.Member
	r.updateMemberLocked(memberWithStatus(member, rpc.MemberStatus_LEFT), opts...)
}

// RemoteUpdate applies an updates received from another node.
func (r *Registry) RemoteUpdate(update *rpc.RemoteMemberUpdate) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if update.Member.Id == r.localID {
		r.logger.Error(
			"remote update: discarding update; attempted to update local member",
			zap.Object("update", newRemoteMemberUpdateLogger(update)),
		)
		return
	}

	if update.Version.Owner == r.localID {
		r.logger.Error(
			"remote update: discarding update; same owner as local node",
			zap.Object("update", newRemoteMemberUpdateLogger(update)),
		)
		return
	}

	ownershipChange := false

	existing, ok := r.members[update.Member.Id]
	if ok {
		if compareVersions(existing.Version, update.Version) <= 0 {
			r.logger.Error(
				"discarding remote member update; outdated version",
				zap.Object("update", newRemoteMemberUpdateLogger(update)),
				zap.Object("update-version", newVersionLogger(update.Version)),
				zap.Object("existing-version", newVersionLogger(existing.Version)),
			)
			return
		}

		if existing.Version.Owner == r.localID {
			ownershipChange = true
		}
	}

	r.setMemberLocked(&VersionedMember{
		Member:  update.Member,
		Version: update.Version,
	})

	r.logger.Info(
		"updated member; remote",
		zap.Object("update", newRemoteMemberUpdateLogger(update)),
	)

	r.notifySubscribersLocked(update, ownershipChange)
}

func (r *Registry) CheckMembersLiveness(opts ...Option) {
	r.mu.Lock()
	defer r.mu.Unlock()

	for id := range r.members {
		r.checkMemberLivenessLocked(id, opts...)
	}

	for id := range r.leftNodes {
		// If we've taken away all members for a left node, it can be
		// discarded.
		if r.membersForOwnerLocked(id) == 0 {
			r.logger.Info("removing left node; no remaining members")
			delete(r.leftNodes, id)
		}
	}
}

func (r *Registry) updateMemberLocked(member *rpc.Member, opts ...Option) {
	options := defaultOptions()
	for _, o := range opts {
		o.apply(options)
	}

	// Don't allow updating the local member.
	if member.Id == r.localID {
		r.logger.Error(
			"attempted to update local member",
			zap.Object("member", newMemberLogger(member)),
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
				zap.Object("member", newMemberLogger(member)),
				zap.Object("update-version", newVersionLogger(version)),
				zap.Object("existing-version", newVersionLogger(existing.Version)),
			)
			return
		}
	}

	versionedMember := &VersionedMember{
		Member:  member,
		Version: version,
	}
	r.setMemberLocked(versionedMember)

	r.logger.Info(
		"updated member; owner",
		zap.Bool("owner", true),
		zap.Object("member", newMemberLogger(versionedMember.Member)),
		zap.Object("version", newVersionLogger(versionedMember.Version)),
	)

	r.notifySubscribersLocked(&rpc.RemoteMemberUpdate{
		Member:  versionedMember.Member,
		Version: versionedMember.Version,
	}, true)
}

func (r *Registry) checkMemberLivenessLocked(id string, opts ...Option) {
	// The local member is always up.
	if id == r.localID {
		return
	}

	m := r.members[id]
	if m.Version.Owner == r.localID {
		r.checkOwnedMemberLivenessLocked(id, opts...)
	} else {
		r.checkRemoteMemberLivenessLocked(id, opts...)
	}
}

func (r *Registry) checkOwnedMemberLivenessLocked(id string, opts ...Option) {
	options := defaultOptions()
	for _, o := range opts {
		o.apply(options)
	}

	m := r.members[id]

	lastSeen := r.lastSeen[id]

	if m.Member.Status == rpc.MemberStatus_LEFT {
		if options.now-lastSeen > r.tombstoneTimeout {
			r.logger.Info(
				"removing left member",
				zap.Int64("last-update", options.now-lastSeen),
			)
			r.deleteMemberLocked(m.Member.Id)
		}
	}

	if m.Member.Status == rpc.MemberStatus_DOWN {
		if options.now-lastSeen > r.reconnectTimeout {
			r.logger.Info(
				"member removed after missing heartbeats",
				zap.Int64("last-update", options.now-lastSeen),
			)
			r.updateMemberLocked(
				memberWithStatus(m.Member, rpc.MemberStatus_LEFT), opts...,
			)
		}
	}

	// If the last contact from the member exceeds the heartbeat timeout,
	// mark the member down.
	if m.Member.Status == rpc.MemberStatus_UP {
		if options.now-lastSeen > r.heartbeatTimeout {
			r.logger.Info(
				"member down after missing heartbeats",
				zap.Int64("last-update", options.now-lastSeen),
			)
			r.updateMemberLocked(
				memberWithStatus(m.Member, rpc.MemberStatus_DOWN), opts...,
			)
		}
	}
}

func (r *Registry) checkRemoteMemberLivenessLocked(id string, opts ...Option) {
	options := defaultOptions()
	for _, o := range opts {
		o.apply(options)
	}

	m := r.members[id]

	// If the owner of the node is still in the cluster, do nothing.
	ownerLastContact, ok := r.leftNodes[m.Version.Owner]
	if !ok {
		return
	}

	// If the owner has left the cluster for more than the heartbeat timeout,
	// try to take ownership of the member. This may lead to nodes competing
	// for ownership, which is ok as one will quickly win.
	if options.now-ownerLastContact > r.heartbeatTimeout {
		// If the member is up, mark it down as it has missed the heartbeat
		// timeout.
		if m.Member.Status == rpc.MemberStatus_UP {
			r.updateMemberLocked(
				memberWithStatus(m.Member, rpc.MemberStatus_DOWN),
				opts...,
			)
		} else {
			r.updateMemberLocked(m.Member, opts...)
		}
	}
}

func (r *Registry) notifySubscribersLocked(update *rpc.RemoteMemberUpdate, owner bool) {
	for s := range r.subs {
		if s.ownerOnly && owner {
			s.onUpdate(update)
		} else if !s.ownerOnly {
			s.onUpdate(update)
		}
	}
}

func (r *Registry) nextVersionLocked(now int64) *rpc.Version {
	v := &rpc.Version{
		Owner:     r.localID,
		Timestamp: now,
		Counter:   0,
	}
	// If the version has the same count as the previous version, increment
	// the counter.
	if r.lastVersion != nil && r.lastVersion.Timestamp == v.Timestamp {
		v.Counter = r.lastVersion.Counter + 1
	}
	r.lastVersion = v
	return v
}

func (r *Registry) membersForOwnerLocked(id string) int {
	count := 0
	for _, m := range r.members {
		if m.Version.Owner == id {
			count++
		}
	}
	return count
}

func (r *Registry) updatesLocked(req *rpc.SubscribeRequest) []*rpc.RemoteMemberUpdate {
	if req.KnownMembers == nil {
		req.KnownMembers = make(map[string]*rpc.Version)
	}

	if req.OwnerOnly {
		return r.ownerOnlyUpdatesLocked(req.KnownMembers)
	}
	return r.allUpdatesLocked(req.KnownMembers)
}

func (r *Registry) ownerOnlyUpdatesLocked(knownMembers map[string]*rpc.Version) []*rpc.RemoteMemberUpdate {
	var updates []*rpc.RemoteMemberUpdate
	for id, m := range r.members {
		knownVersion, ok := knownMembers[id]
		if ok {
			// If either we own the member, or the subscriber thinks we do,
			// and we have a more recent version, send an update.
			if m.Version.Owner == r.localID || knownVersion.Owner == r.localID {
				if compareVersions(knownVersion, m.Version) > 0 {
					updates = append(updates, &rpc.RemoteMemberUpdate{
						Member:  m.Member,
						Version: m.Version,
					})
				}
			}
		} else {
			// If the subscriber doesn't know abou tthe member, send an update.
			if m.Version.Owner == r.localID {
				updates = append(updates, &rpc.RemoteMemberUpdate{
					Member:  m.Member,
					Version: m.Version,
				})
			}
		}
	}

	for id, knownVersion := range knownMembers {
		if _, ok := r.members[id]; ok {
			continue
		}
		if knownVersion.Owner != r.localID {
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
		updates = append(updates, &rpc.RemoteMemberUpdate{
			Member: &rpc.Member{
				Id:     id,
				Status: rpc.MemberStatus_LEFT,
			},
			Version: &rpc.Version{
				Owner:     knownVersion.Owner,
				Timestamp: knownVersion.Timestamp,
				Counter:   knownVersion.Counter + 1,
			},
		})
	}

	return updates
}

func (r *Registry) allUpdatesLocked(knownMembers map[string]*rpc.Version) []*rpc.RemoteMemberUpdate {
	var updates []*rpc.RemoteMemberUpdate
	for id, m := range r.members {
		knownVersion, ok := knownMembers[id]
		if ok {
			// If either we own the member, or the subscriber thinks we do,
			// and we have a more recent version, send an update.
			if compareVersions(knownVersion, m.Version) > 0 {
				updates = append(updates, &rpc.RemoteMemberUpdate{
					Member:  m.Member,
					Version: m.Version,
				})
			}
		} else {
			// If the subscriber doesn't know abou tthe member, send an update.
			updates = append(updates, &rpc.RemoteMemberUpdate{
				Member:  m.Member,
				Version: m.Version,
			})
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
		updates = append(updates, &rpc.RemoteMemberUpdate{
			Member: &rpc.Member{
				Id:     id,
				Status: rpc.MemberStatus_LEFT,
			},
			Version: &rpc.Version{
				Owner:     knownVersion.Owner,
				Timestamp: knownVersion.Timestamp,
				Counter:   knownVersion.Counter + 1,
			},
		})
	}

	return updates
}

func (r *Registry) setMemberLocked(m *VersionedMember) {
	if existing, ok := r.members[m.Member.Id]; ok {
		r.metrics.MembersCount.Dec(map[string]string{
			"status": existing.Member.Status.String(),
			"owner":  existing.Version.Owner,
		})
		if existing.Version.Owner == r.localID {
			r.metrics.MembersOwned.Dec(map[string]string{
				"status": existing.Member.Status.String(),
			})
		}
	}

	r.members[m.Member.Id] = m

	r.metrics.MembersCount.Inc(map[string]string{
		"status": m.Member.Status.String(),
		"owner":  m.Version.Owner,
	})
	if m.Version.Owner == r.localID {
		r.metrics.MembersOwned.Inc(map[string]string{
			"status": m.Member.Status.String(),
		})

	}

	if m.Version.Owner == r.localID {
		r.lastSeen[m.Member.Id] = m.Version.Timestamp
	} else {
		delete(r.lastSeen, m.Member.Id)
	}
}

func (r *Registry) deleteMemberLocked(id string) {
	if existing, ok := r.members[id]; ok {
		r.metrics.MembersCount.Dec(map[string]string{
			"status": existing.Member.Status.String(),
			"owner":  existing.Version.Owner,
		})

		if existing.Version.Owner == r.localID {
			r.metrics.MembersOwned.Dec(map[string]string{
				"status": existing.Member.Status.String(),
			})
		}
	}

	delete(r.members, id)
	delete(r.lastSeen, id)
}

// compareVersions compares lhs and rhs.
//
// If the owners don't match but the timestamps do, lhs is considered greater.
func compareVersions(lhs *rpc.Version, rhs *rpc.Version) int {
	if lhs.Owner != rhs.Owner {
		// Ignore the counter when owners don't match, as the counter only
		// applies locally.
		if lhs.Timestamp < rhs.Timestamp {
			return 1
		}
		if lhs.Timestamp > rhs.Timestamp {
			return -1
		}
		// If the owners don't match, but the timestamps do, favor lhs.
		return -1
	}

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

func memberWithStatus(m *rpc.Member, status rpc.MemberStatus) *rpc.Member {
	cp := copyMember(m)
	cp.Status = status
	return cp
}

func copyMember(m *rpc.Member) *rpc.Member {
	metadata := make(map[string]string)
	for k, v := range m.Metadata {
		metadata[k] = v
	}
	return &rpc.Member{
		Id:       m.Id,
		Status:   m.Status,
		Service:  m.Service,
		Locality: m.Locality,
		Created:  m.Created,
		Revision: m.Revision,
		Metadata: metadata,
	}
}
