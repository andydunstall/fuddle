package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetrics_AddOwnedMember(t *testing.T) {
	r := NewRegistry("local")

	addedMember := randomMember("", "foo")
	r.OwnedMemberAdd(addedMember)

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

func TestMetrics_RemoveOwnedMember(t *testing.T) {
	r := NewRegistry("local")

	addedMember := randomMember("", "foo")
	r.OwnedMemberAdd(addedMember)
	r.OwnedMemberLeave(addedMember.Id)

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

func TestMetrics_AddLocalMember(t *testing.T) {
	localMember := randomMember("local", "fuddle")
	r := NewRegistry("local", WithLocalMember(localMember))

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
