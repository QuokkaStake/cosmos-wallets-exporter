package main

import (
	"sync"

	"github.com/rs/zerolog"
)

type Manager struct {
	Config Config
	Logger zerolog.Logger
}

func NewManager(config Config, logger *zerolog.Logger) *Manager {
	return &Manager{
		Config: config,
		Logger: logger.With().Str("component", "manager").Logger(),
	}
}

func (m *Manager) GetAllBalances() []WalletBalanceEntry {
	len := 0
	for _, chain := range m.Config.Chains {
		for _, _ = range chain.Wallets {
			len++
		}
	}

	balances := make([]WalletBalanceEntry, len)

	var wg sync.WaitGroup
	wg.Add(len)

	index := 0

	for _, chain := range m.Config.Chains {
		rpc := NewRPC(chain.LCDEndpoint, m.Logger)

		for _, wallet := range chain.Wallets {
			go func(wallet Wallet, chain Chain, index int) {
				defer wg.Done()

				balanceToAdd := WalletBalanceEntry{
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
