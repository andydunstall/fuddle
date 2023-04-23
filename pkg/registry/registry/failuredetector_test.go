package registry

import (
	"testing"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"github.com/fuddle-io/fuddle/pkg/testutils"
	"github.com/stretchr/testify/assert"
)

// Tests an owned member that missed its heartbeat timeout is marked down.
func TestFailureDetector_UnreachableMemberMarkedAsDown(t *testing.T) {
	registry := NewRegistry(
		"local",
		WithHeartbeatTimeout(500),
		WithReconnectTimeout(5000),
	)

	addedMember := testutils.RandomMemberState("my-member", "")
	registry.AddMember(addedMember, WithNowTime(100))

	registry.UpdateLiveness(1000)

	m, ok := registry.Member("my-member")
	assert.True(t, ok)
	assert.Equal(t, rpc.Liveness_DOWN, m.Liveness)
	assert.Equal(t, int64(6000), m.Expiry)
}

// Tests an owned member that hasn't recovered in the reconnect timeout is
// marked as left.
func TestFailureDetector_ExpiredDownMemberMarkedAsLeft(t *testing.T) {
	registry := NewRegistry(
		"local",
		WithHeartbeatTimeout(500),
		WithReconnectTimeout(5000),
		WithTombstoneTimeout(50000),
	)

	addedMember := testutils.RandomMemberState("my-member", "")
	registry.AddMember(addedMember, WithNowTime(100))

	// Marks the node down.
	registry.UpdateLiveness(1000)

	// Marks the node left.
	registry.UpdateLiveness(7000)

	m, ok := registry.Member("my-member")
	assert.True(t, ok)
	assert.Equal(t, rpc.Liveness_LEFT, m.Liveness)
	assert.Equal(t, int64(57000), m.Expiry)
}

// Tests an owned left member is removed after the tombstone timeout.
func TestFailureDetector_ExpiredOwnedLeftMemberRemoved(t *testing.T) {
	registry := NewRegistry(
		"local",
		WithHeartbeatTimeout(500),
		WithReconnectTimeout(5000),
		WithTombstoneTimeout(50000),
	)

	addedMember := testutils.RandomMemberState("my-member", "")
	registry.AddMember(addedMember, WithNowTime(100))
	registry.RemoveMember("my-member", WithNowTime(1000))

	m, ok := registry.Member("my-member")
	assert.True(t, ok)
	assert.Equal(t, rpc.Liveness_LEFT, m.Liveness)
	assert.Equal(t, int64(51000), m.Expiry)

	// Removes the member.
	registry.UpdateLiveness(60000)

	// The member should not exist.
	_, ok = registry.Member("my-member")
	assert.False(t, ok)
}

// Tests a remote left member is removed after the tombstone timeout.
func TestFailureDetector_ExpiredRemoteLeftMemberRemoved(t *testing.T) {
	registry := NewRegistry("local")

	addedMember := testutils.RandomMemberState("my-member", "")
	registry.RemoteUpdate(&rpc.Member2{
		State:    addedMember,
		Liveness: rpc.Liveness_LEFT,
		Version: &rpc.Version2{
			OwnerId: "remote",
			Timestamp: &rpc.MonotonicTimestamp{
				Timestamp: 100,
			},
		},
		Expiry: 5000,
	})

	m, ok := registry.Member("my-member")
	assert.True(t, ok)
	assert.Equal(t, rpc.Liveness_LEFT, m.Liveness)
	assert.Equal(t, int64(5000), m.Expiry)

	// Removes the member.
	registry.UpdateLiveness(6000)

	// The member should not exist.
	_, ok = registry.Member("my-member")
	assert.False(t, ok)
}

// Tests up members whose owner is down for over the heartbeat timeout are
// marked down.
func TestFailureDetector_DownNodesMembersMarkedDown(t *testing.T) {
	registry := NewRegistry(
		"local",
		WithHeartbeatTimeout(500),
		WithReconnectTimeout(5000),
	)

	addedMember := testutils.RandomMemberState("my-member", "")
	registry.RemoteUpdate(&rpc.Member2{
		State:    addedMember,
		Liveness: rpc.Liveness_UP,
		Version: &rpc.Version2{
			OwnerId: "remote",
			Timestamp: &rpc.MonotonicTimestamp{
				Timestamp: 100,
			},
		},
	})

	// Mark the owner node down.
	registry.OnNodeLeave("remote", WithNowTime(1000))

	registry.UpdateLiveness(2000)

	m, ok := registry.Member("my-member")
	assert.True(t, ok)
	assert.Equal(t, "local", m.Version.OwnerId)
	assert.Equal(t, rpc.Liveness_DOWN, m.Liveness)
	assert.Equal(t, int64(7000), m.Expiry)
}

// Tests down members whose owner is down for over the heartbeat timeout are
// taken but unchanged.
func TestFailureDetector_DownNodesMembersTaken(t *testing.T) {
	registry := NewRegistry(
		"local",
		WithHeartbeatTimeout(500),
		WithReconnectTimeout(5000),
	)

	addedMember := testutils.RandomMemberState("my-member", "")
	registry.RemoteUpdate(&rpc.Member2{
		State:    addedMember,
		Liveness: rpc.Liveness_DOWN,
		Version: &rpc.Version2{
			OwnerId: "remote",
			Timestamp: &rpc.MonotonicTimestamp{
				Timestamp: 100,
			},
		},
		Expiry: 15000,
	})

	// Mark the owner node down.
	registry.OnNodeLeave("remote", WithNowTime(1000))

	registry.UpdateLiveness(2000)

	m, ok := registry.Member("my-member")
	assert.True(t, ok)
	assert.Equal(t, "local", m.Version.OwnerId)
	assert.Equal(t, rpc.Liveness_DOWN, m.Liveness)
	assert.Equal(t, int64(15000), m.Expiry)
}

func TestFailureDetector_LocalNodeIgnored(t *testing.T) {
	localMember := randomMember("local")
	registry := NewRegistry(
		"local",
		WithLocalMember(localMember),
		WithHeartbeatTimeout(100),
		WithReconnectTimeout(1000),
		WithTombstoneTimeout(10000),
	)

	// The local nodes liveness should never change.
	registry.UpdateLiveness(100000)

	m, ok := registry.Member("local")
	assert.True(t, ok)
	assert.Equal(t, rpc.Liveness_UP, m.Liveness)
}
