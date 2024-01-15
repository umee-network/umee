package keeper

import (
	"testing"

	"github.com/stretchr/testify/assert"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v6/x/metoken/mocks"
)

func TestRedeem_Valid(t *testing.T) {
	k := initMeUSDKeeper(t, NewBankMock(), NewLeverageMock(), NewOracleMock())

	hundredUSDT := sdk.NewCoin(mocks.USDTBaseDenom, sdkmath.NewInt(100_000000))
	_, err := k.swap(sdk.AccAddress{}, mocks.MeUSDDenom, hundredUSDT)
	assert.NoError(t, err)

	fiftyMeUSD := sdk.NewCoin(mocks.MeUSDDenom, sdkmath.NewInt(50_000000))
	resp, err := k.redeem(sdk.AccAddress{}, fiftyMeUSD, mocks.USDTBaseDenom)
	assert.NoError(t, err)

	// exchange_rate = metoken_price / coin_price
	// coins = metokens * exchange_rate
	index, err := k.RegisteredIndex(mocks.MeUSDDenom)
	assert.NoError(t, err)
	p, err := k.Prices(index)
	assert.NoError(t, err)
	usdtPrice, err := p.PriceByBaseDenom(mocks.USDTBaseDenom)
	assert.NoError(t, err)
	exchangeRate := p.Price.Quo(usdtPrice.Price)
	coins := exchangeRate.MulInt(fiftyMeUSD.Amount).TruncateInt()

	// delta_allocation = (target_allocation - current_allocation) / target_allocation
	// delta_allocation = (0.33 - 0.251415376464572197) / 0.33 = 0.238135222834629706
	// fee = balanced_fee + delta_allocation * balanced_fee
	// fee = 0.2 + 0.238135222834629706 * 0.2 = 0.247627044566925941
	// fee_amount = fee * amount
	expectedFee := sdkmath.LegacyMustNewDecFromStr("0.247627044566925941").MulInt(coins).TruncateInt()
	assert.Equal(t, resp.fee, sdk.NewCoin(mocks.USDTBaseDenom, expectedFee))
	assert.Equal(t, coins, resp.fromLeverage.Amount.Add(resp.fromReserves.Amount))
}

func TestRedeem_LeverageUndersupplied(t *testing.T) {
	k := initMeUSDKeeper(t, NewBankMock(), NewLeverageMock(), NewOracleMock())

	hundredUSDT := sdk.NewCoin(mocks.ISTBaseDenom, sdkmath.NewInt(100_000000))
	_, err := k.swap(sdk.AccAddress{}, mocks.MeUSDDenom, hundredUSDT)
	assert.NoError(t, err)

	fifteenMeUSD := sdk.NewCoin(mocks.MeUSDDenom, sdkmath.NewInt(15_000000))
	resp, err := k.redeem(sdk.AccAddress{}, fifteenMeUSD, mocks.ISTBaseDenom)
	assert.NoError(t, err)

	assert.True(t, resp.fromLeverage.IsLT(resp.fromReserves))
}
