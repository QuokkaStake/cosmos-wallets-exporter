package http

import (
	"errors"
	"main/assets"
	"main/pkg/constants"
	loggerPkg "main/pkg/logger"
	"main/pkg/tracing"
	"main/pkg/types"
	"net/http"
	"testing"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/require"
)

func TestHttpClientErrorCreating(t *testing.T) {
	t.Parallel()

	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := NewClient(*logger, "chain", tracer)
	queryInfo, _, err := client.Get("://test", nil, types.HTTPPredicateAlwaysPass(), nil)
	require.Error(t, err)
	require.False(t, queryInfo.Success)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestHttpClientQueryFail(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://example.com",
		httpmock.NewErrorResponder(errors.New("custom error")),
	)
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := NewClient(*logger, "chain", tracer)

	var response interface{}
	_, _, err := client.Get("https://example.com", &response, types.HTTPPredicateCheckHeightAfter(100), nil)
	require.Error(t, err)
	require.ErrorContains(t, err, "custom error")
}

//nolint:paralleltest // disabled due to httpmock usage
func TestHttpClientPredicateFail(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://example.com",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("error.json")).HeaderAdd(http.Header{
			constants.HeaderBlockHeight: []string{"1"},
		}),
	)
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := NewClient(*logger, "chain", tracer)
	queryInfo, _, err := client.Get("https://example.com", nil, types.HTTPPredicateCheckHeightAfter(100), nil)
	require.Error(t, err)
	require.False(t, queryInfo.Success)
}

//nolint:paralleltest // disabled due to httpmock usage
func TestHttpClientJsonParseFail(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://example.com",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("invalid-json.json")),
	)
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := NewClient(*logger, "chain", tracer)
	var response interface{}

	_, _, err := client.Get("https://example.com", &response, types.HTTPPredicateAlwaysPass(), nil)
	require.Error(t, err)
	require.ErrorContains(t, err, "invalid character")
}

//nolint:paralleltest // disabled due to httpmock usage
func TestHttpClientOk(t *testing.T) {
	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://example.com",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("error.json")),
	)
	logger := loggerPkg.GetNopLogger()
	tracer := tracing.InitNoopTracer()
	client := NewClient(*logger, "chain", tracer)

	var response interface{}
	_, _, err := client.Get("https://example.com", &response, types.HTTPPredicateAlwaysPass(), nil)
	require.NoError(t, err)
}
