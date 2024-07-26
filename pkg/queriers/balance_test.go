package queriers

import (
	"context"
	"errors"
	"main/assets"
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
func TestBalanceQuerierFail(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/cosmos/bank/v1beta1/balances/address",
		httpmock.NewErrorResponder(errors.New("custom error")),
	)

	config := &configPkg.Config{Chains: []configPkg.Chain{{
		Name:        "chain",
		LCDEndpoint: "https://example.com",
		Wallets:     []configPkg.Wallet{{Address: "address"}},
	}}}

	tracer := tracing.InitNoopTracer()
	logger := loggerPkg.GetNopLogger()
	querier := NewBalanceQuerier(config, *logger, tracer)

	metrics, queries := querier.GetMetrics(context.Background())
	assert.Len(t, queries, 1)
	assert.False(t, queries[0].Success)

	assert.Len(t, metrics, 1)
	assert.Zero(t, testutil.CollectAndCount(metrics[0]))
}

//nolint:paralleltest // disabled due to httpmock usage
func TestBalanceQuerierOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/cosmos/bank/v1beta1/balances/address",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("balance.json")),
	)

	config := &configPkg.Config{Chains: []configPkg.Chain{{
		Name:        "chain",
		LCDEndpoint: "https://example.com",
		Wallets: []configPkg.Wallet{{
			Address: "address",
			Name:    "name",
			Group:   "group",
		}},
		Denoms: []configPkg.DenomInfo{{Denom: "uatom", DisplayDenom: "atom", DenomExponent: 6}},
	}}}

	tracer := tracing.InitNoopTracer()
	logger := loggerPkg.GetNopLogger()
	querier := NewBalanceQuerier(config, *logger, tracer)

	metrics, queries := querier.GetMetrics(context.Background())
	assert.Len(t, queries, 1)
	assert.True(t, queries[0].Success)
	assert.Len(t, metrics, 1)

	balance, ok := metrics[0].(*prometheus.GaugeVec)
	assert.True(t, ok)

	assert.InDelta(t, 2, testutil.CollectAndCount(balance), 0.001)
	assert.InDelta(t, 0.123456, testutil.ToFloat64(balance.With(prometheus.Labels{
		"chain":   "chain",
		"denom":   "atom",
		"address": "address",
		"name":    "name",
		"group":   "group",
	})), 0.01)
	assert.InDelta(t, 234567, testutil.ToFloat64(balance.With(prometheus.Labels{
		"chain":   "chain",
		"denom":   "ustake",
		"address": "address",
		"name":    "name",
		"group":   "group",
	})), 0.01)
}
