package queriers

import (
	"main/pkg/config"
	"main/pkg/types"

	"github.com/prometheus/client_golang/prometheus"
)

type QueriesQuerier struct {
	Config *config.Config
	Infos  []types.QueryInfo
}

func NewQueriesQuerier(appConfig *config.Config, queryInfos []types.QueryInfo) *QueriesQuerier {
	return &QueriesQuerier{
		Config: appConfig,
		Infos:  queryInfos,
	}
}

func (q *QueriesQuerier) GetMetrics() ([]prometheus.Collector, []types.QueryInfo) {
	successGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_wallets_exporter_success",
			Help: "Whether a scrape was successful",
		},
		[]string{"chain"},
	)

	errorGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_wallets_exporter_error",
			Help: "Whether a scrape has errors",
		},
		[]string{"chain"},
	)

	timingsGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_wallets_exporter_timings",
			Help: "External LCD query timing",
		},
		[]string{"chain", "url"},
	)

	// so we would have this metrics even if there are no requests
	for _, chain := range q.Config.Chains {
		successGauge.With(prometheus.Labels{
			"chain": chain.Name,
		}).Set(0)

		errorGauge.With(prometheus.Labels{
			"chain": chain.Name,
		}).Set(0)
	}

	for _, query := range q.Infos {
		timingsGauge.With(prometheus.Labels{
			"chain": query.Chain,
			"url":   query.URL,
		}).Set(query.Duration.Seconds())

		if query.Success {
			successGauge.With(prometheus.Labels{
				"chain": query.Chain,
			}).Inc()
		} else {
			errorGauge.With(prometheus.Labels{
				"chain": query.Chain,
			}).Inc()
		}
	}

	return []prometheus.Collector{
		successGauge,
		errorGauge,
		timingsGauge,
	}, []types.QueryInfo{}
}
