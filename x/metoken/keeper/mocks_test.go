package keeper

import (
	sdkmath "cosmossdk.io/math"
	"github.com/umee-network/umee/v5/x/metoken/mocks"

	sdk "github.com/cosmos/cosmos-sdk/types"
	ltypes "github.com/umee-network/umee/v5/x/leverage/types"
	otypes "github.com/umee-network/umee/v5/x/oracle/types"
)

type Oracle struct {
	prices otypes.Prices
}

func (o Oracle) AllMedianPrices(_ sdk.Context) otypes.Prices {
	return o.prices
}

func NewOracleMock() Oracle {
	return Oracle{prices: mocks.ValidPrices()}
}

type Leverage struct {
	tokens map[string]ltypes.Token
}

func (l Leverage) GetTokenSettings(_ sdk.Context, denom string) (ltypes.Token, error) {
	ts, ok := l.tokens[denom]
	if !ok {
		return ts, ltypes.ErrNotRegisteredToken.Wrap(denom)
	}
	return ts, nil
}

func (l Leverage) ToUToken(_ sdk.Context, _ sdk.Coin) (sdk.Coin, error) {
	panic("not implemented")
}

func (l Leverage) ToToken(_ sdk.Context, _ sdk.Coin) (sdk.Coin, error) {
	panic("not implemented")
}

func (l Leverage) SupplyFromModule(_ sdk.Context, _ string, _ sdk.Coin) (sdk.Coin, bool, error) {
	panic("not implemented")
}

func (l Leverage) WithdrawToModule(_ sdk.Context, _ string, _ sdk.Coin) (sdk.Coin, bool, error) {
	panic("not implemented")
}

func (l Leverage) ModuleMaxWithdraw(_ sdk.Context, _ sdk.Coin) (sdkmath.Int, error) {
	panic("not implemented")
}

func (l Leverage) GetTotalSupply(_ sdk.Context, _ string) (sdk.Coin, error) {
	panic("not implemented")
}

func (l Leverage) GetAllSupplied(_ sdk.Context, _ sdk.AccAddress) (sdk.Coins, error) {
	panic("not implemented")
}

func NewLeverageMock() Leverage {
	return Leverage{
		tokens: map[string]ltypes.Token{
			mocks.USDTBaseDenom: mocks.ValidToken(mocks.USDTBaseDenom, mocks.USDTSymbolDenom, 6),
			mocks.USDCBaseDenom: mocks.ValidToken(mocks.USDCBaseDenom, mocks.USDCSymbolDenom, 6),
			mocks.ISTBaseDenom:  mocks.ValidToken(mocks.ISTBaseDenom, mocks.ISTSymbolDenom, 6),
			mocks.WBTCBaseDenom: mocks.ValidToken(mocks.WBTCBaseDenom, mocks.WBTCSymbolDenom, 6),
			mocks.ETHBaseDenom:  mocks.ValidToken(mocks.ETHBaseDenom, mocks.ETHSymbolDenom, 6),
			mocks.TestDenom1:    mocks.ValidToken(mocks.TestDenom1, mocks.TestDenom1, 6),
			mocks.CMSTBaseDenom: mocks.ValidToken(mocks.CMSTBaseDenom, mocks.CMSTSymbolDenom, 6),
		},
	}
}
