package queriers

import (
	"context"
	"errors"
	"main/assets"
	coingeckoPkg "main/pkg/coingecko"
	configPkg "main/pkg/config"
	loggerPkg "main/pkg/logger"
	"main/pkg/tracing"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

//nolint:paralleltest // disabled due to httpmock usage
func TestPriceQuerierFail(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.coingecko.com/api/v3/simple/price?ids=cosmos&vs_currencies=usd",
		httpmock.NewErrorResponder(errors.New("custom error")),
	)

	config := &configPkg.Config{Chains: []configPkg.Chain{{
		Name:   "chain",
		Denoms: []configPkg.DenomInfo{{Denom: "atom", CoingeckoCurrency: "cosmos"}},
	}}}

	tracer := tracing.InitNoopTracer()
	logger := loggerPkg.GetDefaultLogger()
	coingecko := coingeckoPkg.NewCoingecko(config, *logger, tracer)
	querier := NewPriceQuerier(config, coingecko, tracer)

	metrics, queries := querier.GetMetrics(context.Background())
	assert.Len(t, queries, 1)
	assert.False(t, queries[0].Success)

	assert.Len(t, metrics, 1)
	assert.Zero(t, testutil.CollectAndCount(metrics[0]))
}

//nolint:paralleltest // disabled due to httpmock usage
func TestPriceQuerierOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://api.coingecko.com/api/v3/simple/price?ids=cosmos,random&vs_currencies=usd",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("coingecko.json")),
	)

	config := &configPkg.Config{Chains: []configPkg.Chain{{
		Name: "chain",
		Denoms: []configPkg.DenomInfo{
			{Denom: "atom", CoingeckoCurrency: "cosmos"},
			{Denom: "random", CoingeckoCurrency: "random"},
			{Denom: "unknown"},
		},
	}}}

	tracer := tracing.InitNoopTracer()
	logger := loggerPkg.GetDefaultLogger()
	coingecko := coingeckoPkg.NewCoingecko(config, *logger, tracer)
	querier := NewPriceQuerier(config, coingecko, tracer)

	metrics, queries := querier.GetMetrics(context.Background())
	assert.Len(t, queries, 1)
	assert.True(t, queries[0].Success)

	assert.Len(t, metrics, 1)

	pricesMetric, ok := metrics[0].(*prometheus.GaugeVec)
	assert.True(t, ok)

	assert.InDelta(t, 1, testutil.CollectAndCount(pricesMetric), 0.001)
	assert.InDelta(t, 5.84, testutil.ToFloat64(pricesMetric.With(prometheus.Labels{
		"chain": "chain",
		"denom": "atom",
	})), 0.01)
}
