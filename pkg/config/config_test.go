package config

import (
	"main/pkg/fs"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

func TestConfigNoChains(t *testing.T) {
	t.Parallel()

	chain := &Config{}
	err := chain.Validate()
	require.Error(t, err)
	require.ErrorContains(t, err, "no chains provided")
}

func TestConfigInvalidChain(t *testing.T) {
	t.Parallel()

	chain := &Config{Chains: []Chain{{}}}
	err := chain.Validate()
	require.Error(t, err)
	require.ErrorContains(t, err, "error in chain 0")
}

func TestConfigValid(t *testing.T) {
	t.Parallel()

	chain := &Config{Chains: []Chain{{
		Name:        "chain",
		LCDEndpoint: "test",
		Wallets:     []Wallet{{Address: "address"}},
	}}}
	err := chain.Validate()
	require.NoError(t, err)
}

func TestGetCoingeckoCurrencies(t *testing.T) {
	t.Parallel()

	chain := &Config{Chains: []Chain{{
		Name:        "chain",
		LCDEndpoint: "test",
		Wallets:     []Wallet{{Address: "address"}},
		Denoms: []DenomInfo{
			{Denom: "uatom", DisplayDenom: "atom", CoingeckoCurrency: "cosmos"},
			{Denom: "unom", DisplayDenom: "nom"},
		},
	}}}
	currencies := chain.GetCoingeckoCurrencies()
	require.Len(t, currencies, 1)
	assert.Equal(t, "cosmos", currencies[0])
}

func TestLoadConfigFailedToLoad(t *testing.T) {
	t.Parallel()

	filesystem := &fs.TestFS{}
	config, err := GetConfig("not-found", filesystem)
	require.Nil(t, config)
	require.Error(t, err)
}

func TestLoadConfigFailedToDecode(t *testing.T) {
	t.Parallel()

	filesystem := &fs.TestFS{}
	config, err := GetConfig("invalid.toml", filesystem)
	require.Nil(t, config)
	require.Error(t, err)
}

func TestLoadConfigValid(t *testing.T) {
	t.Parallel()

	filesystem := &fs.TestFS{}
	config, err := GetConfig("config-valid.toml", filesystem)
	require.NotNil(t, config)
	require.NoError(t, err)
}
