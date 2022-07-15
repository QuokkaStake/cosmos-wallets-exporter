package main

import (
	"encoding/json"
	"fmt"
	"net/http"
	"time"

	"github.com/rs/zerolog"
)

type RPC struct {
	URL    string
	Logger zerolog.Logger
}

func NewRPC(url string, logger zerolog.Logger) *RPC {
	return &RPC{
		URL:    url,
		Logger: logger.With().Str("component", "rpc").Logger(),
	}
}

func (rpc *RPC) GetWalletBalances(address string) (*BalanceResponse, error) {
	url := fmt.Sprintf(
		"%s/cosmos/bank/v1beta1/balances/%s",
		rpc.URL,
		address,
	)

	var response *BalanceResponse
	if err := rpc.Get(url, &response); err != nil {
		return nil, err
	}

	return response, nil
}

func (rpc *RPC) Get(url string, target interface{}) error {
	client := &http.Client{Timeout: 10 * 1000000000}
	start := time.Now()

	req, err := http.NewRequest("GET", url, nil)
	if err != nil {
		return err
	}

	rpc.Logger.Debug().Str("url", url).Msg("Doing a query...")

	res, err := client.Do(req)
	if err != nil {
		rpc.Logger.Warn().Str("url", url).Err(err).Msg("Query failed")
		return err
	}
	defer res.Body.Close()

	rpc.Logger.Debug().Str("url", url).Dur("duration", time.Since(start)).Msg("Query is finished")

	return json.NewDecoder(res.Body).Decode(target)
}
