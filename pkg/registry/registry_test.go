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
		WithLocalMember(localMember),
		WithLogger(testutils.Logger()),
	)
	m, ok := reg.MemberState("local")
	assert.True(t, ok)
	assert.True(t, proto.Equal(localMember, m))

	assert.Equal(t, 1.0, reg.Metrics().MembersCount.Value(map[string]string{
		"status": "up",
		"owner":  "local",
	}))
	assert.Equal(t, 1.0, reg.Metrics().MembersOwned.Value(map[string]string{
		"status": "up",
	}))
}

func TestRegistry_LocalMemberUpdateIgnored(t *testing.T) {
	localMember := randomMember("local")
	reg := NewRegistry(
		"local",
		WithLocalMember(localMember),
		WithLogger(testutils.Logger()),
	)

	updatedMember := randomMember("local")
	reg.AddMember(updatedMember)
	reg.RemoveMember("local")

	// The member should still equal the original member.
	m, ok := reg.MemberState("local")
	assert.True(t, ok)
	assert.True(t, proto.Equal(localMember, m))
}

func TestRegistry_AddMember(t *testing.T) {
	reg := NewRegistry(
		"local",
		WithLogger(testutils.Logger()),
	)

	addedMember := randomMember("my-member")
	reg.AddMember(addedMember)

	m, ok := reg.MemberState("my-member")
	assert.True(t, ok)
	assert.True(t, proto.Equal(addedMember, m))

	assert.Equal(t, 1.0, reg.Metrics().MembersCount.Value(map[string]string{
		"status": "up",
		"owner":  "local",
	}))
	assert.Equal(t, 1.0, reg.Metrics().MembersOwned.Value(map[string]string{
		"status": "up",
	}))
}

func TestRegistry_AddMemberDiscardOutdated(t *testing.T) {
	reg := NewRegistry(
		"local",
		WithLogger(testutils.Logger()),
	)

	addedMember := randomMember("my-member")
	reg.AddMember(addedMember, WithNowTime(100))

	reg.AddMember(randomMember("my-member"), WithNowTime(50))

	m, ok := reg.MemberState("my-member")
	assert.True(t, ok)
	assert.True(t, proto.Equal(addedMember, m))
}

func TestRegistry_SubscribeToAddMember(t *testing.T) {
	reg := NewRegistry(
		"local",
		WithLogger(testutils.Logger()),
	)

	var update *rpc.Member2
	reg.Subscribe(nil, func(u *rpc.Member2) {
		update = u
	})

	addedMember := randomMember("my-member")
	reg.AddMember(addedMember, WithNowTime(100))

	assert.True(t, proto.Equal(&rpc.Member2{
		State:    addedMember,
		Liveness: rpc.Liveness_UP,
		Version: &rpc.Version2{
			OwnerId: "local",
			Timestamp: &rpc.MonotonicTimestamp{
				Timestamp: 100,
				Counter:   0,
			},
		},
	}, update))
}

func TestRegistry_RemoveMember(t *testing.T) {
	reg := NewRegistry(
		"local",
		WithLogger(testutils.Logger()),
	)

	addedMember := randomMember("my-member")
	reg.AddMember(addedMember)
	reg.RemoveMember("my-member")

	m, ok := reg.Member("my-member")
	assert.True(t, ok)
	assert.True(t, proto.Equal(addedMember, m.State))
	assert.Equal(t, rpc.Liveness_LEFT, m.Liveness)

	assert.Equal(t, 1.0, reg.Metrics().MembersCount.Value(map[string]string{
		"status": "left",
		"owner":  "local",
	}))
	assert.Equal(t, 1.0, reg.Metrics().MembersOwned.Value(map[string]string{
		"status": "left",
	}))
}

func TestRegistry_RemoveMemberDiscardOutdated(t *testing.T) {
	reg := NewRegistry(
		"local",
		WithLogger(testutils.Logger()),
	)

	addedMember := randomMember("my-member")
	reg.AddMember(addedMember, WithNowTime(100))

	reg.RemoveMember("my-member", WithNowTime(50))

	m, ok := reg.MemberState("my-member")
	assert.True(t, ok)
	assert.True(t, proto.Equal(addedMember, m))
}

func TestRegistry_SubscribeToRemoveMember(t *testing.T) {
	reg := NewRegistry(
		"local",
		WithLogger(testutils.Logger()),
	)

	addedMember := randomMember("my-member")
	reg.AddMember(addedMember, WithNowTime(100))

	var update *rpc.Member2
	reg.Subscribe(nil, func(u *rpc.Member2) {
		update = u
	})

	reg.RemoveMember("my-member", WithNowTime(200))

	assert.True(t, proto.Equal(&rpc.Member2{
		State:    addedMember,
		Liveness: rpc.Liveness_LEFT,
		Version: &rpc.Version2{
			OwnerId: "local",
			Timestamp: &rpc.MonotonicTimestamp{
				Timestamp: 200,
				Counter:   0,
			},
		},
	}, update))
}

func TestRegistry_RemoteUpdateTakeOwnership(t *testing.T) {
	localMember := randomMember("local")
	reg := NewRegistry(
		"local",
		WithLocalMember(localMember),
		WithLogger(testutils.Logger()),
	)

	addedMember := randomMember("my-member")
	reg.AddMember(addedMember, WithNowTime(100))

	assert.Equal(t, 2.0, reg.Metrics().MembersCount.Value(map[string]string{
		"status": "up",
		"owner":  "local",
	}))
	assert.Equal(t, 2.0, reg.Metrics().MembersOwned.Value(map[string]string{
		"status": "up",
	}))

	updatedMember := randomMember("my-member")
	reg.RemoteUpdate(&rpc.Member2{
		State: updatedMember,
		Version: &rpc.Version2{
			OwnerId: "remote",
			Timestamp: &rpc.MonotonicTimestamp{
				// Outdated time.
				Timestamp: 200,
			},
		},
	})

	// The member should equal the remote update.
	m, ok := reg.MemberState("my-member")
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
	assert.Equal(t, 1.0, reg.Metrics().MembersOwned.Value(map[string]string{
		"status": "up",
	}))
}

func TestRegistry_RemoteUpdateWithOutdatedVersionIgnored(t *testing.T) {
	localMember := randomMember("local")
	reg := NewRegistry(
		"local",
		WithLocalMember(localMember),
		WithLogger(testutils.Logger()),
	)

	addedMember := randomMember("my-member")
	reg.AddMember(addedMember, WithNowTime(200))

	updatedMember := randomMember("my-member")
	reg.RemoteUpdate(&rpc.Member2{
		State: updatedMember,
		Version: &rpc.Version2{
			OwnerId: "remote",
			Timestamp: &rpc.MonotonicTimestamp{
				// Outdated time.
				Timestamp: 100,
			},
		},
	})

	// The member should still equal the original member.
	m, ok := reg.MemberState("my-member")
	assert.True(t, ok)
	assert.True(t, proto.Equal(addedMember, m))
}

func TestRegistry_LocalMemberRemoteUpdateIgnored(t *testing.T) {
	localMember := randomMember("local")
	reg := NewRegistry(
		"local",
		WithLocalMember(localMember),
		WithLogger(testutils.Logger()),
	)

	updatedMember := randomMember("local")
	reg.RemoteUpdate(&rpc.Member2{
		State: updatedMember,
		Version: &rpc.Version2{
			OwnerId: "remote",
			Timestamp: &rpc.MonotonicTimestamp{
				Timestamp: time.Now().UnixMilli() + 1000,
			},
		},
	})

	// The member should still equal the original member.
	m, ok := reg.MemberState("local")
	assert.True(t, ok)
	assert.True(t, proto.Equal(localMember, m))
}

func TestRegistry_RemoteUpdateWithLocalOwnerIgnored(t *testing.T) {
	localMember := randomMember("local")
	reg := NewRegistry(
		"local",
		WithLocalMember(localMember),
		WithLogger(testutils.Logger()),
	)

	addedMember := randomMember("my-member")
	reg.AddMember(addedMember)

	updatedMember := randomMember("my-member")
	reg.RemoteUpdate(&rpc.Member2{
		State: updatedMember,
		Version: &rpc.Version2{
			// Using the same owner as the local node should be ignored.
			OwnerId: "local",
			Timestamp: &rpc.MonotonicTimestamp{
				Timestamp: time.Now().UnixMilli() + 1000,
			},
		},
	})

	// The member should still equal the original member.
	m, ok := reg.MemberState("my-member")
	assert.True(t, ok)
	assert.True(t, proto.Equal(addedMember, m))
}

func TestRegistry_SubscribeToRemoteUpdate(t *testing.T) {
	localMember := randomMember("local")
	reg := NewRegistry(
		"local",
		WithLocalMember(localMember),
		WithLogger(testutils.Logger()),
	)

	var update *rpc.Member2
	reg.Subscribe(nil, func(u *rpc.Member2) {
		update = u
	})

	addedMember := randomMember("my-member")
	remoteUpdate := &rpc.Member2{
		State: addedMember,
		Version: &rpc.Version2{
			OwnerId: "remote",
			Timestamp: &rpc.MonotonicTimestamp{
				Timestamp: 100,
				Counter:   0,
			},
		},
	}
	reg.RemoteUpdate(remoteUpdate)

	assert.True(t, proto.Equal(remoteUpdate, update))
}

func TestRegistry_SubscribeOwnerOnlyIgnoresRemoteUpdate(t *testing.T) {
	reg := NewRegistry(
		"local",
		WithLogger(testutils.Logger()),
	)

	var update *rpc.Member2
	reg.Subscribe(
		&rpc.SubscribeRequest{OwnerOnly: true},
		func(u *rpc.Member2) {
			update = u
		},
	)

	addedMember := randomMember("my-member")
	remoteUpdate := &rpc.Member2{
		State: addedMember,
		Version: &rpc.Version2{
			OwnerId: "remote",
			Timestamp: &rpc.MonotonicTimestamp{
				Timestamp: 100,
				Counter:   0,
			},
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
		WithLocalMember(localMember),
		WithLogger(testutils.Logger()),
	)

	reg.AddMember(randomMember("my-member"), WithNowTime(50))

	var update *rpc.Member2
	reg.Subscribe(
		&rpc.SubscribeRequest{OwnerOnly: true},
		func(u *rpc.Member2) {
			update = u
		},
	)

	addedMember := randomMember("my-member")
	remoteUpdate := &rpc.Member2{
		State: addedMember,
		Version: &rpc.Version2{
			OwnerId: "remote",
			Timestamp: &rpc.MonotonicTimestamp{
				Timestamp: 100,
				Counter:   0,
			},
		},
	}
	reg.RemoteUpdate(remoteUpdate)

	assert.True(t, proto.Equal(remoteUpdate, update))
}

func TestRegistry_MemberNotFound(t *testing.T) {
	reg := NewRegistry(
		"local",
		WithLogger(testutils.Logger()),
	)
	_, ok := reg.Member("foo")
	assert.False(t, ok)
}

func TestRegistry_MarkMemberDownIgnoresLocal(t *testing.T) {
	localMember := randomMember("local")
	reg := NewRegistry(
		"local",
		WithLocalMember(localMember),
		WithLogger(testutils.Logger()),
		WithNowTime(100),
		WithHeartbeatTimeout(500),
	)

	reg.CheckMembersLiveness(
		WithNowTime(1000),
	)

	m, ok := reg.MemberState("local")
	assert.True(t, ok)
	assert.True(t, proto.Equal(localMember, m))
}

func TestRegistry_MarkMemberDownAfterMissingHeartbeats(t *testing.T) {
	reg := NewRegistry(
		"local",
		WithLogger(testutils.Logger()),
		WithHeartbeatTimeout(300),
	)

	addedMember := randomMember("my-member")
	reg.AddMember(addedMember, WithNowTime(100))

	reg.CheckMembersLiveness(
		WithNowTime(500),
	)

	m, ok := reg.Member("my-member")
	assert.True(t, ok)
	assert.True(t, proto.Equal(addedMember, m.State))
	assert.Equal(t, rpc.Liveness_DOWN, m.Liveness)

	assert.Equal(t, 0.0, reg.Metrics().MembersCount.Value(map[string]string{
		"status": "up",
		"owner":  "local",
	}))
	assert.Equal(t, 1.0, reg.Metrics().MembersCount.Value(map[string]string{
		"status": "down",
		"owner":  "local",
	}))
	assert.Equal(t, 1.0, reg.Metrics().MembersOwned.Value(map[string]string{
		"status": "down",
	}))

	// A heartbeat should revive it.
	reg.MemberHeartbeat(addedMember, WithNowTime(600))

	m, ok = reg.Member("my-member")
	assert.True(t, ok)
	assert.True(t, proto.Equal(addedMember, m.State))
	assert.Equal(t, rpc.Liveness_UP, m.Liveness)

	assert.Equal(t, 1.0, reg.Metrics().MembersCount.Value(map[string]string{
		"status": "up",
		"owner":  "local",
	}))
	assert.Equal(t, 0.0, reg.Metrics().MembersCount.Value(map[string]string{
		"status": "down",
		"owner":  "local",
	}))
	assert.Equal(t, 1.0, reg.Metrics().MembersOwned.Value(map[string]string{
		"status": "up",
	}))
}

func TestRegistry_MarkMemberRemovedAfterMissingHeartbeats(t *testing.T) {
	reg := NewRegistry(
		"local",
		WithLogger(testutils.Logger()),
		WithHeartbeatTimeout(300),
		WithReconnectTimeout(800),
	)

	addedMember := randomMember("my-member")
	reg.AddMember(addedMember, WithNowTime(100))

	reg.CheckMembersLiveness(
		WithNowTime(500),
	)

	m, ok := reg.Member("my-member")
	assert.True(t, ok)
	assert.True(t, proto.Equal(addedMember, m.State))
	assert.Equal(t, rpc.Liveness_DOWN, m.Liveness)

	reg.CheckMembersLiveness(
		WithNowTime(1500),
	)

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
	assert.True(t, proto.Equal(addedMember, m.State))
	assert.Equal(t, rpc.Liveness_LEFT, m.Liveness)

	// Adding the member again should revive it.
	reg.MemberHeartbeat(addedMember, WithNowTime(2000))

	m, ok = reg.Member("my-member")
	assert.True(t, ok)
	assert.True(t, proto.Equal(addedMember, m.State))
	assert.Equal(t, rpc.Liveness_UP, m.Liveness)

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
		WithLogger(testutils.Logger()),
		WithTombstoneTimeout(1000),
	)

	addedMember := randomMember("my-member")
	reg.AddMember(addedMember, WithNowTime(100))
	reg.RemoveMember("my-member", WithNowTime(200))

	reg.CheckMembersLiveness(
		WithNowTime(1500),
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
		WithLogger(testutils.Logger()),
		WithHeartbeatTimeout(500),
	)

	addedMember := randomMember("my-member")
	reg.RemoteUpdate(&rpc.Member2{
		State: addedMember,
		Version: &rpc.Version2{
			OwnerId: "remote",
			Timestamp: &rpc.MonotonicTimestamp{
				Timestamp: 100,
				Counter:   0,
			},
		},
	})

	reg.OnNodeLeave("remote", WithNowTime(200))

	reg.CheckMembersLiveness(
		WithNowTime(1000),
	)

	m, ok := reg.Member("my-member")
	assert.True(t, ok)
	assert.True(t, proto.Equal(addedMember, m.State))
	assert.Equal(t, rpc.Liveness_DOWN, m.Liveness)
}

func TestRegistry_UpdatesUnknownMembers(t *testing.T) {
	localMember := randomMember("local")
	reg := NewRegistry(
		"local",
		WithLocalMember(localMember),
		WithLogger(testutils.Logger()),
		WithNowTime(100),
	)

	updates := reg.Updates(&rpc.SubscribeRequest{
		KnownMembers: make(map[string]*rpc.Version2),
	})
	assert.Equal(t, 1, len(updates))
	assert.True(t, proto.Equal(updates[0], &rpc.Member2{
		State: localMember,
		Version: &rpc.Version2{
			OwnerId: "local",
			Timestamp: &rpc.MonotonicTimestamp{
				Timestamp: 100,
				Counter:   0,
			},
		},
	}))

	ownerOnlyUpdates := reg.Updates(&rpc.SubscribeRequest{
		KnownMembers: make(map[string]*rpc.Version2),
		OwnerOnly:    true,
	})
	assert.Equal(t, 1, len(ownerOnlyUpdates))
	assert.True(t, proto.Equal(ownerOnlyUpdates[0], &rpc.Member2{
		State: localMember,
		Version: &rpc.Version2{
			OwnerId: "local",
			Timestamp: &rpc.MonotonicTimestamp{
				Timestamp: 100,
				Counter:   0,
			},
		},
	}))
}

func TestRegistry_KnownMemberNotFound(t *testing.T) {
	reg := NewRegistry(
		"local",
		WithLogger(testutils.Logger()),
		WithNowTime(100),
	)

	updates := reg.Updates(&rpc.SubscribeRequest{
		KnownMembers: map[string]*rpc.Version2{
			"unknown-member": &rpc.Version2{
				OwnerId: "local",
				Timestamp: &rpc.MonotonicTimestamp{
					Timestamp: 50,
					Counter:   0,
				},
			},
		},
	})
	assert.Equal(t, 1, len(updates))
	assert.True(t, proto.Equal(updates[0], &rpc.Member2{
		State: &rpc.MemberState{
			Id: "unknown-member",
		},
		Liveness: rpc.Liveness_LEFT,
		Version: &rpc.Version2{
			OwnerId: "local",
			Timestamp: &rpc.MonotonicTimestamp{
				Timestamp: 50,
				Counter:   1,
			},
		},
	}))

	ownerOnlyUpdates := reg.Updates(&rpc.SubscribeRequest{
		KnownMembers: map[string]*rpc.Version2{
			"unknown-member": &rpc.Version2{
				OwnerId: "local",
				Timestamp: &rpc.MonotonicTimestamp{
					Timestamp: 50,
					Counter:   0,
				},
			},
		},
		OwnerOnly: true,
	})
	assert.Equal(t, 1, len(ownerOnlyUpdates))
	assert.True(t, proto.Equal(ownerOnlyUpdates[0], &rpc.Member2{
		State: &rpc.MemberState{
			Id: "unknown-member",
		},
		Liveness: rpc.Liveness_LEFT,
		Version: &rpc.Version2{
			OwnerId: "local",
			Timestamp: &rpc.MonotonicTimestamp{
				Timestamp: 50,
				Counter:   1,
			},
		},
	}))
}

func TestRegistry_UpdatesKnownMemberOutOfDate(t *testing.T) {
	localMember := randomMember("local")
	reg := NewRegistry(
		"local",
		WithLocalMember(localMember),
		WithLogger(testutils.Logger()),
		WithNowTime(100),
	)

	updates := reg.Updates(&rpc.SubscribeRequest{
		KnownMembers: map[string]*rpc.Version2{
			"local": &rpc.Version2{
				OwnerId: "local",
				Timestamp: &rpc.MonotonicTimestamp{
					Timestamp: 50,
					Counter:   0,
				},
			},
		},
	})
	assert.Equal(t, 1, len(updates))
	assert.True(t, proto.Equal(updates[0], &rpc.Member2{
		State: localMember,
		Version: &rpc.Version2{
			OwnerId: "local",
			Timestamp: &rpc.MonotonicTimestamp{
				Timestamp: 100,
				Counter:   0,
			},
		},
	}))

	ownerOnlyUpdates := reg.Updates(&rpc.SubscribeRequest{
		KnownMembers: map[string]*rpc.Version2{
			"local": &rpc.Version2{
				OwnerId: "local",
				Timestamp: &rpc.MonotonicTimestamp{
					Timestamp: 50,
					Counter:   0,
				},
			},
		},
		OwnerOnly: true,
	})
	assert.Equal(t, 1, len(ownerOnlyUpdates))
	assert.True(t, proto.Equal(ownerOnlyUpdates[0], &rpc.Member2{
		State: localMember,
		Version: &rpc.Version2{
			OwnerId: "local",
			Timestamp: &rpc.MonotonicTimestamp{
				Timestamp: 100,
				Counter:   0,
			},
		},
	}))
}

func randomMember(id string) *rpc.MemberState {
	if id == "" {
		id = uuid.New().String()
	}
	return &rpc.MemberState{
		Id:      id,
		Status:  uuid.New().String(),
		Service: uuid.New().String(),
		Locality: &rpc.Locality{
			Region:           uuid.New().String(),
			AvailabilityZone: uuid.New().String(),
		},
		Started:  rand.Int63(),
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
