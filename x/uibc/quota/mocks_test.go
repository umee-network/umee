// Simple mocks for unit tests

package quota

import (
	"errors"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v6/util/genmap"
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

func (k LeverageKeeper) GetAllRegisteredTokens(_ sdk.Context) []ltypes.Token {
	return genmap.MapValues(k.tokenSettings)
}

func (k LeverageKeeper) ToToken(_ sdk.Context, _ sdk.Coin) (sdk.Coin, error) {
	panic("not implemented")
}

func (k LeverageKeeper) DeriveExchangeRate(_ sdk.Context, _ string) sdkmath.LegacyDec {
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
	prices map[string]sdkmath.LegacyDec
}

func (o Oracle) Price(_ sdk.Context, denom string) (sdkmath.LegacyDec, error) {
	p, ok := o.prices[denom]
	if !ok {
		// When token exists in leverage registry but price is not found we are returning `0`
		// https: //github.com/umee-network/umee/blob/v6.1.0/x/oracle/keeper/historic_avg.go#L126
		return sdkmath.LegacyZeroDec(), nil
	}
	return p, nil
}

func NewOracleMock(denom string, price sdkmath.LegacyDec) Oracle {
	prices := map[string]sdkmath.LegacyDec{}
	prices[denom] = price
	return Oracle{prices: prices}
}
