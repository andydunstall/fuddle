package registry

import (
	"sort"
	"testing"

	"github.com/fuddle-io/fuddle/pkg/testutils"
	"github.com/stretchr/testify/assert"
	"google.golang.org/protobuf/proto"
)

// Tests two random registies learn about one another members after exchanging
// deltas.
func TestReplicaRepair(t *testing.T) {
	registry1 := NewRegistry("node-1")
	registry2 := NewRegistry("node-2")
	for i := 0; i != 25; i++ {
		registry1.AddMember(testutils.RandomMemberState("", ""))
		registry2.AddMember(testutils.RandomMemberState("", ""))
	}

	// Send a digest so registry1 discovers the members it doesn't know about.
	registry1.Delta(registry2.Digest(50))

	delta2 := registry2.Delta(registry1.Digest(50))
	// Create delta 1 again since the first one will be empty as the registry
	// doesn't know what its missing.
	delta1 := registry1.Delta(registry2.Digest(50))
	for _, member := range delta1 {
		registry2.RemoteUpdate(member)
	}
	for _, member := range delta2 {
		registry1.RemoteUpdate(member)
	}

	members1 := registry1.Members()
	members2 := registry2.Members()
	sort.Slice(members1, func(i, j int) bool {
		return members1[i].State.Id < members1[j].State.Id
	})
	sort.Slice(members2, func(i, j int) bool {
		return members2[i].State.Id < members2[j].State.Id
	})

	assert.Equal(t, len(members1), len(members2))
	for i := 0; i != len(members1); i++ {
		assert.True(t, proto.Equal(members1[i], members2[i]))
	}
}
