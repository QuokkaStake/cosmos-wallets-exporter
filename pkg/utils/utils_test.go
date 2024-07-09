package utils

import (
	"main/pkg/constants"
	"net/http"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestBoolToFloat64(t *testing.T) {
	t.Parallel()
	assert.InDelta(t, float64(1), BoolToFloat64(true), 0.001)
	assert.InDelta(t, float64(0), BoolToFloat64(false), 0.001)
}

func TestGetBlockFromHeaderNoValue(t *testing.T) {
	t.Parallel()

	header := http.Header{}
	value, err := GetBlockHeightFromHeader(header)

	require.NoError(t, err)
	assert.Equal(t, int64(0), value)
}

func TestGetBlockFromHeaderInvalidValue(t *testing.T) {
	t.Parallel()

	header := http.Header{
		constants.HeaderBlockHeight: []string{"invalid"},
	}
	value, err := GetBlockHeightFromHeader(header)

	require.Error(t, err)
	assert.Equal(t, int64(0), value)
}

func TestGetBlockFromHeaderValidValue(t *testing.T) {
	t.Parallel()

	header := http.Header{
		constants.HeaderBlockHeight: []string{"123"},
	}
	value, err := GetBlockHeightFromHeader(header)

	require.NoError(t, err)
	assert.Equal(t, int64(123), value)
}
