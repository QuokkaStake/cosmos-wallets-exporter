package tendermint

import (
	"fmt"
	"main/pkg/config"
	"main/pkg/http"
	"main/pkg/types"
	"main/pkg/utils"
	"sync"

	"github.com/rs/zerolog"
)

type RPC struct {
	Client *http.Client
	URL    string
	Logger zerolog.Logger

	LastQueryHeight map[string]int64
	Mutex           sync.Mutex
}

func NewRPC(chain config.Chain, logger zerolog.Logger) *RPC {
	return &RPC{
		Client:          http.NewClient(logger, chain.Name),
		URL:             chain.LCDEndpoint,
		Logger:          logger.With().Str("component", "rpc").Logger(),
		LastQueryHeight: make(map[string]int64),
	}
}

func (rpc *RPC) GetWalletBalances(address string) (*types.BalanceResponse, types.QueryInfo, error) {
	lastHeight, _ := rpc.LastQueryHeight[address]

	url := fmt.Sprintf(
		"%s/cosmos/bank/v1beta1/balances/%s",
		rpc.URL,
		address,
	)

	var response *types.BalanceResponse
	queryInfo, header, err := rpc.Client.Get(url, &response, types.HTTPPredicateCheckHeightAfter(lastHeight))
	if err != nil {
		return nil, queryInfo, err
	}

	newLastHeight, err := utils.GetBlockHeightFromHeader(header)
	if err != nil {
		return nil, queryInfo, err
	}

	rpc.Mutex.Lock()
	rpc.LastQueryHeight[address] = newLastHeight
	rpc.Mutex.Unlock()

	return response, queryInfo, nil
}
