package types

import (
	"sort"

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

func (p *Prices) FilterByBlock(blockNum uint64) *Prices {
	prices := Prices{}
	for _, price := range *p {
		if price.BlockNum == blockNum {
			prices = append(prices, price)
		}
	}
	return &prices
}

func (p *Prices) FilterByDenom(denom string) *Prices {
	prices := Prices{}
	for _, price := range *p {
		if price.ExchangeRateTuple.Denom == denom {
			prices = append(prices, price)
		}
	}
	return &prices
}

func (p *Prices) Sort() *Prices {
	prices := *p
	sort.Slice(
		prices,
		func(i, j int) bool {
			if prices[i].BlockNum == prices[j].BlockNum {
				return prices[i].ExchangeRateTuple.Denom < prices[j].ExchangeRateTuple.Denom
			}
			return prices[i].BlockNum < prices[j].BlockNum
		},
	)
	return &prices
}

func (p *Prices) NewestBlockNum() uint64 {
	blockNum := uint64(0)
	for _, price := range *p {
		if price.BlockNum > blockNum {
			blockNum = price.BlockNum
		}
	}
	return blockNum
}
