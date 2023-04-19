package registry

import (
	"math/rand"
	"testing"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

func TestRegistry_RegisterLocalMember(t *testing.T) {
	localMember := randomMember("local", "fuddle")
	r := NewRegistry("local", WithLocalMember(localMember))

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
