package logger

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestLogger_WarningCount(t *testing.T) {
	logger, err := NewLogger()
	require.NoError(t, err)

	logger.Logger("foo").Warn("log")
	logger.Logger("foo").Warn("log")

	// Info and error logs should not count.
	logger.Logger("foo").Info("log")
	logger.Logger("foo").Error("log")

	assert.Equal(t, 2.0, logger.Metrics().WarningsCount.Value(map[string]string{
		"subsystem": "foo",
	}))
}

func TestLogger_ErrorCount(t *testing.T) {
	logger, err := NewLogger()
	require.NoError(t, err)

	logger.Logger("foo").Error("log")
	logger.Logger("foo").Error("log")

	// Info and warn logs should not count.
	logger.Logger("foo").Info("log")
	logger.Logger("foo").Warn("log")

	assert.Equal(t, 2.0, logger.Metrics().ErrorsCount.Value(map[string]string{
		"subsystem": "foo",
	}))
}
