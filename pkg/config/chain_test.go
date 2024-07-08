package config

import (
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestChainNoName(t *testing.T) {
	t.Parallel()

	chain := &Chain{}
	err := chain.Validate()
	require.Error(t, err)
	require.ErrorContains(t, err, "empty chain name")
}

func TestChainNoLcdEndpoint(t *testing.T) {
	t.Parallel()

	chain := &Chain{Name: "chain"}
	err := chain.Validate()
	require.Error(t, err)
	require.ErrorContains(t, err, "no LCD endpoint provided")
}

func TestChainNoWallets(t *testing.T) {
	t.Parallel()

	chain := &Chain{Name: "chain", LCDEndpoint: "test"}
	err := chain.Validate()
	require.Error(t, err)
	require.ErrorContains(t, err, "no wallets provided")
}

func TestChainInvalidWallet(t *testing.T) {
	t.Parallel()

	chain := &Chain{Name: "chain", LCDEndpoint: "test", Wallets: []Wallet{{}}}
	err := chain.Validate()
	require.Error(t, err)
	require.ErrorContains(t, err, "error in wallet 0")
}

func TestChainValid(t *testing.T) {
	t.Parallel()

	chain := &Chain{
		Name:        "chain",
		LCDEndpoint: "test",
		Wallets:     []Wallet{{Address: "address"}},
	}
	err := chain.Validate()
	require.NoError(t, err)
}

func TestChainFindDenomByName(t *testing.T) {
	t.Parallel()

	chain := &Chain{
		Denoms: []DenomInfo{
			{Denom: "denom1"},
		},
	}

	denom1, found1 := chain.FindDenomByName("denom1")
	require.NotNil(t, denom1)
	assert.True(t, found1)

	denom2, found2 := chain.FindDenomByName("denom2")
	require.Nil(t, denom2)
	assert.False(t, found2)
}
