package config

import (
	"fmt"
	"os"

	"github.com/BurntSushi/toml"
	"github.com/mcuadros/go-defaults"
)

type Wallet struct {
	Address string `toml:"address"`
	Name    string `toml:"name"`
	Group   string `toml:"group"`
}

type DenomInfo struct {
	Denom             string `toml:"denom"`
	DisplayDenom      string `toml:"display-denom"`
	DenomCoefficient  int64  `toml:"denom-coefficient" default:"1000000"`
	CoingeckoCurrency string `toml:"coingecko-currency"`
}

func (d DenomInfo) GetName() string {
	if d.DisplayDenom != "" {
		return d.DisplayDenom
	}

	return d.Denom
}

type Chain struct {
	Name        string      `toml:"name"`
	LCDEndpoint string      `toml:"lcd-endpoint"`
	Denoms      []DenomInfo `toml:"denoms"`
	Wallets     []Wallet    `toml:"wallets"`
}

func (w Wallet) Validate() error {
	if w.Address == "" {
		return fmt.Errorf("address for wallet is not specified")
	}

	return nil
}

func (c *Chain) Validate() error {
	if c.Name == "" {
		return fmt.Errorf("empty chain name")
	}

	if c.LCDEndpoint == "" {
		return fmt.Errorf("no LCD endpoint provided")
	}

	if len(c.Wallets) == 0 {
		return fmt.Errorf("no wallets provided")
	}

	for index, wallet := range c.Wallets {
		if err := wallet.Validate(); err != nil {
			return fmt.Errorf("error in wallet %d: %s", index, err)
		}
	}

	return nil
}

func (c *Chain) FindDenomByName(denom string) (*DenomInfo, bool) {
	for _, denomIterated := range c.Denoms {
		if denomIterated.Denom == denom {
			return &denomIterated, true
		}
	}

	return nil, false
}

type Config struct {
	LogConfig     LogConfig `toml:"log"`
	ListenAddress string    `toml:"listen-address" default:":9550"`
	Chains        []Chain   `toml:"chains"`
}

type LogConfig struct {
	LogLevel   string `toml:"level" default:"info"`
	JSONOutput bool   `toml:"json" default:"false"`
}

func (c *Config) Validate() error {
	if len(c.Chains) == 0 {
		return fmt.Errorf("no chains provided")
	}

	for index, chain := range c.Chains {
		if err := chain.Validate(); err != nil {
			return fmt.Errorf("error in chain %d: %s", index, err)
		}
	}

	return nil
}

func (c *Config) GetCoingeckoCurrencies() []string {
	currencies := []string{}

	for _, chain := range c.Chains {
		for _, denom := range chain.Denoms {
			if denom.CoingeckoCurrency != "" {
				currencies = append(currencies, denom.CoingeckoCurrency)
			}
		}
	}

	return currencies
}

func (c *Config) FindChainAndDenomByCoingeckoCurrency(currency string) (string, string, bool) {
	for _, chain := range c.Chains {
		for _, denom := range chain.Denoms {
			if denom.CoingeckoCurrency == currency {
				return chain.Name, denom.GetName(), true
			}
		}
	}

	return "", "", false
}

func GetConfig(path string) (*Config, error) {
	configBytes, err := os.ReadFile(path)
	if err != nil {
		return nil, err
	}

	configString := string(configBytes)

	configStruct := Config{}
	if _, err = toml.Decode(configString, &configStruct); err != nil {
		return nil, err
	}

	defaults.SetDefaults(&configStruct)
	return &configStruct, nil
}
