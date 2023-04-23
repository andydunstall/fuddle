package testutils

import (
	"math/rand"

	rpc "github.com/fuddle-io/fuddle-rpc/go"
	"github.com/google/uuid"
)

func RandomMemberState(id string, service string) *rpc.MemberState {
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
