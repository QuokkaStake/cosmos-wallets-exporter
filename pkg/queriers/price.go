package queriers

import (
	coingeckoPkg "main/pkg/coingecko"
	"main/pkg/config"
	"main/pkg/types"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/rs/zerolog"
)

type PriceQuerier struct {
	Config    *config.Config
	Coingecko *coingeckoPkg.Coingecko
	Logger    zerolog.Logger
}

func NewPriceQuerier(config *config.Config, coingecko *coingeckoPkg.Coingecko) *PriceQuerier {
	return &PriceQuerier{
		Config:    config,
		Coingecko: coingecko,
	}
}

func (q *PriceQuerier) GetMetrics() ([]prometheus.Collector, []types.QueryInfo) {
	priceGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_wallets_exporter_price",
			Help: "A price of 1 token",
		},
		[]string{"chain"},
	)

	currenciesList := q.Config.GetCoingeckoCurrencies()
	currenciesRates := q.Coingecko.FetchPrices(currenciesList)

	for chain, price := range currenciesRates {
		priceGauge.With(prometheus.Labels{
			"chain": chain,
		}).Set(price)
	}

	return []prometheus.Collector{priceGauge}, []types.QueryInfo{}
}
