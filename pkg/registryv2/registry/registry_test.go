package registry

import (
	"math/rand"
	"testing"
	"time"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestRegistry_AddOwnedMember(t *testing.T) {
	r := NewRegistry("local", time.Now().UnixMilli())

	addedMember := randomMember("", "")
	r.OwnedMemberUpsert(addedMember, time.Now().UnixMilli())

	m, ok := r.Member(addedMember.Id)
	assert.True(t, ok)
	assert.True(t, proto.Equal(addedMember, m))
}

func TestRegistry_LeaveOwnedMember(t *testing.T) {
	r := NewRegistry("local", time.Now().UnixMilli())

	addedMember := randomMember("", "")
	r.OwnedMemberUpsert(addedMember, time.Now().UnixMilli())
	r.OwnedMemberLeave(addedMember.Id, time.Now().UnixMilli())

	_, ok := r.Member(addedMember.Id)
	assert.False(t, ok)
}

func TestRegistry_UpsertOwnedMemberDiscardExpiredUpdate(t *testing.T) {
	r := NewRegistry("local", time.Now().UnixMilli())

	addedMember := randomMember("", "")
	r.OwnedMemberUpsert(addedMember, time.Now().UnixMilli()+1000)

	// Update the member but with a version timestamp less than the already
	// added member.
	updatedMember := randomMember(addedMember.Id, "")
	r.OwnedMemberUpsert(updatedMember, time.Now().UnixMilli()-1000)

	m, ok := r.Member(addedMember.Id)
	assert.True(t, ok)
	assert.True(t, proto.Equal(addedMember, m))
}

func TestRegistry_LeaveOwnedMemberDiscardExpiredUpdate(t *testing.T) {
	r := NewRegistry("local", time.Now().UnixMilli())

	addedMember := randomMember("", "")
	r.OwnedMemberUpsert(addedMember, time.Now().UnixMilli()+1000)

	// Leave but with a version timestamp less than the already
	// added member.
	r.OwnedMemberLeave(addedMember.Id, time.Now().UnixMilli()-1000)

	m, ok := r.Member(addedMember.Id)
	assert.True(t, ok)
	assert.True(t, proto.Equal(addedMember, m))
}

func TestRegistry_UpsertRemoteMember(t *testing.T) {
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

	m, ok := r.Member(addedMember.Id)
	assert.True(t, ok)
	assert.True(t, proto.Equal(addedMember, m))
}

func TestRegistry_UpsertRemoteMemberDiscardExpiredUpdate(t *testing.T) {
	r := NewRegistry("local", time.Now().UnixMilli())

	addedMember := randomMember("", "")
	r.OwnedMemberUpsert(addedMember, time.Now().UnixMilli()+1000)

	// Update the member but with a version timestamp less than the already
	// added member.
	r.RemoteUpsertMember(&rpc.Member2{
		State:    randomMember(addedMember.Id, ""),
		Liveness: rpc.Liveness_UP,
		Version: &rpc.Version2{
			OwnerId: "remote",
			Timestamp: &rpc.MonotonicTimestamp{
				Timestamp: time.Now().UnixMilli() - 1000,
			},
		},
	})

	m, ok := r.Member(addedMember.Id)
	assert.True(t, ok)
	assert.True(t, proto.Equal(addedMember, m))
}

func TestRegistry_RegisterLocalMember(t *testing.T) {
	localMember := randomMember("local", "fuddle")
	r := NewRegistry("local", time.Now().UnixMilli(), WithLocalMember(localMember))

	m, ok := r.Member("local")
	assert.True(t, ok)
	assert.True(t, proto.Equal(localMember, m))
}

func TestRegistry_UpdateLocalMemberDiscarded(t *testing.T) {
	localMember := randomMember("local", "fuddle")
	r := NewRegistry("local", time.Now().UnixMilli(), WithLocalMember(localMember))

	r.OwnedMemberUpsert(randomMember("local", ""), time.Now().UnixMilli())

	// Check the local member was not updated.
	m, ok := r.Member("local")
	assert.True(t, ok)
	assert.True(t, proto.Equal(localMember, m))
}

func TestRegistry_LeaveLocalMemberDiscarded(t *testing.T) {
	localMember := randomMember("local", "fuddle")
	r := NewRegistry("local", time.Now().UnixMilli(), WithLocalMember(localMember))

	r.OwnedMemberLeave(localMember.Id, time.Now().UnixMilli())

	// Check the local member was not deleted.
	m, ok := r.Member("local")
	assert.True(t, ok)
	assert.True(t, proto.Equal(localMember, m))
}

func randomMember(id string, service string) *rpc.MemberState {
	if id == "" {
		id = uuid.New().String()
	}
	if service == "" {
		service = uuid.New().String()
	}
	return &rpc.MemberState{
		Id:      id,
		Service: service,
		Status:  uuid.New().String(),
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
