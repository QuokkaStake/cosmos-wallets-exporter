package queriers

import (
	"main/pkg/config"
	"main/pkg/tendermint"
	"main/pkg/types"
	"main/pkg/utils"
	"sync"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

type BalanceQuerier struct {
	Config *config.Config
	Logger zerolog.Logger
}

func NewBalanceQuerier(config *config.Config, logger zerolog.Logger) *BalanceQuerier {
	return &BalanceQuerier{
		Config: config,
		Logger: logger.With().Str("component", "balance_querier").Logger(),
	}
}

func (q *BalanceQuerier) GetMetrics() ([]prometheus.Collector, []types.QueryInfo) {
	balancesGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_wallets_exporter_balance",
			Help: "A wallet balance (in tokens)",
		},
		[]string{"chain", "address", "name", "group", "denom"},
	)

	var queryInfos []types.QueryInfo

	var wg sync.WaitGroup
	var mutex sync.Mutex

	for _, chain := range q.Config.Chains {
		rpc := tendermint.NewRPC(chain, q.Logger)

		for _, wallet := range chain.Wallets {
			wg.Add(1)
			go func(wallet config.Wallet, chain config.Chain) {
				defer wg.Done()

				balancesResponse, queryInfo, err := rpc.GetWalletBalances(wallet.Address)

				mutex.Lock()
				defer mutex.Unlock()

				queryInfos = append(queryInfos, queryInfo)

				if err != nil {
					q.Logger.Error().
						Err(err).
						Str("chain", chain.Name).
						Str("wallet", wallet.Address).
						Msg("Error querying balance")
					return
				}

				for _, balance := range balancesResponse.Balances {
					denom := balance.Denom
					amount := utils.StrToFloat64(balance.Amount)

					denomInfo, found := chain.FindDenomByName(balance.Denom)
					if found {
						denom = denomInfo.GetName()
						amount /= float64(denomInfo.DenomCoefficient)
					}

					balancesGauge.With(prometheus.Labels{
						"chain":   chain.Name,
						"address": wallet.Address,
						"name":    wallet.Name,
						"group":   wallet.Group,
						"denom":   denom,
					}).Set(amount)
				}
			}(wallet, chain)
		}
	}

	wg.Wait()

	return []prometheus.Collector{balancesGauge}, queryInfos
}
