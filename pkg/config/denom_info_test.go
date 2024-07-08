package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
)

func TestDenomInfoGetName(t *testing.T) {
	t.Parallel()

	assert.Equal(t, "denom", DenomInfo{Denom: "denom"}.GetName())
	assert.Equal(t, "display", DenomInfo{Denom: "denom", DisplayDenom: "display"}.GetName())
}
