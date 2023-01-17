package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

type Prices []Price

func NewPrice(exchangeRate sdk.Dec, denom string, blockNum uint64) *Price {
	return &Price{
		ExchangeRateTuple: ExchangeRateTuple{
			ExchangeRate: exchangeRate,
			Denom:        denom,
		},
		BlockNum: blockNum,
	}
}

func (p *Prices) Decs() []sdk.Dec {
	decs := []sdk.Dec{}
	for _, price := range *p {
		decs = append(decs, price.ExchangeRateTuple.ExchangeRate)
	}
	return decs
}
