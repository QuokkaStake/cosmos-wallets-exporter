package config

import (
	"errors"
)

func (w Wallet) Validate() error {
	if w.Address == "" {
		return errors.New("address for wallet is not specified")
	}

	return nil
}

type Wallet struct {
	Address string `toml:"address"`
	Name    string `toml:"name"`
	Group   string `toml:"group"`
}
