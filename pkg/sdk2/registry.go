package fuddle

import (
	"sync"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"go.uber.org/zap"
	"go.uber.org/zap/zapcore"
)

type subscriber struct {
	Callback func()
}

type versionedMember struct {
	Member  *rpc.Member
	Version *rpc.Version
}

type registry struct {
	// members contains the members in the registry known by the client.
	members map[string]*versionedMember
	localID string

	subscribers map[*subscriber]interface{}

	// mu protects the above fields.
	mu sync.Mutex

	logger *zap.Logger
}

func newRegistry(member Member, logger *zap.Logger) *registry {
	members := make(map[string]*versionedMember)
	members[member.ID] = &versionedMember{
		Member: member.toRPC(),
	}

	return &registry{
		members:     members,
		localID:     member.ID,
		subscribers: make(map[*subscriber]interface{}),
		logger:      logger,
	}
}

func (r *registry) LocalRPCMember() *rpc.Member {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.members[r.localID].Member
}

func (r *registry) Members() []Member {
	r.mu.Lock()
	defer r.mu.Unlock()

	var members []Member
	for _, m := range r.members {
		members = append(members, fromRPC(m.Member))
	}
	return members
}

func (r *registry) KnownVersions() map[string]*rpc.Version {
	r.mu.Lock()
	defer r.mu.Unlock()

	versions := make(map[string]*rpc.Version)
	for id, m := range r.members {
		// Exclude the local member.
		if id == r.localID {
			continue
		}
		versions[id] = m.Version
	}
	return versions
}

func (r *registry) Subscribe(cb func()) func() {
	r.mu.Lock()

	sub := &subscriber{
		Callback: cb,
	}
	r.subscribers[sub] = struct{}{}

	r.mu.Unlock()

	// Ensure calling outside of the mutex.
	cb()

	return func() {
		r.mu.Lock()
		defer r.mu.Unlock()

		delete(r.subscribers, sub)
	}
}

func (r *registry) RemoteUpdate(update *rpc.RemoteMemberUpdate) {
	r.logger.Debug(
		"remote update",
		zap.Object("update", newRemoteMemberUpdateLogger(update)),
	)

	if update.Member.Id == r.localID {
		return
	}

	if update.Member.Status == rpc.MemberStatus_UP {
		r.updateMember(&versionedMember{
			Member:  update.Member,
			Version: update.Version,
		})
	} else {
		r.removeMember(update.Member.Id)
	}

	r.notifySubscribers()
}

func (r *registry) updateMember(m *versionedMember) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.members[m.Member.Id] = m
}

func (r *registry) removeMember(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	delete(r.members, id)
}

func (r *registry) notifySubscribers() {
	r.mu.Lock()

	// Copy the subscribers to avoid calling with the mutex locked.
	subscribers := make([]*subscriber, 0, len(r.subscribers))
	for sub := range r.subscribers {
		subscribers = append(subscribers, sub)
	}

	r.mu.Unlock()

	for _, sub := range subscribers {
		sub.Callback()
	}
}

type versionLogger struct {
	Owner     string
	Timestamp int64
	Counter   uint64
}

func newVersionLogger(v *rpc.Version) *versionLogger {
	return &versionLogger{
		Owner:     v.Owner,
		Timestamp: v.Timestamp,
		Counter:   v.Counter,
	}
}

func (l versionLogger) MarshalLogObject(e zapcore.ObjectEncoder) error {
	e.AddString("owner", l.Owner)
	e.AddInt64("timestamp", l.Timestamp)
	e.AddUint64("counter", l.Counter)
	return nil
}

type metadataLogger map[string]string

func (m metadataLogger) MarshalLogObject(e zapcore.ObjectEncoder) error {
	for k, v := range m {
		e.AddString(k, v)
	}
	return nil
}

type memberLogger struct {
	ID       string
	Status   string
	Service  string
	Locality string
	Created  int64
	Revision string
	Metadata metadataLogger
}

func newMemberLogger(m *rpc.Member) *memberLogger {
	return &memberLogger{
		ID:       m.Id,
		Status:   m.Status.String(),
		Service:  m.Service,
		Locality: m.Locality,
		Created:  m.Created,
		Revision: m.Revision,
		Metadata: m.Metadata,
	}
}

func (l memberLogger) MarshalLogObject(e zapcore.ObjectEncoder) error {
	e.AddString("id", l.ID)
	e.AddString("status", l.Status)
	e.AddString("service", l.Service)
	e.AddString("locality", l.Locality)
	e.AddInt64("created", l.Created)
	e.AddString("revision", l.Revision)
	if err := e.AddObject("metadata", metadataLogger(l.Metadata)); err != nil {
		return err
	}
	return nil
}

type remoteMemberUpdateLogger struct {
	UpdateType rpc.MemberUpdateType
	Member     *memberLogger
	Version    *versionLogger
}

func newRemoteMemberUpdateLogger(u *rpc.RemoteMemberUpdate) *remoteMemberUpdateLogger {
	return &remoteMemberUpdateLogger{
		UpdateType: u.UpdateType,
		Member:     newMemberLogger(u.Member),
		Version:    newVersionLogger(u.Version),
	}
}

func (l remoteMemberUpdateLogger) MarshalLogObject(e zapcore.ObjectEncoder) error {
	e.AddString("update-type", l.UpdateType.String())
	if err := e.AddObject("member", l.Member); err != nil {
		return err
	}
	if err := e.AddObject("version", l.Version); err != nil {
		return err
	}
	return nil
}
