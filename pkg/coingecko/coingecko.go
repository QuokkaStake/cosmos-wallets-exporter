package coingecko

import (
	"main/pkg/config"

	"github.com/rs/zerolog"
	gecko "github.com/superoo7/go-gecko/v3"
)

type Coingecko struct {
	Client *gecko.Client
	Config *config.Config
	Logger zerolog.Logger
}

func NewCoingecko(appConfig *config.Config, logger *zerolog.Logger) *Coingecko {
	return &Coingecko{
		Config: appConfig,
		Client: gecko.NewClient(nil),
		Logger: logger.With().Str("component", "coingecko").Logger(),
	}
}

func (c *Coingecko) FetchPrices(currencies []string) map[string]float64 {
	result, err := c.Client.SimplePrice(currencies, []string{"USD"})
	if err != nil {
		c.Logger.Error().Err(err).Msg("Could not get rate")
		return map[string]float64{}
	}

	prices := map[string]float64{}

	for currencyKey, currencyValue := range *result {
		for _, baseCurrencyValue := range currencyValue {
			chain, found := c.Config.FindChainByCoingeckoCurrency(currencyKey)
			if !found {
				c.Logger.Warn().
					Str("currency", currencyKey).
					Msg("Could not find chain by coingecko currency, which should never happen.")
			} else {
				prices[chain.Name] = float64(baseCurrencyValue)
			}
		}
	}

	return prices
}
