package coingecko

import (
	"fmt"
	"main/pkg/config"
	"main/pkg/http"
	"main/pkg/types"
	"strings"

	"github.com/rs/zerolog"
)

type Response map[string]map[string]float64

type Coingecko struct {
	Client *http.Client
	Config *config.Config
	Logger zerolog.Logger
}

func NewCoingecko(appConfig *config.Config, logger zerolog.Logger) *Coingecko {
	return &Coingecko{
		Config: appConfig,
		Client: http.NewClient(logger, "coingecko"),
		Logger: logger.With().Str("component", "coingecko").Logger(),
	}
}

func (c *Coingecko) FetchPrices(currencies []string) (map[string]float64, types.QueryInfo) {
	ids := strings.Join(currencies, ",")
	url := fmt.Sprintf("https://api.coingecko.com/api/v3/simple/price?ids=%s&vs_currencies=usd", ids)

	var response Response
	queryInfo, err := c.Client.Get(url, &response)
	if err != nil {
		c.Logger.Error().Err(err).Msg("Could not get rate")
		return nil, queryInfo
	}

	prices := map[string]float64{}

	for currencyKey, currencyValue := range response {
		for _, baseCurrencyValue := range currencyValue {
			chain, found := c.Config.FindChainByCoingeckoCurrency(currencyKey)
			if !found {
				c.Logger.Warn().
					Str("currency", currencyKey).
					Msg("Could not find chain by coingecko currency, which should never happen.")
			} else {
				prices[chain.Name] = baseCurrencyValue
			}
		}
	}

	return prices, queryInfo
}
