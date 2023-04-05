package registry

import (
	"testing"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"github.com/stretchr/testify/assert"
)

func TestRegistry_SubscribeLocalMember(t *testing.T) {
	r := NewRegistry(WithRegistryLocalMember(&rpc.Member{
		Id: "local-member",
	}))

	var update *rpc.RemoteMemberUpdate
	unsub := r.Subscribe(nil, func(u *rpc.RemoteMemberUpdate) {
		update = u
	})
	defer unsub()

	assert.Equal(t, "local-member", update.Id)
	assert.Equal(t, rpc.MemberUpdateType_REGISTER, update.UpdateType)
	assert.Equal(t, "local-member", update.Member.Member.Id)
}
