package tendermint

import (
	"encoding/json"
	"fmt"
	"main/pkg/config"
	"main/pkg/types"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

type RPC struct {
	Chain  string
	URL    string
	Logger zerolog.Logger
}

func NewRPC(chain config.Chain, logger zerolog.Logger) *RPC {
	return &RPC{
		Chain:  chain.Name,
		URL:    chain.LCDEndpoint,
		Logger: logger.With().Str("component", "rpc").Logger(),
	}
}

func (rpc *RPC) GetWalletBalances(address string) (*types.BalanceResponse, types.QueryInfo, error) {
	url := fmt.Sprintf(
		"%s/cosmos/bank/v1beta1/balances/%s",
		rpc.URL,
		address,
	)

	var response *types.BalanceResponse
	queryInfo, err := rpc.Get(url, &response)
	if err != nil {
		return nil, queryInfo, err
	}

	return response, queryInfo, nil
}

func (rpc *RPC) Get(url string, target interface{}) (types.QueryInfo, error) {
	client := &http.Client{Timeout: 10 * 1000000000}
	start := time.Now()

	queryInfo := types.QueryInfo{
		Success: false,
		Chain:   rpc.Chain,
		URL:     url,
	}

	req, err := http.NewRequest(http.MethodGet, url, nil)
	if err != nil {
		return queryInfo, err
	}

	req.Header.Set("User-Agent", "cosmos-wallets-exporter")

	rpc.Logger.Debug().Str("url", url).Msg("Doing a query...")

	res, err := client.Do(req)
	queryInfo.Duration = time.Since(start)
	if err != nil {
		rpc.Logger.Warn().Str("url", url).Err(err).Msg("Query failed")
		return queryInfo, err
	}
	defer res.Body.Close()

	rpc.Logger.Debug().Str("url", url).Dur("duration", time.Since(start)).Msg("Query is finished")

	err = json.NewDecoder(res.Body).Decode(target)
	queryInfo.Success = err == nil

	return queryInfo, err
}
