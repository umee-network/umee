package keeper

import (
	"testing"

	"github.com/stretchr/testify/assert"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v5/x/metoken/mocks"
)

func TestSwap(t *testing.T) {
	k := initMeUSDKeeper(t, NewBankMock(), NewLeverageMock(), NewOracleMock())

	tenUSDT := sdk.NewCoin(mocks.USDTBaseDenom, sdkmath.NewInt(10_000000))
	resp, err := k.swap(sdk.AccAddress{}, mocks.MeUSDDenom, tenUSDT)
	assert.NoError(t, err)

	// delta_allocation = (current_allocation - target_allocation) / target_allocation
	// delta_allocation = (0.238679846938775510 - 0.33) / 0.33 = -0.276727736549165121
	// fee = balanced_fee + delta_allocation * balanced_fee
	// fee = 0.2 - 0.276727736549165121 * 0.2 = 0.144654452690167
	// fee_amount = fee * amount = 0.144654452690167 * 10 = 1.44654452690167
	assert.Equal(t, resp.fee, sdk.NewCoin(tenUSDT.Denom, sdkmath.NewInt(1446544)))

	// reserved = (amount - fee) * reserve_portion
	// reserved = (10 - 1.446544) * 0.2 = 1.7106912
	assert.Equal(t, resp.reserved, sdk.NewCoin(tenUSDT.Denom, sdkmath.NewInt(1710691)))

	// leveraged = amount - fee - reserved
	// leveraged = 10 - 1.446544 - 1.710691 = 6.842765
	assert.Equal(t, resp.leveraged, sdk.NewCoin(tenUSDT.Denom, sdkmath.NewInt(6842765)))

	// exchange_rate = coin_price / metoken_price
	// meTokens = (reserved + leveraged) * exchange_rate
	i, err := k.RegisteredIndex(mocks.MeUSDDenom)
	assert.NoError(t, err)
	p, err := k.Prices(i)
	assert.NoError(t, err)
	meTokenPrice, err := p.Price(mocks.MeUSDDenom)
	assert.Equal(t, resp.meTokens, sdk.NewCoin(mocks.MeUSDDenom, mocks.USDTPrice.Quo(meTokenPrice.Price).MulInt(resp.
		reserved.Amount.Add(resp.leveraged.Amount)).TruncateInt()))
}
