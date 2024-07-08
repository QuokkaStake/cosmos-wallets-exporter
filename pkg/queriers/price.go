package queriers

import (
	"context"
	coingeckoPkg "main/pkg/coingecko"
	"main/pkg/config"
	"main/pkg/types"

	"go.opentelemetry.io/otel/trace"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

type PriceQuerier struct {
	Config    *config.Config
	Coingecko *coingeckoPkg.Coingecko
	Logger    zerolog.Logger
	Tracer    trace.Tracer
}

func NewPriceQuerier(
	config *config.Config,
	coingecko *coingeckoPkg.Coingecko,
	tracer trace.Tracer,
) *PriceQuerier {
	return &PriceQuerier{
		Config:    config,
		Coingecko: coingecko,
		Tracer:    tracer,
	}
}

func (q *PriceQuerier) GetMetrics(ctx context.Context) ([]prometheus.Collector, []types.QueryInfo) {
	childCtx, span := q.Tracer.Start(ctx, "Querying prices")
	defer span.End()

	priceGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_wallets_exporter_price",
			Help: "A price of 1 token",
		},
		[]string{"chain", "denom"},
	)

	currenciesList := q.Config.GetCoingeckoCurrencies()
	currenciesRates, queryInfo := q.Coingecko.FetchPrices(currenciesList, childCtx)

	for _, chain := range q.Config.Chains {
		for _, denom := range chain.Denoms {
			if denom.CoingeckoCurrency == "" {
				continue
			}

			if price, ok := currenciesRates[denom.CoingeckoCurrency]; ok {
				priceGauge.With(prometheus.Labels{
					"chain": chain.Name,
					"denom": denom.GetName(),
				}).Set(price)
			}
		}
	}

	return []prometheus.Collector{priceGauge}, []types.QueryInfo{queryInfo}
}
