// Simple mocks for unit tests

package keeper

import (
	"errors"

	sdk "github.com/cosmos/cosmos-sdk/types"

	ltypes "github.com/umee-network/umee/v6/x/leverage/types"
)

type LeverageKeeper struct {
	tokenSettings map[string]ltypes.Token
}

func (k LeverageKeeper) GetTokenSettings(_ sdk.Context, baseDenom string) (ltypes.Token, error) {
	ts, ok := k.tokenSettings[baseDenom]
	if !ok {
		return ts, errors.New("token settings not found")
	}
	return ts, nil
}

func (k LeverageKeeper) ToToken(_ sdk.Context, _ sdk.Coin) (sdk.Coin, error) {
	panic("not implemented")
}

func (k LeverageKeeper) DeriveExchangeRate(_ sdk.Context, _ string) sdk.Dec {
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

func (o Oracle) Price(_ sdk.Context, denom string) (sdk.Dec, error) {
	p, ok := o.prices[denom]
	if !ok {
		// When token exists in leverage registry but price is not found we are returning `0`
		// https://github.com/umee-network/umee/tree/main/x/oracle/keeper/historic_avg.go#L126
		return sdk.ZeroDec(), nil
	}
	return p, nil
}

func NewOracleMock(denom string, price sdk.Dec) Oracle {
	prices := map[string]sdk.Dec{}
	prices[denom] = price
	return Oracle{prices: prices}
}
