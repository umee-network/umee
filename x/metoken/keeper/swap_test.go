package keeper

import (
	"testing"

	"github.com/stretchr/testify/assert"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v6/x/metoken/mocks"
)

func TestSwap_Valid(t *testing.T) {
	assert := assert.New(t)
	k := initMeUSDNoopKeper(t)
	params := k.GetParams()

	indexBal1, err := k.IndexBalances(mocks.MeUSDDenom)
	assert.NoError(err)

	tenUSDT := sdk.NewCoin(mocks.USDTBaseDenom, sdkmath.NewInt(10_000000))
	resp, err := k.swap(sdk.AccAddress{}, mocks.MeUSDDenom, tenUSDT)
	assert.NoError(err)

	// delta_allocation = (current_allocation - target_allocation) / target_allocation
	// delta_allocation = (0.238679846938775510 - 0.33) / 0.33 = -0.276727736549165121
	// fee = balanced_fee + delta_allocation * balanced_fee
	// fee = 0.2 - 0.276727736549165121 * 0.2 = 0.144654452690167
	assert.Equal(resp.fee, sdk.NewCoin(tenUSDT.Denom, sdkmath.NewInt(1446544)))

	// reserved = (amount - fee) * reserve_portion
	// reserved = (10 - 1.446544) * 0.2 = 1.7106912
	assert.Equal(resp.reserved, sdk.NewCoin(tenUSDT.Denom, sdkmath.NewInt(1710691)))

	// leveraged = amount - fee - reserved
	// leveraged = 10 - 1.446544 - 1.710691 = 6.842765
	assert.Equal(resp.leveraged, sdk.NewCoin(tenUSDT.Denom, sdkmath.NewInt(6842765)))

	indexBal2, err := k.IndexBalances(mocks.MeUSDDenom)
	usdBal2, n := indexBal2.AssetBalance(tenUSDT.Denom)
	assert.GreaterOrEqual(n, 0)
	usdBal1, _ := indexBal1.AssetBalance(tenUSDT.Denom)
	toAuction := params.RewardsAuctionFeeFactor.Mul(resp.fee.Amount)
	assert.Equal(usdBal2.Fees.Sub(usdBal1.Fees), resp.fee.Amount.Sub(toAuction))

	// exchange_rate = coin_price / metoken_price
	// meTokens = (reserved + leveraged) * exchange_rate
	i, err := k.RegisteredIndex(mocks.MeUSDDenom)
	assert.NoError(err)
	p, err := k.Prices(i)
	assert.NoError(err)
	assert.Equal(
		resp.meTokens, sdk.NewCoin(
			mocks.MeUSDDenom, mocks.USDTPrice.Quo(p.Price).MulInt(
				resp.
					reserved.Amount.Add(resp.leveraged.Amount),
			).TruncateInt(),
		),
	)
}

func TestSwap_Errors(t *testing.T) {
	k := initMeUSDNoopKeper(t)

	_, err := k.swap(sdk.AccAddress{}, mocks.MeUSDDenom, sdk.NewCoin("nonexistingMetoken", sdkmath.OneInt()))
	assert.ErrorContains(t, err, "not found")

	moreMaxSupply := sdk.NewCoin(mocks.USDTBaseDenom, sdkmath.NewInt(10000000_000000))
	_, err = k.swap(sdk.AccAddress{}, mocks.MeUSDDenom, moreMaxSupply)
	assert.ErrorContains(t, err, "not possible to mint the desired amount")
}

func TestSwap_LeverageOversupplied(t *testing.T) {
	l := NewLeverageMock()
	ts := l.tokens[mocks.USDTBaseDenom]
	ts.MaxSupply = sdkmath.NewInt(10_000000)
	l.tokens[mocks.USDTBaseDenom] = ts
	k := initMeUSDKeeper(t, NewBankMock(), l, NewOracleMock())

	thirtyUSDT := sdk.NewCoin(mocks.USDTBaseDenom, sdkmath.NewInt(30_000000))
	resp, err := k.swap(sdk.AccAddress{}, mocks.MeUSDDenom, thirtyUSDT)
	assert.NoError(t, err)
	assert.True(t, resp.reserved.IsGTE(resp.leveraged))
}

func TestBreakFees(t *testing.T) {
	assert := assert.New(t)
	k := initMeUSDNoopKeper(t)
	p := k.GetParams()
	assert.EqualValues(p.RewardsAuctionFeeFactor, 1000)

	amount := sdkmath.NewInt(2_000)
	toAuction, toRevenue := k.breakFee(amount)
	assert.Equal(toAuction, sdkmath.NewInt(200))
	assert.Equal(toRevenue, sdkmath.NewInt(1_800))

	amount = sdkmath.NewInt(9)
	toAuction, toRevenue = k.breakFee(amount)
	assert.Equal(toAuction, sdkmath.NewInt(0))
	assert.Equal(toRevenue, amount)

	amount = sdkmath.NewInt(2_001)
	toAuction, toRevenue = k.breakFee(amount)
	assert.Equal(toAuction, sdkmath.NewInt(200))
	assert.Equal(toRevenue, sdkmath.NewInt(1_801))
}
