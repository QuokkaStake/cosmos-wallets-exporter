package logger

import (
	"main/pkg/config"
	"testing"

	"github.com/stretchr/testify/require"
)

func TestGetDefaultLogger(t *testing.T) {
	t.Parallel()

	logger := GetDefaultLogger()
	require.NotNil(t, logger)
}

func TestGetLoggerInvalidLogLevel(t *testing.T) {
	t.Parallel()

	defer func() {
		if r := recover(); r == nil {
			require.Fail(t, "Expected to have a panic here!")
		}
	}()

	GetLogger(config.LogConfig{LogLevel: "invalid"})
}

func TestGetLoggerValidPlain(t *testing.T) {
	t.Parallel()

	logger := GetLogger(config.LogConfig{LogLevel: "info"})
	require.NotNil(t, logger)
}

func TestGetLoggerValidJSON(t *testing.T) {
	t.Parallel()

	logger := GetLogger(config.LogConfig{LogLevel: "info", JSONOutput: true})
	require.NotNil(t, logger)
}

func TestGetLoggerNop(t *testing.T) {
	t.Parallel()

	logger := GetNopLogger()
	require.NotNil(t, logger)
}
