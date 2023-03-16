package queriers

import (
	"main/pkg/config"
	"main/pkg/types"

	"github.com/prometheus/client_golang/prometheus"
)

type DenomCoefficientQuerier struct {
	Config *config.Config
}

func NewDenomCoefficientQuerier(config *config.Config) *DenomCoefficientQuerier {
	return &DenomCoefficientQuerier{
		Config: config,
	}
}

func (q *DenomCoefficientQuerier) GetMetrics() ([]prometheus.Collector, []types.QueryInfo) {
	denomCoefficientGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_wallets_exporter_denom_coefficient",
			Help: "Denom coefficient info",
		},
		[]string{"chain", "denom", "display_denom"},
	)

	for _, chain := range q.Config.Chains {
		denomCoefficientGauge.With(prometheus.Labels{
			"chain":         chain.Name,
			"display_denom": chain.Denom,
			"denom":         chain.BaseDenom,
		}).Set(float64(chain.DenomCoefficient))
	}

	return []prometheus.Collector{denomCoefficientGauge}, []types.QueryInfo{}
}
