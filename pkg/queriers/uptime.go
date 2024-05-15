package queriers

import (
	"context"
	"main/pkg/types"
	"time"

	"go.opentelemetry.io/otel/trace"

	"github.com/prometheus/client_golang/prometheus"
)

type UptimeQuerier struct {
	StartTime time.Time
	Tracer    trace.Tracer
}

func NewUptimeQuerier(tracer trace.Tracer) *UptimeQuerier {
	return &UptimeQuerier{
		StartTime: time.Now(),
		Tracer:    tracer,
	}
}

func (u *UptimeQuerier) GetMetrics(ctx context.Context) ([]prometheus.Collector, []types.QueryInfo) {
	uptimeMetricsGauge := prometheus.NewGaugeVec(
		prometheus.GaugeOpts{
			Name: "cosmos_wallets_exporter_start_time",
			Help: "Unix timestamp on when the app was started. Useful for annotations.",
		},
		[]string{},
	)

	uptimeMetricsGauge.With(prometheus.Labels{}).Set(float64(u.StartTime.Unix()))
	return []prometheus.Collector{uptimeMetricsGauge}, []types.QueryInfo{}
}
