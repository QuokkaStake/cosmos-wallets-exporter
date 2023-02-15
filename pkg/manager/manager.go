package manager

import (
	"main/pkg/coingecko"
	"main/pkg/config"
	"main/pkg/tendermint"
	"main/pkg/types"
	"main/pkg/utils"
	"sync"
	"time"

	"github.com/rs/zerolog"
)

type Manager struct {
	Config    *config.Config
	Coingecko *coingecko.Coingecko
	Logger    zerolog.Logger
}

func NewManager(config *config.Config, logger *zerolog.Logger) *Manager {
	return &Manager{
		Config:    config,
		Coingecko: coingecko.NewCoingecko(logger),
		Logger:    logger.With().Str("component", "manager").Logger(),
	}
}

func (m *Manager) GetAllBalances() []types.WalletBalanceEntry {
	currenciesList := m.Config.GetCoingeckoCurrencies()
	currenciesRates := m.Coingecko.FetchPrices(currenciesList)

	length := 0
	for _, chain := range m.Config.Chains {
		for range chain.Wallets {
			length++
		}
	}

	balances := make([]types.WalletBalanceEntry, length)

	var wg sync.WaitGroup
	wg.Add(length)

	index := 0

	for _, chain := range m.Config.Chains {
		rpc := tendermint.NewRPC(chain.LCDEndpoint, m.Logger)

		for _, wallet := range chain.Wallets {
			go func(wallet config.Wallet, chain config.Chain, index int) {
				defer wg.Done()

				start := time.Now()

				balanceToAdd := types.WalletBalanceEntry{
					Chain:  chain.Name,
					Wallet: wallet,
				}

				balance, err := rpc.GetWalletBalances(wallet.Address)
				if err != nil {
					m.Logger.Error().
						Err(err).
						Str("chain", chain.Name).
						Str("wallet", wallet.Address).
						Msg("Error querying balance")
					balanceToAdd.Success = false
				} else {
					balanceToAdd.Success = true
					balanceToAdd.Balances = balance.Balances
					balanceToAdd.UsdPrice = m.MaybeGetUsdPrice(chain, balance.Balances, currenciesRates)
				}

				balanceToAdd.Duration = time.Since(start)

				balances[index] = balanceToAdd
			}(wallet, chain, index)

			index++
		}
	}

	wg.Wait()

	return balances
}

func (m *Manager) MaybeGetUsdPrice(
	chain config.Chain,
	balances types.Balances,
	rates map[string]float64,
) float64 {
	price, hasPrice := rates[chain.CoingeckoCurrency]
	if !hasPrice {
		return 0
	}

	var usdPriceTotal float64 = 0
	for _, balance := range balances {
		if balance.Denom == chain.BaseDenom {
			usdPriceTotal += utils.StrToFloat64(balance.Amount) * price / float64(chain.DenomCoefficient)
		}
	}

	return usdPriceTotal
}
