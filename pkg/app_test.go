package pkg

import (
	"io"
	"main/assets"
	"main/pkg/fs"
	"net/http"
	"testing"
	"time"

	"github.com/jarcoal/httpmock"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
)

//nolint:paralleltest // disabled
func TestAppLoadConfigError(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			require.Fail(t, "Expected to have a panic here!")
		}
	}()

	filesystem := &fs.TestFS{}

	app := NewApp(filesystem, "not-found-config.toml", "1.2.3")
	app.Start()
}

//nolint:paralleltest // disabled
func TestAppLoadConfigInvalid(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			require.Fail(t, "Expected to have a panic here!")
		}
	}()

	filesystem := &fs.TestFS{}

	app := NewApp(filesystem, "config-invalid.toml", "1.2.3")
	app.Start()
}

//nolint:paralleltest // disabled
func TestAppFailToStart(t *testing.T) {
	defer func() {
		if r := recover(); r == nil {
			require.Fail(t, "Expected to have a panic here!")
		}
	}()

	filesystem := &fs.TestFS{}

	app := NewApp(filesystem, "config-invalid-listen-address.toml", "1.2.3")
	app.Start()
}

//nolint:paralleltest // disabled
func TestAppStopOperation(t *testing.T) {
	filesystem := &fs.TestFS{}

	app := NewApp(filesystem, "config-valid.toml", "1.2.3")
	app.Stop()
	assert.True(t, true)
}

//nolint:paralleltest // disabled
func TestAppLoadConfigOk(t *testing.T) {
	filesystem := &fs.TestFS{}

	app := NewApp(filesystem, "config-valid.toml", "1.2.3")
	go app.Start()

	for {
		request, err := http.Get("http://localhost:9550/healthcheck")
		_ = request.Body.Close()
		if err == nil {
			break
		}

		time.Sleep(time.Millisecond * 100)
	}

	httpmock.Activate()
	defer httpmock.DeactivateAndReset()

	httpmock.RegisterResponder(
		"GET",
		"https://example.com/cosmos/bank/v1beta1/balances/address",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("balance.json")),
	)

	httpmock.RegisterResponder(
		"GET",
		"https://api.coingecko.com/api/v3/simple/price?ids=cosmos&vs_currencies=usd",
		httpmock.NewBytesResponder(200, assets.GetBytesOrPanic("coingecko.json")),
	)

	httpmock.RegisterResponder("GET", "http://localhost:9550/healthcheck", httpmock.InitialTransport.RoundTrip)
	httpmock.RegisterResponder("GET", "http://localhost:9550/metrics", httpmock.InitialTransport.RoundTrip)

	response, err := http.Get("http://localhost:9550/metrics")
	require.NoError(t, err)
	require.NotEmpty(t, response)

	body, err := io.ReadAll(response.Body)
	require.NoError(t, err)
	require.NotEmpty(t, body)

	err = response.Body.Close()
	require.NoError(t, err)
}
