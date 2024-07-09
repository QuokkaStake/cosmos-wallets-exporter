package config

import (
	"errors"
	"fmt"
	"main/pkg/fs"

	"github.com/BurntSushi/toml"
	"github.com/creasty/defaults"
)

type Config struct {
	TracingConfig TracingConfig `toml:"tracing"`
	LogConfig     LogConfig     `toml:"log"`
	ListenAddress string        `default:":9550" toml:"listen-address"`
	Chains        []Chain       `toml:"chains"`
}

func (c *Config) Validate() error {
	if len(c.Chains) == 0 {
		return errors.New("no chains provided")
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

func GetConfig(path string, filesystem fs.FS) (*Config, error) {
	configBytes, err := filesystem.ReadFile(path)
	if err != nil {
		return nil, err
	}

	configString := string(configBytes)

	configStruct := Config{}
	if _, err = toml.Decode(configString, &configStruct); err != nil {
		return nil, err
	}

	defaults.MustSet(&configStruct)
	return &configStruct, nil
}
