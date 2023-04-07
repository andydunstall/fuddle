package registry

import (
	"testing"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"github.com/fuddle-io/fuddle/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestRegistry_GetLocalMember(t *testing.T) {
	r := NewRegistry("local-member", WithRegistryLocalMember(&rpc.Member{
		Id: "local-member",
	}), WithRegistryNowTime(100), WithRegistryLogger(testutils.Logger()))

	expectedLocalMember := &VersionedMember{
		Member: &rpc.Member{
			Id: "local-member",
		},
		Version: &rpc.Version{
			Owner:     "local-member",
			Timestamp: 100,
			Counter:   0,
		},
	}
	member, ok := r.Member("local-member")
	assert.True(t, ok)
	assert.True(t, expectedLocalMember.Equal(member))
}

func TestRegistry_RegisterOwnedMember(t *testing.T) {
	r := NewRegistry("local", WithRegistryLogger(testutils.Logger()))

	var recvUpdate *rpc.RemoteMemberUpdate
	r.Subscribe(&rpc.SubscribeRequest{
		OwnerOnly: true,
	}, func(update *rpc.RemoteMemberUpdate) {
		recvUpdate = update
	})

	r.LocalUpdate(&rpc.LocalMemberUpdate{
		UpdateType: rpc.MemberUpdateType_REGISTER,
		Member: &rpc.Member{
			Id: "member-1",
		},
	}, WithRegistryNowTime(100))

	expectedVersionedMember := &VersionedMember{
		Member: &rpc.Member{
			Id: "member-1",
		},
		Version: &rpc.Version{
			Owner:     "local",
			Timestamp: 100,
			Counter:   0,
		},
	}
	member, ok := r.Member("member-1")
	assert.True(t, ok)
	assert.True(t, expectedVersionedMember.Equal(member))

	expectedUpdate := &rpc.RemoteMemberUpdate{
		UpdateType: rpc.MemberUpdateType_REGISTER,
		Member:     expectedVersionedMember.Member,
		Version:    expectedVersionedMember.Version,
	}
	assert.True(t, proto.Equal(expectedUpdate, recvUpdate))
}

func TestRegistry_UnregisterOwnedMember(t *testing.T) {
	r := NewRegistry("local", WithRegistryLogger(testutils.Logger()))

	r.LocalUpdate(&rpc.LocalMemberUpdate{
		UpdateType: rpc.MemberUpdateType_REGISTER,
		Member: &rpc.Member{
			Id: "member-1",
		},
	}, WithRegistryNowTime(100))

	var recvUpdate *rpc.RemoteMemberUpdate
	r.Subscribe(&rpc.SubscribeRequest{
		OwnerOnly: true,
	}, func(update *rpc.RemoteMemberUpdate) {
		recvUpdate = update
	})

	r.LocalUpdate(&rpc.LocalMemberUpdate{
		UpdateType: rpc.MemberUpdateType_UNREGISTER,
		Member: &rpc.Member{
			Id: "member-1",
		},
	}, WithRegistryNowTime(200))

	_, ok := r.Member("member-1")
	assert.False(t, ok)

	expectedUpdate := &rpc.RemoteMemberUpdate{
		UpdateType: rpc.MemberUpdateType_UNREGISTER,
		Member: &rpc.Member{
			Id: "member-1",
		},
		Version: &rpc.Version{
			Owner:     "local",
			Timestamp: 200,
			Counter:   0,
		},
	}
	assert.True(t, proto.Equal(expectedUpdate, recvUpdate))
}

func TestRegistry_UpdateOwnedMember(t *testing.T) {
	r := NewRegistry("local", WithRegistryLogger(testutils.Logger()))

	r.LocalUpdate(&rpc.LocalMemberUpdate{
		UpdateType: rpc.MemberUpdateType_REGISTER,
		Member: &rpc.Member{
			Id: "member-1",
		},
	}, WithRegistryNowTime(100))

	var recvUpdate *rpc.RemoteMemberUpdate
	r.Subscribe(&rpc.SubscribeRequest{
		OwnerOnly: true,
	}, func(update *rpc.RemoteMemberUpdate) {
		recvUpdate = update
	})

	// Update the member by adding metadata.
	r.LocalUpdate(&rpc.LocalMemberUpdate{
		UpdateType: rpc.MemberUpdateType_REGISTER,
		Member: &rpc.Member{
			Id: "member-1",
			Metadata: map[string]string{
				"foo": "bar",
			},
		},
	}, WithRegistryNowTime(100))

	expectedVersionedMember := &VersionedMember{
		Member: &rpc.Member{
			Id: "member-1",
			Metadata: map[string]string{
				"foo": "bar",
			},
		},
		Version: &rpc.Version{
			Owner:     "local",
			Timestamp: 100,
			// Since the update had the same timestamp, the counter should be
			// incremented.
			Counter: 1,
		},
	}
	member, ok := r.Member("member-1")
	assert.True(t, ok)
	assert.True(t, expectedVersionedMember.Equal(member))

	expectedUpdate := &rpc.RemoteMemberUpdate{
		UpdateType: rpc.MemberUpdateType_REGISTER,
		Member:     expectedVersionedMember.Member,
		Version:    expectedVersionedMember.Version,
	}
	assert.True(t, proto.Equal(expectedUpdate, recvUpdate))
}

func TestRegistry_RegisterRemoteMember(t *testing.T) {
	r := NewRegistry("local", WithRegistryLogger(testutils.Logger()))

	var recvUpdate *rpc.RemoteMemberUpdate
	r.Subscribe(&rpc.SubscribeRequest{
		OwnerOnly: false,
	}, func(update *rpc.RemoteMemberUpdate) {
		recvUpdate = update
	})

	versionedMember := &VersionedMember{
		Member: &rpc.Member{
			Id: "member-1",
		},
		Version: &rpc.Version{
			Owner:     "remote",
			Timestamp: 100,
			Counter:   0,
		},
	}
	r.RemoteUpdate(&rpc.RemoteMemberUpdate{
		UpdateType: rpc.MemberUpdateType_REGISTER,
		Member:     versionedMember.Member,
		Version:    versionedMember.Version,
	})

	member, ok := r.Member("member-1")
	assert.True(t, ok)
	assert.True(t, versionedMember.Equal(member))

	expectedUpdate := &rpc.RemoteMemberUpdate{
		UpdateType: rpc.MemberUpdateType_REGISTER,
		Member: &rpc.Member{
			Id: "member-1",
		},
		Version: &rpc.Version{
			Owner:     "remote",
			Timestamp: 100,
			Counter:   0,
		},
	}
	assert.True(t, proto.Equal(expectedUpdate, recvUpdate))
}

func TestRegistry_UnregisterRemoteMember(t *testing.T) {
	r := NewRegistry("local", WithRegistryLogger(testutils.Logger()))

	versionedMember := &VersionedMember{
		Member: &rpc.Member{
			Id: "member-1",
		},
		Version: &rpc.Version{
			Owner:     "remote",
			Timestamp: 100,
			Counter:   0,
		},
	}
	r.RemoteUpdate(&rpc.RemoteMemberUpdate{
		UpdateType: rpc.MemberUpdateType_REGISTER,
		Member:     versionedMember.Member,
		Version:    versionedMember.Version,
	})

	var recvUpdate *rpc.RemoteMemberUpdate
	r.Subscribe(&rpc.SubscribeRequest{
		OwnerOnly: false,
	}, func(update *rpc.RemoteMemberUpdate) {
		recvUpdate = update
	})

	r.RemoteUpdate(&rpc.RemoteMemberUpdate{
		UpdateType: rpc.MemberUpdateType_UNREGISTER,
		Member: &rpc.Member{
			Id: "member-1",
		},
		Version: &rpc.Version{
			Owner:     "remote",
			Timestamp: 200,
			Counter:   0,
		},
	})

	_, ok := r.Member("member-1")
	assert.False(t, ok)

	expectedUpdate := &rpc.RemoteMemberUpdate{
		UpdateType: rpc.MemberUpdateType_UNREGISTER,
		Member: &rpc.Member{
			Id: "member-1",
		},
		Version: &rpc.Version{
			Owner:     "remote",
			Timestamp: 200,
			Counter:   0,
		},
	}
	assert.True(t, proto.Equal(expectedUpdate, recvUpdate))
}

func TestRegistry_DiscardOutOfDateRemoteRegister(t *testing.T) {
	r := NewRegistry("local", WithRegistryLogger(testutils.Logger()))

	// Add an update with owner remote-1 and timestamp 200.
	versionedMember1 := &VersionedMember{
		Member: &rpc.Member{
			Id: "member-1",
		},
		Version: &rpc.Version{
			Owner:     "remote-1",
			Timestamp: 200,
			Counter:   0,
		},
	}
	r.RemoteUpdate(&rpc.RemoteMemberUpdate{
		UpdateType: rpc.MemberUpdateType_REGISTER,
		Member:     versionedMember1.Member,
		Version:    versionedMember1.Version,
	})

	// Add another update with owner remote-2 and timestamp 100, which should
	// be discarded.
	versionedMember2 := &VersionedMember{
		Member: &rpc.Member{
			Id: "member-1",
		},
		Version: &rpc.Version{
			Owner:     "remote-2",
			Timestamp: 100,
			Counter:   0,
		},
	}
	r.RemoteUpdate(&rpc.RemoteMemberUpdate{
		UpdateType: rpc.MemberUpdateType_REGISTER,
		Member:     versionedMember2.Member,
		Version:    versionedMember2.Version,
	})

	member, ok := r.Member("member-1")
	assert.True(t, ok)
	// The original versioned member should be retained.
	assert.True(t, versionedMember1.Equal(member))
}

func TestRegistry_DiscardLocalUpdateToLocalMember(t *testing.T) {
	r := NewRegistry("local-member", WithRegistryLocalMember(&rpc.Member{
		Id: "local-member",
	}), WithRegistryNowTime(100), WithRegistryLogger(testutils.Logger()))

	r.LocalUpdate(&rpc.LocalMemberUpdate{
		UpdateType: rpc.MemberUpdateType_REGISTER,
		Member: &rpc.Member{
			Id: "local-member",
			Metadata: map[string]string{
				"foo": "bar",
			},
		},
	}, WithRegistryNowTime(200))

	expectedLocalMember := &VersionedMember{
		Member: &rpc.Member{
			Id: "local-member",
		},
		Version: &rpc.Version{
			Owner:     "local-member",
			Timestamp: 100,
			Counter:   0,
		},
	}
	member, ok := r.Member("local-member")
	assert.True(t, ok)
	assert.True(t, expectedLocalMember.Equal(member))
}

func TestRegistry_DiscardOutOfDateLocalUpdate(t *testing.T) {
	r := NewRegistry("local", WithRegistryLogger(testutils.Logger()))

	versionedMember := &VersionedMember{
		Member: &rpc.Member{
			Id: "member-1",
		},
		Version: &rpc.Version{
			Owner:     "remote",
			Timestamp: 200,
			Counter:   0,
		},
	}
	r.RemoteUpdate(&rpc.RemoteMemberUpdate{
		UpdateType: rpc.MemberUpdateType_REGISTER,
		Member:     versionedMember.Member,
		Version:    versionedMember.Version,
	})

	r.LocalUpdate(&rpc.LocalMemberUpdate{
		UpdateType: rpc.MemberUpdateType_REGISTER,
		Member: &rpc.Member{
			Id: "member-1",
			Metadata: map[string]string{
				"foo": "bar",
			},
		},
	}, WithRegistryNowTime(100))

	member, ok := r.Member("member-1")
	assert.True(t, ok)
	assert.True(t, versionedMember.Equal(member))
}

func TestRegistry_DiscardLocalUnregisterIfNotOwner(t *testing.T) {
	r := NewRegistry("local", WithRegistryLogger(testutils.Logger()))

	versionedMember := &VersionedMember{
		Member: &rpc.Member{
			Id: "member-1",
		},
		Version: &rpc.Version{
			Owner:     "remote",
			Timestamp: 100,
			Counter:   0,
		},
	}
	r.RemoteUpdate(&rpc.RemoteMemberUpdate{
		UpdateType: rpc.MemberUpdateType_REGISTER,
		Member:     versionedMember.Member,
		Version:    versionedMember.Version,
	})

	r.LocalUpdate(&rpc.LocalMemberUpdate{
		UpdateType: rpc.MemberUpdateType_UNREGISTER,
		Member: &rpc.Member{
			Id: "member-1",
		},
	}, WithRegistryNowTime(200))

	member, ok := r.Member("member-1")
	assert.True(t, ok)
	assert.True(t, versionedMember.Equal(member))
}
