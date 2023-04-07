package fuddle

import (
	"sync"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
)

type subscriber struct {
	Callback func()
}

type versionedMember struct {
	Member  *rpc.Member
	Version *rpc.Version
}

type registry struct {
	members map[string]*versionedMember

	// localMembers is a set containing the members registered by this client.
	localMembers map[string]interface{}

	subscribers map[*subscriber]interface{}

	// mu protects the above fields.
	mu sync.Mutex
}

func newRegistry() *registry {
	return &registry{
		members:      make(map[string]*versionedMember),
		localMembers: make(map[string]interface{}),
		subscribers:  make(map[*subscriber]interface{}),
	}
}

func (r *registry) RPCMember(id string) (*rpc.Member, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	m, ok := r.members[id]
	return m.Member, ok
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

func (r *registry) LocalMemberIDs() []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	var memberIDs []string
	for id := range r.localMembers {
		memberIDs = append(memberIDs, id)
	}
	return memberIDs
}

func (r *registry) LocalMembers() []Member {
	r.mu.Lock()
	defer r.mu.Unlock()

	var members []Member
	for id := range r.localMembers {
		members = append(members, fromRPC(r.members[id].Member))
	}
	return members
}

func (r *registry) KnownVersions() map[string]*rpc.Version {
	r.mu.Lock()
	defer r.mu.Unlock()

	versions := make(map[string]*rpc.Version)
	for id, m := range r.members {
		// Exclude local members from the known versions as the server doesn't
		// send us our own members.
		if _, ok := r.localMembers[id]; !ok {
			versions[id] = m.Version
		}
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

func (r *registry) RegisterLocal(member *rpc.Member) {
	r.mu.Lock()

	r.members[member.Id] = &versionedMember{
		Member: member,
	}
	r.localMembers[member.Id] = struct{}{}

	r.mu.Unlock()

	r.notifySubscribers()
}

func (r *registry) UnregisterLocal(id string) {
	r.mu.Lock()

	delete(r.members, id)
	delete(r.localMembers, id)

	r.mu.Unlock()

	r.notifySubscribers()
}

func (r *registry) UpdateMetadataLocal(id string, metadata map[string]string) {
	r.mu.Lock()

	member, ok := r.members[id]
	if !ok {
		r.mu.Unlock()
		return
	}

	for k, v := range metadata {
		member.Member.Metadata[k] = v
	}

	r.mu.Unlock()

	r.notifySubscribers()
}

func (r *registry) ApplyRemoteUpdate(update *rpc.RemoteMemberUpdate) {
	r.mu.Lock()

	// Ignore updates about local members.
	if _, ok := r.localMembers[update.Member.Id]; ok {
		r.mu.Unlock()
		return
	}

	switch update.UpdateType {
	case rpc.MemberUpdateType_REGISTER:
		r.applyRegisterUpdateLocked(update)
	case rpc.MemberUpdateType_UNREGISTER:
		r.applyUnregisterUpdateLocked(update)
	}

	r.mu.Unlock()

	r.notifySubscribers()
}

func (r *registry) applyRegisterUpdateLocked(update *rpc.RemoteMemberUpdate) {
	if update.Member.Status == rpc.MemberStatus_LEFT {
		delete(r.members, update.Member.Id)
	} else {
		r.members[update.Member.Id] = &versionedMember{
			Member:  update.Member,
			Version: update.Version,
		}
	}
}

func (r *registry) applyUnregisterUpdateLocked(update *rpc.RemoteMemberUpdate) {
	delete(r.members, update.Member.Id)
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
