package config

import (
	"errors"
	"fmt"
)

type Chain struct {
	Name        string      `toml:"name"`
	LCDEndpoint string      `toml:"lcd-endpoint"`
	Denoms      []DenomInfo `toml:"denoms"`
	Wallets     []Wallet    `toml:"wallets"`
}

func (c *Chain) Validate() error {
	if c.Name == "" {
		return errors.New("empty chain name")
	}

	if c.LCDEndpoint == "" {
		return errors.New("no LCD endpoint provided")
	}

	if len(c.Wallets) == 0 {
		return errors.New("no wallets provided")
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
