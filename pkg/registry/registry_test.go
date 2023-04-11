package registry

import (
	"math/rand"
	"testing"
	"time"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"github.com/fuddle-io/fuddle/pkg/testutils"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestRegistry_AddLocalMember(t *testing.T) {
	localMember := randomMember("local")
	reg := NewRegistry(
		"local",
		WithRegistryLocalMember(localMember),
		WithRegistryLogger(testutils.Logger()),
	)
	m, ok := reg.Member("local")
	assert.True(t, ok)
	assert.True(t, proto.Equal(localMember, m))

	assert.Equal(t, 1.0, reg.Metrics().MembersCount.Value(map[string]string{
		"status": "up",
		"owner":  "local",
	}))
}

func TestRegistry_LocalMemberUpdateIgnored(t *testing.T) {
	localMember := randomMember("local")
	reg := NewRegistry(
		"local",
		WithRegistryLocalMember(localMember),
		WithRegistryLogger(testutils.Logger()),
	)

	updatedMember := randomMember("local")
	reg.AddMember(updatedMember)
	reg.RemoveMember("local")

	// The member should still equal the original member.
	m, ok := reg.Member("local")
	assert.True(t, ok)
	assert.True(t, proto.Equal(localMember, m))
}

func TestRegistry_AddMember(t *testing.T) {
	reg := NewRegistry(
		"local",
		WithRegistryLogger(testutils.Logger()),
	)

	addedMember := randomMember("my-member")
	reg.AddMember(addedMember)

	m, ok := reg.Member("my-member")
	assert.True(t, ok)
	assert.True(t, proto.Equal(addedMember, m))

	assert.Equal(t, 1.0, reg.Metrics().MembersCount.Value(map[string]string{
		"status": "up",
		"owner":  "local",
	}))
}

func TestRegistry_AddMemberDiscardOutdated(t *testing.T) {
	reg := NewRegistry(
		"local",
		WithRegistryLogger(testutils.Logger()),
	)

	addedMember := randomMember("my-member")
	reg.AddMember(addedMember, WithRegistryNowTime(100))

	reg.AddMember(randomMember("my-member"), WithRegistryNowTime(50))

	m, ok := reg.Member("my-member")
	assert.True(t, ok)
	assert.True(t, proto.Equal(addedMember, m))
}

func TestRegistry_SubscribeToAddMember(t *testing.T) {
	reg := NewRegistry(
		"local",
		WithRegistryLogger(testutils.Logger()),
	)

	var update *rpc.RemoteMemberUpdate
	reg.Subscribe(nil, func(u *rpc.RemoteMemberUpdate) {
		update = u
	})

	addedMember := randomMember("my-member")
	reg.AddMember(addedMember, WithRegistryNowTime(100))

	assert.True(t, proto.Equal(&rpc.RemoteMemberUpdate{
		Member: addedMember,
		Version: &rpc.Version{
			Owner:     "local",
			Timestamp: 100,
			Counter:   0,
		},
	}, update))
}

func TestRegistry_RemoveMember(t *testing.T) {
	reg := NewRegistry(
		"local",
		WithRegistryLogger(testutils.Logger()),
	)

	addedMember := randomMember("my-member")
	reg.AddMember(addedMember)
	reg.RemoveMember("my-member")

	expectedMember := memberWithStatus(addedMember, rpc.MemberStatus_LEFT)

	m, ok := reg.Member("my-member")
	assert.True(t, ok)
	assert.True(t, proto.Equal(expectedMember, m))

	assert.Equal(t, 1.0, reg.Metrics().MembersCount.Value(map[string]string{
		"status": "left",
		"owner":  "local",
	}))
}

func TestRegistry_RemoveMemberDiscardOutdated(t *testing.T) {
	reg := NewRegistry(
		"local",
		WithRegistryLogger(testutils.Logger()),
	)

	addedMember := randomMember("my-member")
	reg.AddMember(addedMember, WithRegistryNowTime(100))

	reg.RemoveMember("my-member", WithRegistryNowTime(50))

	m, ok := reg.Member("my-member")
	assert.True(t, ok)
	assert.True(t, proto.Equal(addedMember, m))
}

func TestRegistry_SubscribeToRemoveMember(t *testing.T) {
	reg := NewRegistry(
		"local",
		WithRegistryLogger(testutils.Logger()),
	)

	addedMember := randomMember("my-member")
	reg.AddMember(addedMember, WithRegistryNowTime(100))

	var update *rpc.RemoteMemberUpdate
	reg.Subscribe(nil, func(u *rpc.RemoteMemberUpdate) {
		update = u
	})

	reg.RemoveMember("my-member", WithRegistryNowTime(200))

	expectedMember := memberWithStatus(addedMember, rpc.MemberStatus_LEFT)

	assert.True(t, proto.Equal(&rpc.RemoteMemberUpdate{
		Member: expectedMember,
		Version: &rpc.Version{
			Owner:     "local",
			Timestamp: 200,
			Counter:   0,
		},
	}, update))
}

func TestRegistry_RemoteUpdateTakeOwnership(t *testing.T) {
	localMember := randomMember("local")
	reg := NewRegistry(
		"local",
		WithRegistryLocalMember(localMember),
		WithRegistryLogger(testutils.Logger()),
	)

	addedMember := randomMember("my-member")
	reg.AddMember(addedMember, WithRegistryNowTime(100))

	assert.Equal(t, 2.0, reg.Metrics().MembersCount.Value(map[string]string{
		"status": "up",
		"owner":  "local",
	}))

	updatedMember := randomMember("my-member")
	reg.RemoteUpdate(&rpc.RemoteMemberUpdate{
		Member: updatedMember,
		Version: &rpc.Version{
			Owner: "remote",
			// Outdated time.
			Timestamp: 200,
		},
	})

	// The member should equal the remote update.
	m, ok := reg.Member("my-member")
	assert.True(t, ok)
	assert.True(t, proto.Equal(updatedMember, m))

	assert.Equal(t, 1.0, reg.Metrics().MembersCount.Value(map[string]string{
		"status": "up",
		"owner":  "local",
	}))
	assert.Equal(t, 1.0, reg.Metrics().MembersCount.Value(map[string]string{
		"status": "up",
		"owner":  "remote",
	}))
}

func TestRegistry_RemoteUpdateWithOutdatedVersionIgnored(t *testing.T) {
	localMember := randomMember("local")
	reg := NewRegistry(
		"local",
		WithRegistryLocalMember(localMember),
		WithRegistryLogger(testutils.Logger()),
	)

	addedMember := randomMember("my-member")
	reg.AddMember(addedMember, WithRegistryNowTime(200))

	updatedMember := randomMember("my-member")
	reg.RemoteUpdate(&rpc.RemoteMemberUpdate{
		Member: updatedMember,
		Version: &rpc.Version{
			Owner: "remote",
			// Outdated time.
			Timestamp: 100,
		},
	})

	// The member should still equal the original member.
	m, ok := reg.Member("my-member")
	assert.True(t, ok)
	assert.True(t, proto.Equal(addedMember, m))
}

func TestRegistry_LocalMemberRemoteUpdateIgnored(t *testing.T) {
	localMember := randomMember("local")
	reg := NewRegistry(
		"local",
		WithRegistryLocalMember(localMember),
		WithRegistryLogger(testutils.Logger()),
	)

	updatedMember := randomMember("local")
	reg.RemoteUpdate(&rpc.RemoteMemberUpdate{
		Member: updatedMember,
		Version: &rpc.Version{
			Owner:     "remote",
			Timestamp: time.Now().UnixMilli() + 1000,
		},
	})

	// The member should still equal the original member.
	m, ok := reg.Member("local")
	assert.True(t, ok)
	assert.True(t, proto.Equal(localMember, m))
}

func TestRegistry_RemoteUpdateWithLocalOwnerIgnored(t *testing.T) {
	localMember := randomMember("local")
	reg := NewRegistry(
		"local",
		WithRegistryLocalMember(localMember),
		WithRegistryLogger(testutils.Logger()),
	)

	addedMember := randomMember("my-member")
	reg.AddMember(addedMember)

	updatedMember := randomMember("my-member")
	reg.RemoteUpdate(&rpc.RemoteMemberUpdate{
		Member: updatedMember,
		Version: &rpc.Version{
			// Using the same owner as the local node should be ignored.
			Owner:     "local",
			Timestamp: time.Now().UnixMilli() + 1000,
		},
	})

	// The member should still equal the original member.
	m, ok := reg.Member("my-member")
	assert.True(t, ok)
	assert.True(t, proto.Equal(addedMember, m))
}

func TestRegistry_SubscribeToRemoteUpdate(t *testing.T) {
	localMember := randomMember("local")
	reg := NewRegistry(
		"local",
		WithRegistryLocalMember(localMember),
		WithRegistryLogger(testutils.Logger()),
	)

	var update *rpc.RemoteMemberUpdate
	reg.Subscribe(nil, func(u *rpc.RemoteMemberUpdate) {
		update = u
	})

	addedMember := randomMember("my-member")
	remoteUpdate := &rpc.RemoteMemberUpdate{
		Member: addedMember,
		Version: &rpc.Version{
			Owner:     "remote",
			Timestamp: 100,
		},
	}
	reg.RemoteUpdate(remoteUpdate)

	assert.True(t, proto.Equal(remoteUpdate, update))
}

func TestRegistry_SubscribeOwnerOnlyIgnoresRemoteUpdate(t *testing.T) {
	reg := NewRegistry(
		"local",
		WithRegistryLogger(testutils.Logger()),
	)

	var update *rpc.RemoteMemberUpdate
	reg.Subscribe(
		&rpc.SubscribeRequest{OwnerOnly: true},
		func(u *rpc.RemoteMemberUpdate) {
			update = u
		},
	)

	addedMember := randomMember("my-member")
	remoteUpdate := &rpc.RemoteMemberUpdate{
		Member: addedMember,
		Version: &rpc.Version{
			Owner:     "remote",
			Timestamp: 100,
		},
	}
	reg.RemoteUpdate(remoteUpdate)

	assert.Nil(t, update)
}

// Tests when the local node loses ownership of a member, it notifies owner-only
// subscribers about the update.
func TestRegistry_SubscribeOwnerOnlyReceivesOwnershipChangesUpdates(t *testing.T) {
	localMember := randomMember("local")
	reg := NewRegistry(
		"local",
		WithRegistryLocalMember(localMember),
		WithRegistryLogger(testutils.Logger()),
	)

	reg.AddMember(randomMember("my-member"), WithRegistryNowTime(50))

	var update *rpc.RemoteMemberUpdate
	reg.Subscribe(
		&rpc.SubscribeRequest{OwnerOnly: true},
		func(u *rpc.RemoteMemberUpdate) {
			update = u
		},
	)

	addedMember := randomMember("my-member")
	remoteUpdate := &rpc.RemoteMemberUpdate{
		Member: addedMember,
		Version: &rpc.Version{
			Owner:     "remote",
			Timestamp: 100,
		},
	}
	reg.RemoteUpdate(remoteUpdate)

	assert.True(t, proto.Equal(remoteUpdate, update))
}

func TestRegistry_MemberNotFound(t *testing.T) {
	reg := NewRegistry(
		"local",
		WithRegistryLogger(testutils.Logger()),
	)
	_, ok := reg.Member("foo")
	assert.False(t, ok)
}

func TestRegistry_MarkMemberDownIgnoresLocal(t *testing.T) {
	localMember := randomMember("local")
	reg := NewRegistry(
		"local",
		WithRegistryLocalMember(localMember),
		WithRegistryLogger(testutils.Logger()),
		WithRegistryNowTime(100),
		WithHeartbeatTimeout(500),
	)

	expectedMember := copyMember(localMember)

	reg.CheckMembersLiveness(
		WithRegistryNowTime(1000),
	)

	m, ok := reg.Member("local")
	assert.True(t, ok)
	assert.True(t, proto.Equal(expectedMember, m))
}

func TestRegistry_MarkMemberDownAfterMissingHeartbeats(t *testing.T) {
	reg := NewRegistry(
		"local",
		WithRegistryLogger(testutils.Logger()),
		WithHeartbeatTimeout(300),
	)

	addedMember := randomMember("my-member")
	reg.AddMember(addedMember, WithRegistryNowTime(100))

	reg.CheckMembersLiveness(
		WithRegistryNowTime(500),
	)

	expectedMember := memberWithStatus(addedMember, rpc.MemberStatus_DOWN)

	m, ok := reg.Member("my-member")
	assert.True(t, ok)
	assert.True(t, proto.Equal(expectedMember, m))

	assert.Equal(t, 0.0, reg.Metrics().MembersCount.Value(map[string]string{
		"status": "up",
		"owner":  "local",
	}))
	assert.Equal(t, 1.0, reg.Metrics().MembersCount.Value(map[string]string{
		"status": "down",
		"owner":  "local",
	}))

	// Adding the member again should revive it.
	reg.AddMember(addedMember, WithRegistryNowTime(600))

	m, ok = reg.Member("my-member")
	assert.True(t, ok)
	assert.True(t, proto.Equal(addedMember, m))

	assert.Equal(t, 1.0, reg.Metrics().MembersCount.Value(map[string]string{
		"status": "up",
		"owner":  "local",
	}))
	assert.Equal(t, 0.0, reg.Metrics().MembersCount.Value(map[string]string{
		"status": "down",
		"owner":  "local",
	}))
}

func TestRegistry_MarkMemberRemovedAfterMissingHeartbeats(t *testing.T) {
	reg := NewRegistry(
		"local",
		WithRegistryLogger(testutils.Logger()),
		WithHeartbeatTimeout(300),
		WithReconnectTimeout(800),
	)

	addedMember := randomMember("my-member")
	reg.AddMember(addedMember, WithRegistryNowTime(100))

	reg.CheckMembersLiveness(
		WithRegistryNowTime(500),
	)

	expectedMember := memberWithStatus(addedMember, rpc.MemberStatus_DOWN)

	m, ok := reg.Member("my-member")
	assert.True(t, ok)
	assert.True(t, proto.Equal(expectedMember, m))

	reg.CheckMembersLiveness(
		WithRegistryNowTime(1500),
	)

	expectedMember = memberWithStatus(addedMember, rpc.MemberStatus_LEFT)

	assert.Equal(t, 0.0, reg.Metrics().MembersCount.Value(map[string]string{
		"status": "up",
		"owner":  "local",
	}))
	assert.Equal(t, 1.0, reg.Metrics().MembersCount.Value(map[string]string{
		"status": "left",
		"owner":  "local",
	}))

	m, ok = reg.Member("my-member")
	assert.True(t, ok)
	assert.True(t, proto.Equal(expectedMember, m))

	// Adding the member again should revive it.
	reg.AddMember(addedMember, WithRegistryNowTime(2000))

	m, ok = reg.Member("my-member")
	assert.True(t, ok)
	assert.True(t, proto.Equal(addedMember, m))

	assert.Equal(t, 1.0, reg.Metrics().MembersCount.Value(map[string]string{
		"status": "up",
		"owner":  "local",
	}))
	assert.Equal(t, 0.0, reg.Metrics().MembersCount.Value(map[string]string{
		"status": "left",
		"owner":  "local",
	}))
}

func TestRegistry_MarkMemberLeftMemberRemovedAfterTombstoneTimeout(t *testing.T) {
	reg := NewRegistry(
		"local",
		WithRegistryLogger(testutils.Logger()),
		WithTombstoneTimeout(1000),
	)

	addedMember := randomMember("my-member")
	reg.AddMember(addedMember, WithRegistryNowTime(100))
	reg.RemoveMember("my-member", WithRegistryNowTime(200))

	reg.CheckMembersLiveness(
		WithRegistryNowTime(1500),
	)

	_, ok := reg.Member("my-member")
	assert.False(t, ok)

	assert.Equal(t, 0.0, reg.Metrics().MembersCount.Value(map[string]string{
		"status": "up",
		"owner":  "local",
	}))
	assert.Equal(t, 0.0, reg.Metrics().MembersCount.Value(map[string]string{
		"status": "down",
		"owner":  "local",
	}))
	assert.Equal(t, 0.0, reg.Metrics().MembersCount.Value(map[string]string{
		"status": "left",
		"owner":  "local",
	}))
}

func TestRegistry_RemoveOwnedNodesMembers(t *testing.T) {
	reg := NewRegistry(
		"local",
		WithRegistryLogger(testutils.Logger()),
		WithHeartbeatTimeout(500),
	)

	addedMember := randomMember("my-member")
	reg.RemoteUpdate(&rpc.RemoteMemberUpdate{
		Member: addedMember,
		Version: &rpc.Version{
			Owner:     "remote",
			Timestamp: 100,
		},
	})

	reg.OnNodeLeave("remote", WithRegistryNowTime(200))

	reg.CheckMembersLiveness(
		WithRegistryNowTime(1000),
	)

	expectedMember := memberWithStatus(addedMember, rpc.MemberStatus_DOWN)

	member, ok := reg.Member("my-member")
	assert.True(t, ok)
	assert.True(t, proto.Equal(expectedMember, member))
}

func TestRegistry_UpdatesUnknownMembers(t *testing.T) {
	localMember := randomMember("local")
	reg := NewRegistry(
		"local",
		WithRegistryLocalMember(localMember),
		WithRegistryLogger(testutils.Logger()),
		WithRegistryNowTime(100),
	)

	updates := reg.Updates(&rpc.SubscribeRequest{
		KnownMembers: make(map[string]*rpc.Version),
	})
	assert.Equal(t, 1, len(updates))
	assert.True(t, proto.Equal(updates[0], &rpc.RemoteMemberUpdate{
		Member: localMember,
		Version: &rpc.Version{
			Owner:     "local",
			Timestamp: 100,
			Counter:   0,
		},
	}))

	ownerOnlyUpdates := reg.Updates(&rpc.SubscribeRequest{
		KnownMembers: make(map[string]*rpc.Version),
		OwnerOnly:    true,
	})
	assert.Equal(t, 1, len(ownerOnlyUpdates))
	assert.True(t, proto.Equal(ownerOnlyUpdates[0], &rpc.RemoteMemberUpdate{
		Member: localMember,
		Version: &rpc.Version{
			Owner:     "local",
			Timestamp: 100,
			Counter:   0,
		},
	}))
}

func TestRegistry_KnownMemberNotFound(t *testing.T) {
	reg := NewRegistry(
		"local",
		WithRegistryLogger(testutils.Logger()),
		WithRegistryNowTime(100),
	)

	updates := reg.Updates(&rpc.SubscribeRequest{
		KnownMembers: map[string]*rpc.Version{
			"unknown-member": &rpc.Version{
				Owner:     "local",
				Timestamp: 50,
			},
		},
	})
	assert.Equal(t, 1, len(updates))
	assert.True(t, proto.Equal(updates[0], &rpc.RemoteMemberUpdate{
		Member: &rpc.Member{
			Id:     "unknown-member",
			Status: rpc.MemberStatus_LEFT,
		},
		Version: &rpc.Version{
			Owner:     "local",
			Timestamp: 50,
			Counter:   1,
		},
	}))

	ownerOnlyUpdates := reg.Updates(&rpc.SubscribeRequest{
		KnownMembers: map[string]*rpc.Version{
			"unknown-member": &rpc.Version{
				Owner:     "local",
				Timestamp: 50,
			},
		},
		OwnerOnly: true,
	})
	assert.Equal(t, 1, len(ownerOnlyUpdates))
	assert.True(t, proto.Equal(ownerOnlyUpdates[0], &rpc.RemoteMemberUpdate{
		Member: &rpc.Member{
			Id:     "unknown-member",
			Status: rpc.MemberStatus_LEFT,
		},
		Version: &rpc.Version{
			Owner:     "local",
			Timestamp: 50,
			Counter:   1,
		},
	}))
}

func TestRegistry_UpdatesKnownMemberOutOfDate(t *testing.T) {
	localMember := randomMember("local")
	reg := NewRegistry(
		"local",
		WithRegistryLocalMember(localMember),
		WithRegistryLogger(testutils.Logger()),
		WithRegistryNowTime(100),
	)

	updates := reg.Updates(&rpc.SubscribeRequest{
		KnownMembers: map[string]*rpc.Version{
			"local": &rpc.Version{
				Owner:     "local",
				Timestamp: 50,
			},
		},
	})
	assert.Equal(t, 1, len(updates))
	assert.True(t, proto.Equal(updates[0], &rpc.RemoteMemberUpdate{
		Member: localMember,
		Version: &rpc.Version{
			Owner:     "local",
			Timestamp: 100,
			Counter:   0,
		},
	}))

	ownerOnlyUpdates := reg.Updates(&rpc.SubscribeRequest{
		KnownMembers: map[string]*rpc.Version{
			"local": &rpc.Version{
				Owner:     "local",
				Timestamp: 50,
			},
		},
		OwnerOnly: true,
	})
	assert.Equal(t, 1, len(ownerOnlyUpdates))
	assert.True(t, proto.Equal(ownerOnlyUpdates[0], &rpc.RemoteMemberUpdate{
		Member: localMember,
		Version: &rpc.Version{
			Owner:     "local",
			Timestamp: 100,
			Counter:   0,
		},
	}))
}

func TestCompareVersions(t *testing.T) {
	tests := []struct {
		Name     string
		LHS      *rpc.Version
		RHS      *rpc.Version
		Expected int
	}{
		{
			Name: "owners not equal but timestamps equal",
			LHS: &rpc.Version{
				Owner: "foo",
			},
			RHS: &rpc.Version{
				Owner: "bar",
			},
			// LHS greater even though timestamps equal.
			Expected: -1,
		},
		{
			Name: "owners not equal lhs timestamp greater",
			LHS: &rpc.Version{
				Owner:     "foo",
				Timestamp: 10,
			},
			RHS: &rpc.Version{
				Owner:     "bar",
				Timestamp: 5,
			},
			Expected: -1,
		},
		{
			Name: "owners not equal rhs timestamp greater",
			LHS: &rpc.Version{
				Owner:     "foo",
				Timestamp: 5,
			},
			RHS: &rpc.Version{
				Owner:     "bar",
				Timestamp: 10,
			},
			Expected: 1,
		},
		{
			Name: "owners equal lhs timestamp greater",
			LHS: &rpc.Version{
				Owner:     "foo",
				Timestamp: 10,
			},
			RHS: &rpc.Version{
				Owner:     "foo",
				Timestamp: 5,
			},
			Expected: -1,
		},
		{
			Name: "owners equal rhs timestamp greater",
			LHS: &rpc.Version{
				Owner:     "foo",
				Timestamp: 5,
			},
			RHS: &rpc.Version{
				Owner:     "foo",
				Timestamp: 10,
			},
			Expected: 1,
		},
		{
			Name: "owners equal lhs counter greater",
			LHS: &rpc.Version{
				Owner:     "foo",
				Timestamp: 10,
				Counter:   10,
			},
			RHS: &rpc.Version{
				Owner:     "foo",
				Timestamp: 10,
				Counter:   5,
			},
			Expected: -1,
		},
		{
			Name: "owners equal rhs counter greater",
			LHS: &rpc.Version{
				Owner:     "foo",
				Timestamp: 10,
				Counter:   5,
			},
			RHS: &rpc.Version{
				Owner:     "foo",
				Timestamp: 10,
				Counter:   10,
			},
			Expected: 1,
		},
		{
			Name: "equal",
			LHS: &rpc.Version{
				Owner:     "foo",
				Timestamp: 10,
				Counter:   10,
			},
			RHS: &rpc.Version{
				Owner:     "foo",
				Timestamp: 10,
				Counter:   10,
			},
			Expected: 0,
		},
	}

	for _, tt := range tests {
		t.Run(tt.Name, func(t *testing.T) {
			assert.Equal(t, tt.Expected, compareVersions(tt.LHS, tt.RHS))
		})
	}
}

func randomMember(id string) *rpc.Member {
	if id == "" {
		id = uuid.New().String()
	}
	return &rpc.Member{
		Id:       id,
		Service:  uuid.New().String(),
		Locality: uuid.New().String(),
		Created:  rand.Int63(),
		Revision: uuid.New().String(),
		Metadata: map[string]string{
			uuid.New().String(): uuid.New().String(),
			uuid.New().String(): uuid.New().String(),
			uuid.New().String(): uuid.New().String(),
			uuid.New().String(): uuid.New().String(),
			uuid.New().String(): uuid.New().String(),
		},
	}
}
