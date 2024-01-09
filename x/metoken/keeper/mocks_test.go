package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v6/util/coin"
	ltypes "github.com/umee-network/umee/v6/x/leverage/types"
	"github.com/umee-network/umee/v6/x/metoken/mocks"
	otypes "github.com/umee-network/umee/v6/x/oracle/types"
)

type Oracle struct {
	prices otypes.Prices
}

func (o Oracle) AllMedianPrices(_ sdk.Context) otypes.Prices {
	return o.prices
}

func (o Oracle) SetExchangeRate(_ sdk.Context, _ string, _ sdk.Dec) {
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

func (l Leverage) ToUToken(_ sdk.Context, coin sdk.Coin) (sdk.Coin, error) {
	return coin, nil
}

func (l Leverage) ToToken(_ sdk.Context, coin sdk.Coin) (sdk.Coin, error) {
	return coin, nil
}

func (l Leverage) SupplyFromModule(_ sdk.Context, _ string, _ sdk.Coin) (sdk.Coin, bool, error) {
	return sdk.Coin{}, true, nil
}

func (l Leverage) WithdrawToModule(_ sdk.Context, _ string, coin sdk.Coin) (sdk.Coin, bool, error) {
	return coin, true, nil
}

func (l Leverage) ModuleMaxWithdraw(_ sdk.Context, coin sdk.Coin, _ sdk.AccAddress) (sdkmath.Int, error) {
	if coin.Denom == mocks.ISTBaseDenom {
		return sdk.MustNewDecFromStr("0.5").MulInt(coin.Amount).TruncateInt(), nil
	}
	return coin.Amount, nil
}

func (l Leverage) GetTotalSupply(_ sdk.Context, denom string) (sdk.Coin, error) {
	return coin.Zero(denom), nil
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

type Bank struct{}

func NewBankMock() Bank {
	return Bank{}
}

func (b Bank) MintCoins(_ sdk.Context, _ string, _ sdk.Coins) error {
	return nil
}

func (b Bank) BurnCoins(_ sdk.Context, _ string, _ sdk.Coins) error {
	return nil
}

func (b Bank) SendCoinsFromModuleToAccount(_ sdk.Context, _ string, _ sdk.AccAddress, _ sdk.Coins) error {
	return nil
}

func (b Bank) SendCoinsFromAccountToModule(_ sdk.Context, _ sdk.AccAddress, _ string, _ sdk.Coins) error {
	return nil
}
