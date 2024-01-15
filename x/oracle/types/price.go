package types

import (
	"sort"

	sdkmath "cosmossdk.io/math"
)

type Prices []Price

func NewPrice(exchangeRate sdkmath.LegacyDec, denom string, blockNum uint64) Price {
	return Price{
		ExchangeRateTuple: ExchangeRateTuple{
			ExchangeRate: exchangeRate,
			Denom:        denom,
		},
		BlockNum: blockNum,
	}
}

func (p Prices) Decs() []sdkmath.LegacyDec {
	decs := []sdkmath.LegacyDec{}
	for _, price := range p {
		decs = append(decs, price.ExchangeRateTuple.ExchangeRate)
	}
	return decs
}

func (p Prices) FilterByBlock(blockNum uint64) Prices {
	prices := Prices{}
	for _, price := range p {
		if price.BlockNum == blockNum {
			prices = append(prices, price)
		}
	}
	return prices
}

func (p Prices) FilterByDenom(denom string) Prices {
	prices := Prices{}
	for _, price := range p {
		if price.ExchangeRateTuple.Denom == denom {
			prices = append(prices, price)
		}
	}
	return prices
}

func (p Prices) Sort() Prices {
	sort.Slice(
		p,
		func(i, j int) bool {
			if p[i].BlockNum == p[j].BlockNum {
				return p[i].ExchangeRateTuple.Denom < p[j].ExchangeRateTuple.Denom
			}
			return p[i].BlockNum < p[j].BlockNum
		},
	)
	return p
}
