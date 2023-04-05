package keeper

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	ltypes "github.com/umee-network/umee/v4/x/leverage/types"
)

type LeverageKeeper struct {
	tokenSettings map[string]ltypes.Token
}

func (k LeverageKeeper) GetTokenSettings(ctx sdk.Context, baseDenom string) (ltypes.Token, error) {
	ts, ok := k.tokenSettings[baseDenom]
	if !ok {
		return ts, errors.New("token settings not found")
	}
	return ts, nil
}
func (k LeverageKeeper) ExchangeUToken(ctx sdk.Context, uToken sdk.Coin) (sdk.Coin, error) {
	panic("not implemented")
}
func (k LeverageKeeper) DeriveExchangeRate(ctx sdk.Context, denom string) sdk.Dec {
	panic("not implemented")
}

func NewLeverageKeeperMock(denoms ...string) LeverageKeeper {
	tokenSettings := map[string]ltypes.Token{}
	for _, d := range denoms {
		tokenSettings[d] = ltypes.Token{
			BaseDenom:   d,
			SymbolDenom: d,
		}
	}
	return LeverageKeeper{tokenSettings: tokenSettings}
}

type Oracle struct {
	prices map[string]sdk.Dec
}

func (o Oracle) Price(ctx sdk.Context, denom string) (sdk.Dec, error) {
	p, ok := o.prices[denom]
	if !ok {
		return p, ltypes.ErrNotRegisteredToken.Wrap(denom)
	}
	return p, nil
}

func NewOracleMock(denom string, price sdk.Dec) Oracle {
	prices := map[string]sdk.Dec{}
	prices[denom] = price
	return Oracle{prices: prices}
}
