package registry

import (
	"sync"
	"time"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
)

type subHandle struct {
	onUpdate func(update *rpc.RemoteMemberUpdate)
}

type Registry struct {
	localMember *rpc.VersionedMember

	subs map[*subHandle]interface{}

	// mu is a mutex protecting the above fields.
	mu sync.Mutex
}

func NewRegistry(opts ...Option) *Registry {
	options := defaultRegistryOptions()
	for _, o := range opts {
		o.apply(options)
	}

	var localMember *rpc.VersionedMember
	if options.localMember != nil {
		localMember = &rpc.VersionedMember{
			Member: options.localMember,
			Owner: &rpc.Owner{
				Owner:     options.localMember.Id,
				Timestamp: options.now.UnixMilli(),
			},
			Version: 1,
		}
	}

	return &Registry{
		localMember: localMember,
		subs:        make(map[*subHandle]interface{}),
	}
}

func (r *Registry) LocalUpdate(update *rpc.LocalMemberUpdate) {
	versionedMember := &rpc.VersionedMember{
		Member: update.Member,
		Owner: &rpc.Owner{
			Owner:     r.localMember.Member.Id,
			Timestamp: time.Now().UnixMilli(),
		},
		Version: 1,
	}
	for s := range r.subs {
		s.onUpdate(&rpc.RemoteMemberUpdate{
			Id:         versionedMember.Member.Id,
			UpdateType: rpc.MemberUpdateType_REGISTER,
			Member:     versionedMember,
		})
	}
}

func (r *Registry) RemoteUpdate(update *rpc.RemoteMemberUpdate) {
	for s := range r.subs {
		s.onUpdate(update)
	}
}

func (r *Registry) Subscribe(req *rpc.SubscribeRequest, onUpdate func(update *rpc.RemoteMemberUpdate)) func() {
	r.mu.Lock()
	defer r.mu.Unlock()

	handle := &subHandle{
		onUpdate: onUpdate,
	}
	r.subs[handle] = struct{}{}

	// TODO(AD) for now only sending local member regardless of subscribe
	// request.
	if r.localMember != nil {
		onUpdate(&rpc.RemoteMemberUpdate{
			Id:         r.localMember.Member.Id,
			UpdateType: rpc.MemberUpdateType_REGISTER,
			Member:     r.localMember,
		})
	}

	return func() {
		r.mu.Lock()
		defer r.mu.Unlock()

		delete(r.subs, handle)
	}
}
