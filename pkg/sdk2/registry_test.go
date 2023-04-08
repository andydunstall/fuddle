package fuddle

import (
	"math/rand"
	"testing"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
)

func TestRegistry_RemoteUpdateAddMember(t *testing.T) {
	reg := newRegistry()

	addedMember := randomMember("member-1")
	reg.RemoteUpdate(&rpc.RemoteMemberUpdate{
		Member: addedMember,
		Version: &rpc.Version{
			Owner:     "remote-1",
			Timestamp: 123,
		},
	})

	assert.Equal(t, []Member{fromRPC(addedMember)}, reg.Members())
}

func TestRegistry_RemoteUpdateRemoveMember(t *testing.T) {
	reg := newRegistry()

	reg.RemoteUpdate(&rpc.RemoteMemberUpdate{
		Member: randomMember("member-1"),
		Version: &rpc.Version{
			Owner:     "remote-1",
			Timestamp: 123,
		},
	})
	reg.RemoteUpdate(&rpc.RemoteMemberUpdate{
		Member: &rpc.Member{
			Id:     "member-1",
			Status: rpc.MemberStatus_LEFT,
		},
		Version: &rpc.Version{
			Owner:     "remote-1",
			Timestamp: 123,
		},
	})

	assert.Nil(t, reg.Members())
}

func TestRegistry_Subscribe(t *testing.T) {
	reg := newRegistry()

	count := 0
	reg.Subscribe(func() {
		count++
	})

	reg.RemoteUpdate(&rpc.RemoteMemberUpdate{
		Member: randomMember("member-1"),
		Version: &rpc.Version{
			Owner:     "remote-1",
			Timestamp: 123,
		},
	})
	reg.RemoteUpdate(&rpc.RemoteMemberUpdate{
		Member: &rpc.Member{
			Id:     "member-1",
			Status: rpc.MemberStatus_LEFT,
		},
		Version: &rpc.Version{
			Owner:     "remote-1",
			Timestamp: 123,
		},
	})

	assert.Equal(t, 3, count)
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
