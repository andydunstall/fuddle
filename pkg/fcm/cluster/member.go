package cluster

import (
	"context"
	"fmt"
	"math/rand"
	"time"

	"github.com/fuddle-io/fuddle-go"
	"github.com/google/uuid"
	"go.uber.org/zap"
)

// MemberNode is a random member that registers with Fuddle.
type MemberNode struct {
	ID       string
	Registry *fuddle.Fuddle
}

func NewMemberNode(id string, fuddleAddrs []string, logger *zap.Logger) (*MemberNode, error) {
	member := randomMember(id)

	ctx, cancel := context.WithTimeout(context.Background(), time.Second*5)
	defer cancel()

	registry, err := fuddle.Connect(
		ctx,
		member,
		fuddleAddrs,
		fuddle.WithLogger(logger),
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

func randomMember(id string) fuddle.Member {
	return fuddle.Member{
		ID:      id,
		Service: uuid.New().String(),
		Locality: fuddle.Locality{
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
