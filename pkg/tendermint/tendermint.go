package tendermint

import (
	"fmt"
	"main/pkg/config"
	"main/pkg/http"
	"main/pkg/types"

	"github.com/rs/zerolog"
)

type RPC struct {
	Client *http.Client
	URL    string
	Logger zerolog.Logger
}

func NewRPC(chain config.Chain, logger zerolog.Logger) *RPC {
	return &RPC{
		Client: http.NewClient(logger, chain.Name),
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
	queryInfo, err := rpc.Client.Get(url, &response)
	if err != nil {
		return nil, queryInfo, err
	}

	return response, queryInfo, nil
}
