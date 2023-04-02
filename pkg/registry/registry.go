package registry

import (
	"sync"

	"go.uber.org/zap"
)

type subHandle struct {
	cb        func(id string)
	ownerOnly bool
}

type Registry struct {
	members      map[string]interface{}
	ownedMembers map[string]interface{}
	localID      string

	subs map[*subHandle]interface{}

	// mu is a mutex protecting the above fields.
	mu sync.Mutex

	logger *zap.Logger
}

func NewRegistry(localID string, logger *zap.Logger) *Registry {
	members := make(map[string]interface{})
	ownedMembers := make(map[string]interface{})

	members[localID] = struct{}{}
	ownedMembers[localID] = struct{}{}

	return &Registry{
		members:      members,
		ownedMembers: ownedMembers,
		localID:      localID,
		subs:         make(map[*subHandle]interface{}),
		logger:       logger,
	}
}

func (r *Registry) RegisterLocal(id string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.members[id] = struct{}{}
	r.ownedMembers[id] = struct{}{}

	for s := range r.subs {
		s.cb(id)
	}
}

func (r *Registry) RegisterRemote(id string, owner string) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.members[id] = struct{}{}

	for s := range r.subs {
		if !s.ownerOnly {
			s.cb(id)
		}
	}
}

func (r *Registry) Members(ownerOnly bool) []string {
	r.mu.Lock()
	defer r.mu.Unlock()

	var members []string
	if ownerOnly {
		for id := range r.ownedMembers {
			members = append(members, id)
		}
	} else {
		for id := range r.members {
			members = append(members, id)
		}
	}
	return members
}

func (r *Registry) Subscribe(ownerOnly bool, cb func(id string)) func() {
	r.mu.Lock()
	defer r.mu.Unlock()

	handle := &subHandle{
		cb:        cb,
		ownerOnly: ownerOnly,
	}
	r.subs[handle] = struct{}{}

	if ownerOnly {
		for id := range r.ownedMembers {
			cb(id)
		}
	} else {
		for id := range r.members {
			cb(id)
		}
	}

	return func() {
		r.mu.Lock()
		defer r.mu.Unlock()

		delete(r.subs, handle)
	}
}
