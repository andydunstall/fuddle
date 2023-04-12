package cluster

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/fuddle-io/fuddle-go"
	"github.com/google/uuid"
)

// MemberNode is a random member that registers with Fuddle.
type MemberNode struct {
	ID       string
	Registry *fuddle.Fuddle
}

func NewMemberNode(fuddleAddrs []string) (*MemberNode, error) {
	member := randomMember()

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	registry, err := fuddle.Register(
		ctx,
		fuddleAddrs,
		member,
	)
	if err != nil {
		return nil, fmt.Errorf("member register: %w", err)
	}

	return &MemberNode{
		ID:       member.ID,
		Registry: registry,
	}, nil
}

func (n *MemberNode) Shutdown() {
	n.Registry.Close()
}

func randomMember() fuddle.Member {
	return fuddle.Member{
		ID:       "member-" + uuid.New().String()[:8],
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
