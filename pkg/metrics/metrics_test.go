package metrics

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestGauge(t *testing.T) {
	gauge := NewGauge("foo", []string{"a", "b", "c"}, "")

	gauge.Set(5.0, map[string]string{
		"a": "1",
		"b": "2",
		"c": "3",
	})
	gauge.Set(3.0, map[string]string{
		"a": "8",
		"b": "9",
		"c": "10",
	})

	assert.Equal(t, 3.0, gauge.Value(map[string]string{
		"a": "8",
		"b": "9",
		"c": "10",
	}))
	assert.Equal(t, 5.0, gauge.Value(map[string]string{
		"a": "1",
		"b": "2",
		"c": "3",
	}))
	assert.Equal(t, 0.0, gauge.Value(map[string]string{
		"a": "99",
		"b": "2",
		"c": "3",
	}))

}
