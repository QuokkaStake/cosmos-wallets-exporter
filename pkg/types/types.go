package types

import (
	"context"
	"main/pkg/config"
	"time"

	"github.com/prometheus/client_golang/prometheus"
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
}

type QueryInfo struct {
	Chain    string
	Success  bool
	URL      string
	Duration time.Duration
}

type Querier interface {
	GetMetrics(ctx context.Context) ([]prometheus.Collector, []QueryInfo)
}
