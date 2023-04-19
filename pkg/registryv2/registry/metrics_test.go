package registry

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestMetrics_RegisterLocalMember(t *testing.T) {
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
