package queriers

import (
	configPkg "main/pkg/config"
	"main/pkg/types"
	"testing"
	"time"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"
	"github.com/stretchr/testify/assert"
)

func TestQueriesQuerier(t *testing.T) {
	t.Parallel()

	config := &configPkg.Config{Chains: []configPkg.Chain{{Name: "chain"}, {Name: "chain2"}}}

	queries := []types.QueryInfo{
		{Chain: "chain", Success: true, URL: "url1", Duration: 5 * time.Second},
		{Chain: "chain", Success: false, URL: "url2", Duration: 3 * time.Second},
	}

	querier := NewQueriesQuerier(config, queries)
	metrics, queryInfos := querier.GetMetrics()
	assert.Empty(t, queryInfos)
	assert.Len(t, metrics, 3)

	successGauge, ok := metrics[0].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.Equal(t, 2, testutil.CollectAndCount(successGauge))
	assert.InEpsilon(t, float64(1), testutil.ToFloat64(successGauge.With(prometheus.Labels{
		"chain": "chain",
	})), 0.01)
	assert.Zero(t, testutil.ToFloat64(successGauge.With(prometheus.Labels{
		"chain": "chain2",
	})))

	errorGauge, ok := metrics[1].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.Equal(t, 2, testutil.CollectAndCount(errorGauge))
	assert.InEpsilon(t, float64(1), testutil.ToFloat64(errorGauge.With(prometheus.Labels{
		"chain": "chain",
	})), 0.01)
	assert.Zero(t, testutil.ToFloat64(errorGauge.With(prometheus.Labels{
		"chain": "chain2",
	})))

	timingsGauge, ok := metrics[2].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.Equal(t, 2, testutil.CollectAndCount(timingsGauge))
	assert.InEpsilon(t, float64(5), testutil.ToFloat64(timingsGauge.With(prometheus.Labels{
		"chain": "chain",
		"url":   "url1",
	})), 0.01)
	assert.InEpsilon(t, float64(3), testutil.ToFloat64(timingsGauge.With(prometheus.Labels{
		"chain": "chain",
		"url":   "url2",
	})), 0.01)
}
