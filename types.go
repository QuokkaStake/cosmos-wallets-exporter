package main

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
	Wallet   Wallet
	Balances []Balance
}
