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
	// members contains the members in the registry known by the client.
	members map[string]*versionedMember

	subscribers map[*subscriber]interface{}

	// mu protects the above fields.
	mu sync.Mutex
}

func newRegistry() *registry {
	return &registry{
		members:     make(map[string]*versionedMember),
		subscribers: make(map[*subscriber]interface{}),
	}
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
