package main

import "time"

type Balance struct {
	Denom  string `json:"denom"`
	Amount string `json:"amount"`
}

type BalanceResponse struct {
	Balances []Balance `json:"balances"`
}

type WalletBalanceEntry struct {
	Chain    string
	Success  bool
	Duration time.Duration
	Wallet   Wallet
	Balances []Balance
}
