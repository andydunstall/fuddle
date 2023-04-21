package registry

import (
	"testing"
	"time"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"github.com/stretchr/testify/assert"
)

func TestMetrics_AddOwnedMember(t *testing.T) {
	r := NewRegistry("local", time.Now().UnixMilli())

	addedMember := randomMember("", "foo")
	r.OwnedMemberUpsert(addedMember, time.Now().UnixMilli())

	assert.Equal(t, 1.0, r.Metrics().MembersCount.Value(map[string]string{
		"liveness": "up",
		"owner":    "local",
		"service":  "foo",
	}))
	assert.Equal(t, 1.0, r.Metrics().MembersOwned.Value(map[string]string{
		"liveness": "up",
		"service":  "foo",
	}))
}

func TestMetrics_LeaveOwnedMember(t *testing.T) {
	r := NewRegistry("local", time.Now().UnixMilli())

	addedMember := randomMember("", "foo")
	r.OwnedMemberUpsert(addedMember, time.Now().UnixMilli())
	r.OwnedMemberLeave(addedMember.Id, time.Now().UnixMilli())

	assert.Equal(t, 0.0, r.Metrics().MembersCount.Value(map[string]string{
		"liveness": "up",
		"owner":    "local",
		"service":  "foo",
	}))
	assert.Equal(t, 0.0, r.Metrics().MembersOwned.Value(map[string]string{
		"liveness": "up",
		"service":  "foo",
	}))
	assert.Equal(t, 1.0, r.Metrics().MembersCount.Value(map[string]string{
		"liveness": "left",
		"owner":    "local",
		"service":  "foo",
	}))
	assert.Equal(t, 1.0, r.Metrics().MembersOwned.Value(map[string]string{
		"liveness": "left",
		"service":  "foo",
	}))
}

func TestMetrics_AddRemoteMember(t *testing.T) {
	r := NewRegistry("local", time.Now().UnixMilli())

	addedMember := randomMember("", "")
	r.RemoteUpsertMember(&rpc.Member2{
		State:    addedMember,
		Liveness: rpc.Liveness_UP,
		Version: &rpc.Version2{
			OwnerId: "remote",
			Timestamp: &rpc.MonotonicTimestamp{
				Timestamp: time.Now().UnixMilli() + 1000,
			},
		},
	})

	assert.Equal(t, 1.0, r.Metrics().MembersCount.Value(map[string]string{
		"liveness": "up",
		"owner":    "remote",
		"service":  addedMember.Service,
	}))
	// Should not update owned members.
	assert.Equal(t, 0.0, r.Metrics().MembersOwned.Value(map[string]string{
		"liveness": "up",
		"service":  "foo",
	}))
}

func TestMetrics_UpsertRemoteMember(t *testing.T) {
	r := NewRegistry("local", time.Now().UnixMilli())

	addedMember := randomMember("", "")
	r.RemoteUpsertMember(&rpc.Member2{
		State:    addedMember,
		Liveness: rpc.Liveness_UP,
		Version: &rpc.Version2{
			OwnerId: "remote-1",
			Timestamp: &rpc.MonotonicTimestamp{
				Timestamp: time.Now().UnixMilli() + 1000,
			},
		},
	})
	// Update the owner and status of the member.
	r.RemoteUpsertMember(&rpc.Member2{
		State:    addedMember,
		Liveness: rpc.Liveness_LEFT,
		Version: &rpc.Version2{
			OwnerId: "remote-2",
			Timestamp: &rpc.MonotonicTimestamp{
				Timestamp: time.Now().UnixMilli() + 2000,
			},
		},
	})

	assert.Equal(t, 0.0, r.Metrics().MembersCount.Value(map[string]string{
		"liveness": "up",
		"owner":    "remote-1",
		"service":  addedMember.Service,
	}))
	assert.Equal(t, 1.0, r.Metrics().MembersCount.Value(map[string]string{
		"liveness": "left",
		"owner":    "remote-2",
		"service":  addedMember.Service,
	}))
}

func TestMetrics_AddLocalMember(t *testing.T) {
	localMember := randomMember("local", "fuddle")
	r := NewRegistry("local", time.Now().UnixMilli(), WithLocalMember(localMember))

	assert.Equal(t, 1.0, r.Metrics().MembersCount.Value(map[string]string{
		"liveness": "up",
		"owner":    "local",
		"service":  "fuddle",
	}))
	assert.Equal(t, 1.0, r.Metrics().MembersOwned.Value(map[string]string{
		"liveness": "up",
		"service":  "fuddle",
	}))
}
