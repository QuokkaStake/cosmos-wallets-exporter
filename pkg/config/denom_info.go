package config

type DenomInfo struct {
	Denom             string `toml:"denom"`
	DisplayDenom      string `toml:"display-denom"`
	DenomCoefficient  int64  `default:"1000000"         toml:"denom-coefficient"`
	CoingeckoCurrency string `toml:"coingecko-currency"`
}

func (d DenomInfo) GetName() string {
	if d.DisplayDenom != "" {
		return d.DisplayDenom
	}

	return d.Denom
}
