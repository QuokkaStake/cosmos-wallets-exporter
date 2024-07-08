package queriers

import (
	"context"
	"main/pkg/tracing"
	"testing"

	"github.com/prometheus/client_golang/prometheus"
	"github.com/prometheus/client_golang/prometheus/testutil"

	"github.com/stretchr/testify/assert"
)

func TestUptimeQuerier(t *testing.T) {
	t.Parallel()

	querier := NewUptimeQuerier(tracing.InitNoopTracer())
	metrics, queryInfos := querier.GetMetrics(context.Background())
	assert.Empty(t, queryInfos)
	assert.NotEmpty(t, metrics)

	gauge, ok := metrics[0].(*prometheus.GaugeVec)
	assert.True(t, ok)
	assert.NotEmpty(t, testutil.ToFloat64(gauge))
}
