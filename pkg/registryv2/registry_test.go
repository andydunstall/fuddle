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
	"testing"
	"time"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"github.com/fuddle-io/fuddle/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestRegistry_LookupLocalMember(t *testing.T) {
	localMember := testutils.RandomMember()
	r := NewRegistry(localMember)

	expectedMember := localMember
	expectedMember.Status = rpc.MemberStatus_UP
	expectedMember.Version = 1

	m, ok := r.Member(localMember.Id)
	assert.True(t, ok)
	assert.True(t, proto.Equal(expectedMember, m))
}

func TestRegistry_LookupMemberNotFound(t *testing.T) {
	r := NewRegistry(testutils.RandomMember())

	_, ok := r.Member("not-found")
	assert.False(t, ok)
}

func TestRegistry_Register(t *testing.T) {
	r := NewRegistry(testutils.RandomMember())

	member := testutils.RandomMember()
	assert.NoError(t, r.Register(member))

	expectedMember := member
	expectedMember.Status = rpc.MemberStatus_UP
	expectedMember.Version = 1

	m, ok := r.Member(member.Id)
	assert.True(t, ok)
	assert.True(t, proto.Equal(expectedMember, m))
}

func TestRegistry_RegisterInvalidMember(t *testing.T) {
	r := NewRegistry(testutils.RandomMember())

	memberMissingID := testutils.RandomMember()
	memberMissingID.Id = ""

	memberMissingClientID := testutils.RandomMember()
	memberMissingClientID.ClientId = ""

	memberMissingMetadata := testutils.RandomMember()
	memberMissingMetadata.Metadata = nil

	tests := []struct {
		Name   string
		Member *rpc.Member
	}{
		{
			Name:   "missing id",
			Member: memberMissingID,
		},
		{
			Name:   "missing client id",
			Member: memberMissingClientID,
		},
		{
			Name:   "missing metadata",
			Member: memberMissingMetadata,
		},
	}
	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			assert.Equal(t, ErrInvalidUpdate, r.Register(tt.Member))
		})
	}
}

func TestRegistry_RegisterAlreadyRegisteredWithDifferentClientID(t *testing.T) {
	r := NewRegistry(testutils.RandomMember())

	m1 := testutils.RandomMember()
	assert.NoError(t, r.Register(m1))

	m2 := CopyMember(m1)
	m2.ClientId = "no-match"

	assert.Equal(t, ErrAlreadyRegistered, r.Register(m2))
}

func TestRegistry_RegisterThenUnregister(t *testing.T) {
	r := NewRegistry(testutils.RandomMember())

	member := testutils.RandomMember()
	assert.NoError(t, r.Register(member))
	assert.True(t, r.Unregister(member.Id))

	_, ok := r.Member(member.Id)
	assert.False(t, ok)

}

func TestRegistry_UpdateMemberMetadata(t *testing.T) {
	r := NewRegistry(testutils.RandomMember())

	member := testutils.RandomMember()
	assert.NoError(t, r.Register(member))

	metadata := testutils.RandomMetadata()
	assert.NoError(t, r.UpdateMemberMetadata(member.Id, metadata))

	expectedMember := CopyMember(member)
	for k, v := range metadata {
		expectedMember.Metadata[k] = v
	}
	expectedMember.Version = 2

	m, ok := r.Member(member.Id)
	assert.True(t, ok)
	assert.True(t, proto.Equal(expectedMember, m))
}

func TestRegistry_UpdateMemberMetadataInvalid(t *testing.T) {
	r := NewRegistry(testutils.RandomMember())

	member := testutils.RandomMember()
	assert.NoError(t, r.Register(member))

	assert.Equal(
		t,
		ErrInvalidUpdate,
		r.UpdateMemberMetadata(member.Id, nil),
	)
}

func TestRegistry_UpdateMemberMetadataNotRegistered(t *testing.T) {
	r := NewRegistry(testutils.RandomMember())

	assert.Equal(
		t,
		ErrNotRegistered,
		r.UpdateMemberMetadata("not-found", make(map[string]string)),
	)
}

func TestRegistry_HeartbeatRevivesDownNode(t *testing.T) {
	r := NewRegistry(testutils.RandomMember(), WithHeartbeatTimeout(10*time.Second))

	member := testutils.RandomMember()
	assert.NoError(t, r.Register(member, WithTime(time.Unix(0, 0))))

	r.MarkFailedMembers(WithTime(time.Unix(20, 0)))

	m, ok := r.Member(member.Id)
	assert.True(t, ok)
	assert.Equal(t, rpc.MemberStatus_DOWN, m.Status)
	assert.Equal(t, uint64(2), m.Version)

	// A heartbeat before the reconnect timeout expires should mark the member
	// as up.
	r.Heartbeat(member.ClientId, WithTime(time.Unix(25, 0)))

	m, ok = r.Member(member.Id)
	assert.True(t, ok)
	assert.Equal(t, rpc.MemberStatus_UP, m.Status)
	// Updating the status should update the version.
	assert.Equal(t, uint64(3), m.Version)
}

func TestRegistry_RegisterRevivesDownNode(t *testing.T) {
	r := NewRegistry(testutils.RandomMember(), WithHeartbeatTimeout(10*time.Second))

	member := testutils.RandomMember()
	assert.NoError(t, r.Register(member, WithTime(time.Unix(0, 0))))

	r.MarkFailedMembers(WithTime(time.Unix(20, 0)))

	m, ok := r.Member(member.Id)
	assert.True(t, ok)
	assert.Equal(t, rpc.MemberStatus_DOWN, m.Status)
	assert.Equal(t, uint64(2), m.Version)

	// A re-register before the reconnect timeout expires should mark the member
	// as up.
	assert.NoError(t, r.Register(member, WithTime(time.Unix(25, 0))))

	m, ok = r.Member(member.Id)
	assert.True(t, ok)
	assert.Equal(t, rpc.MemberStatus_UP, m.Status)
	// Updating the status should update the version.
	assert.Equal(t, uint64(3), m.Version)
}

// Tests members that receive regular heartbeats are not marked as failed.
// Registers a member at t100, receives a heartbeat at t105, t110, and t115, then
// marks failed members at t120 with a t10 heartbeat timeout.
func TestRegistry_MarkFailedMembersIgnoresMembersWithHeartbeat(t *testing.T) {
	r := NewRegistry(testutils.RandomMember(), WithHeartbeatTimeout(10*time.Second))

	member := testutils.RandomMember()
	assert.NoError(t, r.Register(member, WithTime(time.Unix(100, 0))))

	r.Heartbeat(member.ClientId, WithTime(time.Unix(105, 0)))
	r.Heartbeat(member.ClientId, WithTime(time.Unix(110, 0)))
	r.Heartbeat(member.ClientId, WithTime(time.Unix(115, 0)))

	r.MarkFailedMembers(WithTime(time.Unix(120, 0)))

	m, ok := r.Member(member.Id)
	assert.True(t, ok)
	assert.Equal(t, rpc.MemberStatus_UP, m.Status)
	// The version should not have updated.
	assert.Equal(t, uint64(1), m.Version)
}

// Tests members that never receive a heartbeat are marked as failed.
// Registers a member at t0, then marks failed members at t20 with a t10 heartbeat
// timeout.
func TestRegistry_MarkFailedMembersFailsMembersWithNoHeartbeats(t *testing.T) {
	r := NewRegistry(testutils.RandomMember(), WithHeartbeatTimeout(10*time.Second))

	member := testutils.RandomMember()
	assert.NoError(t, r.Register(member, WithTime(time.Unix(0, 0))))

	// Add heartbeat for another client ID.
	r.Heartbeat("foo", WithTime(time.Unix(19, 0)))

	r.MarkFailedMembers(WithTime(time.Unix(20, 0)))

	m, ok := r.Member(member.Id)
	assert.True(t, ok)
	assert.Equal(t, rpc.MemberStatus_DOWN, m.Status)
	assert.Equal(t, uint64(2), m.Version)
}

// Tests MarkFailedMembers doesn't mark members that have just registered but not
// yet sent a heartbeat as down.
// Registers a member at t100, then marks failed members at t105 with a t10 heartbeat
// timeout.
func TestRegistry_MarkFailedMembersIgnoresJustRegistered(t *testing.T) {
	r := NewRegistry(testutils.RandomMember(), WithHeartbeatTimeout(10*time.Second))

	member := testutils.RandomMember()
	assert.NoError(t, r.Register(member, WithTime(time.Unix(100, 0))))

	r.MarkFailedMembers(WithTime(time.Unix(105, 0)))

	m, ok := r.Member(member.Id)
	assert.True(t, ok)
	assert.Equal(t, rpc.MemberStatus_UP, m.Status)
	assert.Equal(t, uint64(1), m.Version)
}

func TestRegistry_MarkFailedMembersIgnoresLocalNode(t *testing.T) {
	localMember := testutils.RandomMember()
	r := NewRegistry(localMember)

	r.MarkFailedMembers(WithTime(time.Unix(100, 0)))

	m, ok := r.Member(localMember.Id)
	assert.True(t, ok)
	assert.Equal(t, rpc.MemberStatus_UP, m.Status)
	assert.Equal(t, uint64(1), m.Version)
}

func TestRegistry_UnregisterFailedMembers(t *testing.T) {
	r := NewRegistry(
		testutils.RandomMember(),
		WithHeartbeatTimeout(10*time.Second),
		WithReconnectTimeout(100*time.Second),
	)

	member := testutils.RandomMember()
	assert.NoError(t, r.Register(member, WithTime(time.Unix(0, 0))))

	r.MarkFailedMembers(WithTime(time.Unix(20, 0)))

	m, ok := r.Member(member.Id)
	assert.True(t, ok)
	assert.Equal(t, rpc.MemberStatus_DOWN, m.Status)

	// Unregistering failed members BEFORE the reconnect timeout should do
	// nothing.
	r.UnregisterFailedMembers(WithTime(time.Unix(50, 0)))

	// The member should still be registered.
	_, ok = r.Member(member.Id)
	assert.True(t, ok)

	// Marking failed members again should not change the down time.
	r.MarkFailedMembers(WithTime(time.Unix(100, 0)))

	// Unregistering failed members AFTER the reconnect timeout unregister the
	// member.
	r.UnregisterFailedMembers(WithTime(time.Unix(150, 0)))

	_, ok = r.Member(member.Id)
	assert.False(t, ok)
}

func TestRegistry_UnregisterFailedMembersIgnoresRevivedNode(t *testing.T) {
	r := NewRegistry(
		testutils.RandomMember(),
		WithHeartbeatTimeout(10*time.Second),
		WithReconnectTimeout(100*time.Second),
	)

	member := testutils.RandomMember()
	assert.NoError(t, r.Register(member, WithTime(time.Unix(0, 0))))

	r.MarkFailedMembers(WithTime(time.Unix(20, 0)))

	m, ok := r.Member(member.Id)
	assert.True(t, ok)
	assert.Equal(t, rpc.MemberStatus_DOWN, m.Status)

	// Revive the member.
	r.Heartbeat(member.ClientId, WithTime(time.Unix(145, 0)))

	r.UnregisterFailedMembers(WithTime(time.Unix(150, 0)))

	// The member should still be registered.
	_, ok = r.Member(member.Id)
	assert.True(t, ok)
}

func TestRegistry_SubscribeBootstrapReceiveRegisters(t *testing.T) {
	localMember := testutils.RandomMember()
	r := NewRegistry(localMember)

	m1 := testutils.RandomMember()
	assert.NoError(t, r.Register(m1))
	m2 := testutils.RandomMember()
	assert.NoError(t, r.Register(m2))
	m3 := testutils.RandomMember()
	assert.NoError(t, r.Register(m3))

	receivedRegisterUpdates := make(map[string]interface{})

	// Subscribe knowing about the local member and m1, so expect to receive
	// m2 and m3.
	r.Subscribe(map[string]uint64{
		localMember.Id: 1,
		m1.Id:          1,
	}, func(update *rpc.MemberUpdate) {
		assert.Equal(t, rpc.MemberUpdateType_REGISTER, update.UpdateType)
		receivedRegisterUpdates[update.Id] = struct{}{}
	})

	expectedIDs := map[string]interface{}{
		m2.Id: struct{}{},
		m3.Id: struct{}{},
	}
	assert.Equal(t, expectedIDs, receivedRegisterUpdates)
}

func TestRegistry_SubscribeBootstrapReceiveStateUpdates(t *testing.T) {
	localMember := testutils.RandomMember()
	r := NewRegistry(localMember)

	m1 := testutils.RandomMember()
	assert.NoError(t, r.Register(m1))
	assert.NoError(t, r.UpdateMemberMetadata(m1.Id, testutils.RandomMetadata()))
	assert.NoError(t, r.UpdateMemberMetadata(m1.Id, testutils.RandomMetadata()))
	// Hacky just to update status.
	r.setStatusLocked(m1.Id, rpc.MemberStatus_DOWN, time.Now())

	m2 := testutils.RandomMember()
	assert.NoError(t, r.Register(m2))
	assert.NoError(t, r.UpdateMemberMetadata(m2.Id, testutils.RandomMetadata()))
	assert.NoError(t, r.UpdateMemberMetadata(m2.Id, testutils.RandomMetadata()))
	// Hacky just to update status.
	r.setStatusLocked(m2.Id, rpc.MemberStatus_DOWN, time.Now())
	r.setStatusLocked(m2.Id, rpc.MemberStatus_UP, time.Now())

	receivedStateUpdates := make(map[string]uint64)

	// Subscribe knowing about the local member and m1, so expect to receive
	// m2 and m3.
	r.Subscribe(map[string]uint64{
		localMember.Id: 1,
		// Missing updates for members.
		m1.Id: 1,
		m2.Id: 2,
	}, func(update *rpc.MemberUpdate) {
		assert.Equal(t, rpc.MemberUpdateType_STATE, update.UpdateType)
		receivedStateUpdates[update.Id] = update.Member.Version
	})

	expectedIDs := map[string]uint64{
		m1.Id: 4,
		m2.Id: 5,
	}
	assert.Equal(t, expectedIDs, receivedStateUpdates)
}

func TestRegistry_SubscribeBootstrapReceiveUnregisters(t *testing.T) {
	localMember := testutils.RandomMember()
	r := NewRegistry(localMember)

	m1 := testutils.RandomMember()
	assert.NoError(t, r.Register(m1))

	receivedUnregisterUpdates := make(map[string]interface{})

	// Subscribe knowing about the local member and m1, so expect to receive
	// m2 and m3.
	r.Subscribe(map[string]uint64{
		localMember.Id: 1,
		m1.Id:          1,
		// Member not in the registry.
		"member-1": 5,
		"member-2": 7,
		"member-3": 2,
	}, func(update *rpc.MemberUpdate) {
		assert.Equal(t, rpc.MemberUpdateType_UNREGISTER, update.UpdateType)
		receivedUnregisterUpdates[update.Id] = struct{}{}
	})

	expectedIDs := map[string]interface{}{
		"member-1": struct{}{},
		"member-2": struct{}{},
		"member-3": struct{}{},
	}
	assert.Equal(t, expectedIDs, receivedUnregisterUpdates)
}

func TestRegistry_SubscribeUpdatesReceiveRegisters(t *testing.T) {
	localMember := testutils.RandomMember()
	r := NewRegistry(localMember)

	var receivedMember *rpc.Member
	r.Subscribe(map[string]uint64{localMember.Id: 1}, func(update *rpc.MemberUpdate) {
		assert.Equal(t, rpc.MemberUpdateType_REGISTER, update.UpdateType)
		receivedMember = update.Member
	})

	member := testutils.RandomMember()
	assert.NoError(t, r.Register(member))

	expectedMember := member
	expectedMember.Version = 1

	assert.Equal(t, expectedMember, receivedMember)
}

func TestRegistry_SubscribeUpdatesReceiveStateUpdates(t *testing.T) {
	localMember := testutils.RandomMember()
	r := NewRegistry(localMember)

	member := testutils.RandomMember()
	assert.NoError(t, r.Register(member))

	var receivedMember *rpc.Member
	r.Subscribe(map[string]uint64{
		localMember.Id: 1,
		member.Id:      1,
	}, func(update *rpc.MemberUpdate) {
		assert.Equal(t, rpc.MemberUpdateType_STATE, update.UpdateType)
		receivedMember = update.Member
	})

	expectedMember := member
	expectedMember.Status = rpc.MemberStatus_DOWN
	expectedMember.Version = 2

	r.setStatusLocked(member.Id, rpc.MemberStatus_DOWN, time.Now())

	assert.Equal(t, expectedMember, receivedMember)
}

func TestRegistry_SubscribeUpdatesReceiveReRegisterUpdates(t *testing.T) {
	localMember := testutils.RandomMember()
	r := NewRegistry(localMember)

	member := testutils.RandomMember()
	assert.NoError(t, r.Register(member))

	r.setStatusLocked(member.Id, rpc.MemberStatus_DOWN, time.Now())

	var receivedMember *rpc.Member
	r.Subscribe(map[string]uint64{
		localMember.Id: 1,
		member.Id:      1,
	}, func(update *rpc.MemberUpdate) {
		assert.Equal(t, rpc.MemberUpdateType_STATE, update.UpdateType)
		receivedMember = update.Member
	})

	// Register the member again to revive it.
	assert.NoError(t, r.Register(member))

	expectedMember := member
	expectedMember.Status = rpc.MemberStatus_UP
	expectedMember.Version = 3

	assert.Equal(t, expectedMember, receivedMember)
}

func TestRegistry_SubscribeUpdatesReceiveUnregisters(t *testing.T) {
	localMember := testutils.RandomMember()
	r := NewRegistry(localMember)

	member := testutils.RandomMember()
	assert.NoError(t, r.Register(member))

	var receivedMember string
	r.Subscribe(map[string]uint64{
		localMember.Id: 1,
		member.Id:      1,
	}, func(update *rpc.MemberUpdate) {
		assert.Equal(t, rpc.MemberUpdateType_UNREGISTER, update.UpdateType)
		receivedMember = update.Id
	})

	r.Unregister(member.Id)

	assert.Equal(t, member.Id, receivedMember)
}

func TestRegistry_Unsubscribe(t *testing.T) {
	localMember := testutils.RandomMember()
	r := NewRegistry(localMember)

	unsubscribe := r.Subscribe(map[string]uint64{localMember.Id: 1}, func(update *rpc.MemberUpdate) {
		t.Error("unexpected notification")
	})
	unsubscribe()

	assert.NoError(t, r.Register(testutils.RandomMember()))
}

func TestRegistry_Members(t *testing.T) {
	localMember := testutils.RandomMember()
	r := NewRegistry(localMember)

	expectedMember := localMember
	expectedMember.Status = rpc.MemberStatus_UP
	expectedMember.Version = 1

	members := r.Members()
	assert.Equal(t, []*rpc.Member{expectedMember}, members)
}
