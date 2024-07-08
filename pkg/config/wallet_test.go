package config

import (
	"testing"

	"github.com/stretchr/testify/require"
)

func TestWalletNoName(t *testing.T) {
	t.Parallel()

	wallet := Wallet{}
	err := wallet.Validate()
	require.Error(t, err)
	require.ErrorContains(t, err, "address for wallet is not specified")
}

func TestWalletValid(t *testing.T) {
	t.Parallel()

	wallet := Wallet{Address: "wallet"}
	err := wallet.Validate()
	require.NoError(t, err)
}
