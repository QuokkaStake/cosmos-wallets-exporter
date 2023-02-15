package types

import (
	"main/pkg/config"
	"time"
)

type Balance struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

type Balances []Balance

type BalanceResponse struct {
	Balances Balances `json:"balances"`
}

type WalletBalanceEntry struct {
	Chain    string
	Success  bool
	Duration time.Duration
	Wallet   config.Wallet
	Balances Balances
	UsdPrice float64
}
