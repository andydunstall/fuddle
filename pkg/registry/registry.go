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
	members map[string]*VersionedMember

	localID string

	subs map[*subHandle]interface{}

	lastVersion *rpc.Version

	// mu is a mutex protecting the above fields.
	mu sync.Mutex

	logger *zap.Logger
}

func NewRegistry(localID string, opts ...Option) *Registry {
	options := defaultRegistryOptions()
	for _, o := range opts {
		o.apply(options)
	}

	r := &Registry{
		members: make(map[string]*VersionedMember),
		localID: localID,
		subs:    make(map[*subHandle]interface{}),
		logger:  options.logger,
	}

	if options.localMember != nil {
		localMember := &VersionedMember{
			Member:  options.localMember,
			Version: r.nextVersionLocked(options.now),
		}
		r.members[localID] = localMember

		r.logger.Info(
			"registered local member",
			zap.Object("member", newMemberLogger(localMember.Member)),
			zap.Object("version", newVersionLogger(localMember.Version)),
		)
	}

	return r
}

func (r *Registry) Member(id string) (*VersionedMember, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	m, ok := r.members[id]
	return m, ok
}

func (r *Registry) Members() []*VersionedMember {
	r.mu.Lock()
	defer r.mu.Unlock()

	members := make([]*VersionedMember, 0, len(r.members))
	for _, m := range r.members {
		members = append(members, m)
	}
	return members
}

func (r *Registry) Updates(req *rpc.SubscribeRequest, opts ...Option) []*rpc.RemoteMemberUpdate {
	options := defaultRegistryOptions()
	for _, o := range opts {
		o.apply(options)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	return r.updatesLocked(req, options.now)
}

func (r *Registry) LocalRegister(member *rpc.Member, opts ...Option) {
	r.mu.Lock()
	defer r.mu.Unlock()

	member.Status = rpc.MemberStatus_UP
	r.localRegisterLocked(member, opts...)
}

func (r *Registry) LocalUnregister(id string, opts ...Option) {
	r.mu.Lock()
	defer r.mu.Unlock()

	m, ok := r.members[id]
	if !ok {
		return
	}

	m.Member.Status = rpc.MemberStatus_LEFT
	r.localRegisterLocked(m.Member, opts...)
}

func (r *Registry) LocalUpdate(update *rpc.LocalMemberUpdate, opts ...Option) {
	r.mu.Lock()
	defer r.mu.Unlock()

	options := defaultRegistryOptions()
	for _, o := range opts {
		o.apply(options)
	}

	if update.Member.Id == r.localID {
		r.logger.Error(
			"attempted to update local member",
			zap.Object("update", newLocalMemberUpdateLogger(update)),
		)
		return
	}

	v := r.nextVersionLocked(options.now)

	existing, exists := r.members[update.Member.Id]
	if exists {
		if existing.Version.Owner != v.Owner {
			// Compare timestamps to choose a winner. Note ignore the counter
			// since the counter only applies locally.
			//
			// This should never happen! It likely means there is clock skew
			// between nodes.
			if existing.Version.Timestamp > v.Timestamp {
				r.logger.Error(
					"discarding local update; outdated version",
					zap.Object("update", newLocalMemberUpdateLogger(update)),
					zap.Object("update-version", newVersionLogger(v)),
					zap.Object("existing-version", newVersionLogger(existing.Version)),
				)
				return
			}
		}
	}

	switch update.UpdateType {
	case rpc.MemberUpdateType_REGISTER:
		versionedMember := &VersionedMember{
			Member:  update.Member,
			Version: v,
		}
		r.members[update.Member.Id] = versionedMember

		r.logger.Info(
			"registered member; owner",
			zap.Object("update", newLocalMemberUpdateLogger(update)),
			zap.Object("member", newMemberLogger(versionedMember.Member)),
			zap.Object("update-version", newVersionLogger(versionedMember.Version)),
		)

		r.notifySubscribersLocked(&rpc.RemoteMemberUpdate{
			UpdateType: rpc.MemberUpdateType_REGISTER,
			Member:     versionedMember.Member,
			Version:    versionedMember.Version,
		}, true)

	case rpc.MemberUpdateType_UNREGISTER:
		// If the member isn't registered do nothing.
		if !exists {
			r.logger.Warn(
				"discarding unregister; node doesn't exist",
				zap.Object("update", newLocalMemberUpdateLogger(update)),
			)
			return
		}

		// If the member does exist but we arn't the owner, this is an error as
		// only the owner should receive unregister updates.
		if existing.Version.Owner != r.localID {
			r.logger.Error(
				"discarding unregister; local node is not the owner",
				zap.Object("update", newLocalMemberUpdateLogger(update)),
				zap.Object("update-version", newVersionLogger(v)),
				zap.Object("existing-version", newVersionLogger(existing.Version)),
			)
			return
		}

		delete(r.members, update.Member.Id)

		r.logger.Info(
			"unregistered member; owner",
			zap.Object("update", newLocalMemberUpdateLogger(update)),
			zap.Object("member", newMemberLogger(existing.Member)),
			zap.Object("update-version", newVersionLogger(v)),
		)

		r.notifySubscribersLocked(&rpc.RemoteMemberUpdate{
			UpdateType: rpc.MemberUpdateType_UNREGISTER,
			Member:     update.Member,
			Version:    v,
		}, true)
	}
}

func (r *Registry) RemoteUpdate(update *rpc.RemoteMemberUpdate) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if update.Member.Id == r.localID {
		r.logger.Error(
			"attempted to update local member",
			zap.Object("update", newRemoteMemberUpdateLogger(update)),
		)
		return
	}

	if update.Version.Owner == r.localID {
		r.logger.Error(
			"remote update has same owner as local node",
			zap.Object("update", newRemoteMemberUpdateLogger(update)),
		)
		return
	}

	lostOwnership := false

	existing, ok := r.members[update.Member.Id]
	if ok {
		if existing.Version.Owner != update.Version.Owner {
			// Compare timestamps to choose a winner. Note ignore the counter
			// since the counter only applies locally.
			if existing.Version.Timestamp > update.Version.Timestamp {
				r.logger.Error(
					"discarding remote update; outdated version",
					zap.Object("update", newRemoteMemberUpdateLogger(update)),
					zap.Object("update-version", newVersionLogger(update.Version)),
					zap.Object("existing-version", newVersionLogger(existing.Version)),
				)
				return
			}
		}

		if existing.Version.Owner == r.localID {
			lostOwnership = true
		}
	}

	switch update.UpdateType {
	case rpc.MemberUpdateType_REGISTER:
		r.members[update.Member.Id] = &VersionedMember{
			Member:  update.Member,
			Version: update.Version,
		}

		r.logger.Info(
			"registered member; remote",
			zap.Object("update", newRemoteMemberUpdateLogger(update)),
		)
	case rpc.MemberUpdateType_UNREGISTER:
		delete(r.members, update.Member.Id)

		r.logger.Info(
			"unregistered member; remote",
			zap.Object("update", newRemoteMemberUpdateLogger(update)),
		)
	}

	// If we lost the ownership we still notify 'owner only' subscribers so they
	// know the new owner.
	r.notifySubscribersLocked(update, lostOwnership)
}

// Subscribe subscribes to updates to the registry.
//
// onUpdate will be called with the updates. It MUST NOT modify the update.
func (r *Registry) Subscribe(req *rpc.SubscribeRequest, onUpdate func(update *rpc.RemoteMemberUpdate), opts ...Option) func() {
	options := defaultRegistryOptions()
	for _, o := range opts {
		o.apply(options)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	handle := &subHandle{
		onUpdate:  onUpdate,
		ownerOnly: req.OwnerOnly,
	}
	r.subs[handle] = struct{}{}

	for _, u := range r.updatesLocked(req, options.now) {
		onUpdate(u)
	}

	return func() {
		r.mu.Lock()
		defer r.mu.Unlock()

		delete(r.subs, handle)
	}
}

func (r *Registry) MarkDownNodes(opts ...Option) {
	options := defaultRegistryOptions()
	for _, o := range opts {
		o.apply(options)
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	for _, m := range r.members {
		if m.Version.Owner != r.localID {
			continue
		}

		if m.Member.Status == rpc.MemberStatus_LEFT {
			if options.now-m.Version.Timestamp > options.tombstoneTimeout {
				r.logger.Info(
					"removing left member",
					zap.Int64("last-contact", options.now-m.Version.Timestamp),
				)
				m.Member.Status = rpc.MemberStatus_LEFT
				r.removeLocked(m.Member.Id, opts...)
			}
		}

		if m.Member.Status == rpc.MemberStatus_DOWN {
			if options.now-m.Version.Timestamp > options.reconnectTimeout {
				r.logger.Info(
					"unregistering down member",
					zap.Int64("down-since", options.now-m.Version.Timestamp),
				)
				m.Member.Status = rpc.MemberStatus_LEFT
				r.localRegisterLocked(m.Member, opts...)
			}
		}

		if m.Member.Status == rpc.MemberStatus_UP {
			if options.now-m.Version.Timestamp > options.heartbeatTimeout {
				r.logger.Info(
					"marking member down",
					zap.Int64("last-contact", options.now-m.Version.Timestamp),
				)
				m.Member.Status = rpc.MemberStatus_DOWN
				r.localRegisterLocked(m.Member, opts...)
			}
		}
	}
}

func (r *Registry) localRegisterLocked(member *rpc.Member, opts ...Option) {
	options := defaultRegistryOptions()
	for _, o := range opts {
		o.apply(options)
	}

	if member.Id == r.localID {
		r.logger.Error(
			"attempted to update local member",
			zap.Object("member", newMemberLogger(member)),
		)
		return
	}

	v := r.nextVersionLocked(options.now)

	existing, exists := r.members[member.Id]
	if exists {
		if existing.Version.Owner != v.Owner {
			// Compare timestamps to choose a winner. Note ignore the counter
			// since the counter only applies locally.
			//
			// This should never happen! It likely means there is clock skew
			// between nodes.
			//
			// If it does happen, the member will keep re-registering with every
			// heartbeat so should eventually have a late enough timestamp.
			if existing.Version.Timestamp > v.Timestamp {
				r.logger.Error(
					"discarding local register; outdated version",
					zap.Object("member", newMemberLogger(member)),
					zap.Object("update-version", newVersionLogger(v)),
					zap.Object("existing-version", newVersionLogger(existing.Version)),
				)
				return
			}
		}
	}

	versionedMember := &VersionedMember{
		Member:  member,
		Version: v,
	}
	r.members[member.Id] = versionedMember

	r.logger.Info(
		"registered member; owner",
		zap.Object("member", newMemberLogger(versionedMember.Member)),
		zap.Object("update-version", newVersionLogger(versionedMember.Version)),
	)

	r.notifySubscribersLocked(&rpc.RemoteMemberUpdate{
		UpdateType: rpc.MemberUpdateType_REGISTER,
		Member:     versionedMember.Member,
		Version:    versionedMember.Version,
	}, true)
}

func (r *Registry) removeLocked(id string, opts ...Option) {
	options := defaultRegistryOptions()
	for _, o := range opts {
		o.apply(options)
	}

	if id == r.localID {
		r.logger.Error(
			"attempted to remove local member",
			zap.String("id", id),
		)
		return
	}

	existing, exists := r.members[id]
	// If the member isn't registered do nothing.
	if !exists {
		return
	}

	delete(r.members, id)

	r.logger.Info(
		"removed member",
		zap.Object("member", newMemberLogger(existing.Member)),
	)
}

func (r *Registry) updatesLocked(req *rpc.SubscribeRequest, now int64) []*rpc.RemoteMemberUpdate {
	if req.KnownMembers == nil {
		req.KnownMembers = make(map[string]*rpc.Version)
	}

	var updates []*rpc.RemoteMemberUpdate

	for _, m := range r.members {
		knownVersion, ok := req.KnownMembers[m.Member.Id]
		if ok {
			if compareVersions(knownVersion, m.Version) > 0 {
				// If the subscriber thinks we own the node, but we don't, send
				// a REGISTER even if ownerOnly.
				if m.Version.Owner == r.localID || !req.OwnerOnly || knownVersion.Owner == r.localID {
					updates = append(updates, &rpc.RemoteMemberUpdate{
						UpdateType: rpc.MemberUpdateType_REGISTER,
						Member:     m.Member,
						Version:    m.Version,
					})
				}
			}
		} else {
			if m.Version.Owner == r.localID || !req.OwnerOnly {
				updates = append(updates, &rpc.RemoteMemberUpdate{
					UpdateType: rpc.MemberUpdateType_REGISTER,
					Member:     m.Member,
					Version:    m.Version,
				})
			}
		}
	}

	for id, knownVersion := range req.KnownMembers {
		if _, ok := r.members[id]; !ok {
			// Only return unregisters for nodes the subscriber thinks we
			// own.
			if req.OwnerOnly && knownVersion.Owner == r.localID {
				// TODO(AD) For now send a version of now, but need to improve
				// this to avoid unregistering a member that has actually
				// moved.
				updates = append(updates, &rpc.RemoteMemberUpdate{
					UpdateType: rpc.MemberUpdateType_UNREGISTER,
					Member: &rpc.Member{
						Id: id,
					},
					Version: r.nextVersionLocked(now),
				})
			} else if !req.OwnerOnly {
				updates = append(updates, &rpc.RemoteMemberUpdate{
					UpdateType: rpc.MemberUpdateType_UNREGISTER,
					Member: &rpc.Member{
						Id: id,
					},
					Version: r.nextVersionLocked(now),
				})
			}
		}
	}

	return updates
}

// notifySubscribersLocked sends the update to the subscribers. The update will
// only be sent to 'owner only' subscribers if the local node is the owner
// of the member being updated.
func (r *Registry) notifySubscribersLocked(update *rpc.RemoteMemberUpdate, owner bool) {
	for s := range r.subs {
		if owner {
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

func compareVersions(lhs *rpc.Version, rhs *rpc.Version) int {
	if lhs.Owner != rhs.Owner {
		if lhs.Timestamp < rhs.Timestamp {
			return 1
		}
		if lhs.Timestamp > rhs.Timestamp {
			return -1
		}
		return 0
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
