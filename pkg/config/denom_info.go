package config

type DenomInfo struct {
	Denom             string `toml:"denom"`
	DisplayDenom      string `toml:"display-denom"`
	DenomExponent     int    `default:"6"               toml:"denom-exponent"`
	CoingeckoCurrency string `toml:"coingecko-currency"`
}

func (d DenomInfo) GetName() string {
	if d.DisplayDenom != "" {
		return d.DisplayDenom
	}

	return d.Denom
}
