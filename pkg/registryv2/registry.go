// Copyright (C) 2023 Andrew Dunstall
//
// Fuddle is free software: you can redistribute it and/or modify
// it under the terms of the GNU General Public License as published by
// the Free Software Foundation, either version 3 of the License, or
// (at your option) any later version.
//
// Fuddle is distributed in the hope that it will be useful,
// but WITHOUT ANY WARRANTY; without even the implied warranty of
// MERCHANTABILITY or FITNESS FOR A PARTICULAR PURPOSE.  See the
// GNU General Public License for more details.
//
// You should have received a copy of the GNU General Public License
// along with this program.  If not, see <https://www.gnu.org/licenses/>.

package registry

import (
	"fmt"
	"sync"
	"time"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"go.uber.org/zap"
)

var (
	ErrAlreadyRegistered = fmt.Errorf("member already registered")
	ErrNotRegistered     = fmt.Errorf("member not registered")
	ErrInvalidUpdate     = fmt.Errorf("invalid update")
)

type subscriber struct {
	Callback func(update *rpc.MemberUpdate)
}

type Registry struct {
	members map[string]*rpc.Member

	// lastContact contains the time the member was last heard from, either
	// from updates or heartbeats.
	lastContact map[string]time.Time

	// downMembers contains the members that are down and the times they went down.
	downMembers map[string]time.Time

	// localID is the ID of the local member. This member cannot be updated or
	// marked as down.
	localID string

	subscribers map[*subscriber]interface{}

	// mu is a mutex protecting the fields above.
	mu sync.Mutex

	// heartbeatTimeout is the time a member can have no contact before it is
	// marked as down.
	heartbeatTimeout time.Duration

	// reconnectTimeout is the time a down member has to reconnect before it is
	// unregistered.
	reconnectTimeout time.Duration

	logger *zap.Logger
}

func NewRegistry(localMember *rpc.Member, opts ...Option) *Registry {
	options := options{
		heartbeatTimeout: time.Second * 30,
		reconnectTimeout: time.Minute * 30,
		logger:           zap.NewNop(),
	}
	for _, o := range opts {
		o.apply(&options)
	}

	members := make(map[string]*rpc.Member)

	// The local member is always healthy.
	localMember.Status = rpc.MemberStatus_UP
	localMember.Version = 1
	members[localMember.Id] = localMember

	return &Registry{
		members:          members,
		lastContact:      make(map[string]time.Time),
		downMembers:      make(map[string]time.Time),
		localID:          localMember.Id,
		subscribers:      make(map[*subscriber]interface{}),
		heartbeatTimeout: options.heartbeatTimeout,
		reconnectTimeout: options.reconnectTimeout,
		logger:           options.logger,
	}
}

func (r *Registry) Register(member *rpc.Member, opts ...Option) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	options := options{
		time: time.Now(),
	}
	for _, o := range opts {
		o.apply(&options)
	}

	if member.Id == "" || member.ClientId == "" || member.Metadata == nil {
		return ErrInvalidUpdate
	}

	version := uint64(1)
	// We support clients re-registering the same node again, though it must
	// have the same client ID as before.
	existing, alreadyExists := r.members[member.Id]
	if alreadyExists {
		if existing.ClientId != member.ClientId {
			return ErrAlreadyRegistered
		}

		version = existing.Version + 1
	}

	member = CopyMember(member)

	member.Status = rpc.MemberStatus_UP
	member.Version = version
	r.members[member.Id] = member

	r.lastContact[member.Id] = options.time

	updateType := rpc.MemberUpdateType_REGISTER
	if alreadyExists {
		updateType = rpc.MemberUpdateType_STATE
	}

	update := &rpc.MemberUpdate{
		Id:         member.Id,
		UpdateType: updateType,
		Member:     CopyMember(member),
	}
	// Note call subscribers with mutex locked to guarantee order.
	for sub := range r.subscribers {
		sub.Callback(update)
	}

	return nil
}

func (r *Registry) Unregister(id string) bool {
	r.mu.Lock()
	defer r.mu.Unlock()

	return r.unregisterLocked(id)
}

func (r *Registry) UpdateMemberMetadata(id string, metadata map[string]string) error {
	m, ok := r.members[id]
	if !ok {
		return ErrNotRegistered
	}
	if metadata == nil {
		return ErrInvalidUpdate
	}

	for k, v := range metadata {
		m.Metadata[k] = v
	}
	m.Version++

	r.notifyStatusUpdateLocked(m)

	return nil
}

func (r *Registry) Subscribe(versions map[string]uint64, cb func(update *rpc.MemberUpdate)) func() {
	r.mu.Lock()
	defer r.mu.Unlock()

	if versions == nil {
		versions = make(map[string]uint64)
	}

	// Send any missing members or out of date.
	for id, member := range r.members {
		version, ok := versions[id]

		if !ok {
			update := &rpc.MemberUpdate{
				Id:         member.Id,
				UpdateType: rpc.MemberUpdateType_REGISTER,
				// Copy as being passed out of the mutex.
				Member: CopyMember(member),
			}
			cb(update)
			continue
		}

		if version < member.Version {
			update := &rpc.MemberUpdate{
				Id:         member.Id,
				UpdateType: rpc.MemberUpdateType_STATE,
				// Copy as being passed out of the mutex.
				Member: CopyMember(member),
			}
			cb(update)
			continue
		}
	}

	// Send any unregisters.
	for id := range versions {
		if _, ok := r.members[id]; !ok {
			update := &rpc.MemberUpdate{
				Id:         id,
				UpdateType: rpc.MemberUpdateType_UNREGISTER,
			}
			cb(update)
		}
	}

	sub := &subscriber{
		Callback: cb,
	}
	r.subscribers[sub] = struct{}{}

	return func() {
		r.mu.Lock()
		defer r.mu.Unlock()

		delete(r.subscribers, sub)
	}
}

func (r *Registry) Heartbeat(clientID string, opts ...Option) {
	r.mu.Lock()
	defer r.mu.Unlock()

	options := options{
		time: time.Now(),
	}
	for _, o := range opts {
		o.apply(&options)
	}

	for id, member := range r.members {
		if member.ClientId != clientID {
			continue
		}

		r.lastContact[id] = options.time

		r.setStatusLocked(id, rpc.MemberStatus_UP, options.time)
	}
}

func (r *Registry) Member(id string) (*rpc.Member, bool) {
	r.mu.Lock()
	defer r.mu.Unlock()

	m, ok := r.members[id]
	if !ok {
		return nil, false
	}
	// Copy as being passed out of the mutex.
	return CopyMember(m), ok
}

func (r *Registry) Members() []*rpc.Member {
	r.mu.Lock()
	defer r.mu.Unlock()

	var members []*rpc.Member
	for _, m := range r.members {
		members = append(members, m)
	}
	return members
}

func (r *Registry) MarkFailedMembers(opts ...Option) {
	r.mu.Lock()
	defer r.mu.Unlock()

	options := options{
		time: time.Now(),
	}
	for _, o := range opts {
		o.apply(&options)
	}

	for id := range r.members {
		// Never mark the local member as failed.
		if id == r.localID {
			continue
		}

		lastContact := r.lastContact[id]
		if lastContact.Add(r.heartbeatTimeout).Before(options.time) {
			r.logger.Warn("member down", zap.String("id", id))

			r.setStatusLocked(id, rpc.MemberStatus_DOWN, options.time)
		}
	}
}

func (r *Registry) UnregisterFailedMembers(opts ...Option) {
	r.mu.Lock()
	defer r.mu.Unlock()

	options := options{
		time: time.Now(),
	}
	for _, o := range opts {
		o.apply(&options)
	}

	for id, downTime := range r.downMembers {
		if downTime.Add(r.reconnectTimeout).Before(options.time) {
			r.logger.Warn("down member unregistered", zap.String("id", id))
			r.unregisterLocked(id)
		}
	}
}

func (r *Registry) unregisterLocked(id string) bool {
	if _, ok := r.members[id]; !ok {
		return false
	}

	delete(r.members, id)
	delete(r.lastContact, id)
	delete(r.downMembers, id)

	update := &rpc.MemberUpdate{
		Id:         id,
		UpdateType: rpc.MemberUpdateType_UNREGISTER,
	}
	// Note call subscribers with mutex locked to guarantee order.
	for sub := range r.subscribers {
		sub.Callback(update)
	}

	return true
}

func (r *Registry) setStatusLocked(id string, status rpc.MemberStatus, t time.Time) bool {
	m := r.members[id]

	// If the status is unchanged do nothing.
	if m.Status == status {
		return false
	}

	m.Status = status
	m.Version++

	switch status {
	case rpc.MemberStatus_UP:
		delete(r.downMembers, id)
	case rpc.MemberStatus_DOWN:
		r.downMembers[id] = t
	}

	r.notifyStatusUpdateLocked(m)

	return true
}

func (r *Registry) notifyStatusUpdateLocked(member *rpc.Member) {
	update := &rpc.MemberUpdate{
		Id:         member.Id,
		UpdateType: rpc.MemberUpdateType_STATE,
		Member:     member,
	}
	// Note call subscribers with mutex locked to guarantee order.
	for sub := range r.subscribers {
		sub.Callback(update)
	}
}

func CopyMember(m *rpc.Member) *rpc.Member {
	metadata := make(map[string]string)
	for k, v := range m.Metadata {
		metadata[k] = v
	}
	return &rpc.Member{
		Id:       m.Id,
		ClientId: m.ClientId,
		Status:   m.Status,
		Version:  m.Version,
		Service:  m.Service,
		Locality: m.Locality,
		Created:  m.Created,
		Revision: m.Revision,
		Metadata: metadata,
	}
}
