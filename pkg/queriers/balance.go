package queriers

import (
	"context"
	"main/pkg/config"
	"main/pkg/tendermint"
	"main/pkg/types"
	"main/pkg/utils"
	"sync"

	"go.opentelemetry.io/otel/attribute"
	"go.opentelemetry.io/otel/trace"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

type BalanceQuerier struct {
	Config *config.Config
	Logger zerolog.Logger
	RPCs   []*tendermint.RPC
	Tracer trace.Tracer
}

func NewBalanceQuerier(
	config *config.Config,
	logger zerolog.Logger,
	tracer trace.Tracer,
) *BalanceQuerier {
	rpcs := make([]*tendermint.RPC, len(config.Chains))

	for index, chain := range config.Chains {
		rpcs[index] = tendermint.NewRPC(chain, logger, tracer)
	}

	return &BalanceQuerier{
		Config: config,
		Logger: logger.With().Str("component", "balance_querier").Logger(),
		RPCs:   rpcs,
		Tracer: tracer,
	}
}

func (q *BalanceQuerier) GetMetrics(ctx context.Context) ([]prometheus.Collector, []types.QueryInfo) {
	childCtx, span := q.Tracer.Start(ctx, "Querying balance metrics")
	defer span.End()

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

	for index, chain := range q.Config.Chains {
		rpc := q.RPCs[index]

		for _, wallet := range chain.Wallets {
			wg.Add(1)
			go func(wallet config.Wallet, chain config.Chain, rpc *tendermint.RPC) {
				chainCtx, chainSpan := q.Tracer.Start(childCtx, "Querying chain and wallet")
				chainSpan.SetAttributes(attribute.String("chain", chain.Name))
				chainSpan.SetAttributes(attribute.String("wallet", wallet.Address))
				defer chainSpan.End()

				defer wg.Done()

				balancesResponse, queryInfo, err := rpc.GetWalletBalances(wallet.Address, chainCtx)

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
			}(wallet, chain, rpc)
		}
	}

	wg.Wait()

	return []prometheus.Collector{balancesGauge}, queryInfos
}
