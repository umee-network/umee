package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/umee-network/umee/v6/x/metoken/mocks"
)

func TestSwapFee(t *testing.T) {
	k := initMeUSDKeeper(t, nil, nil, nil)

	index, err := k.RegisteredIndex(mocks.MeUSDDenom)
	require.NoError(t, err)

	balance, err := k.IndexBalances(mocks.MeUSDDenom)
	require.NoError(t, err)
	prices := meUSDIndexPricesAdjustedToBalance(t, balance)

	_, err = k.swapFee(index, prices, sdk.NewCoin("inexistingAsset", sdkmath.ZeroInt()))
	require.ErrorIs(t, err, sdkerrors.ErrNotFound)

	// target_allocation = 0 -> fee = max_fee * coin_amount
	// fee = 0.5 * 10 = 5
	i, usdtAsset := index.AcceptedAsset(mocks.USDTBaseDenom)
	require.True(t, i >= 0)

	usdtAsset.TargetAllocation = sdk.ZeroDec()
	index.SetAcceptedAsset(usdtAsset)
	tenUSDT := sdk.NewCoin(mocks.USDTBaseDenom, sdkmath.NewInt(10_000000))

	fee, err := k.swapFee(index, prices, tenUSDT)
	require.NoError(t, err)
	require.True(t, fee.Amount.Equal(sdkmath.NewInt(5_000000)))

	// swap_fee = balanced_fee + delta_allocation * balanced_fee
	// swap_fee = 0.2 + (-0.276727736549164797) * 0.2 = 0.144654452690167041
	// fee = swap_fee * coin_amount
	// fee = 0.144654452690167041 * 10 = 1.44654452690167041
	usdtAsset.TargetAllocation = sdk.MustNewDecFromStr("0.33")
	index.SetAcceptedAsset(usdtAsset)

	fee, err = k.swapFee(index, prices, tenUSDT)
	require.NoError(t, err)
	require.True(t, fee.Amount.Equal(sdkmath.NewInt(1_446544)))
}

func TestRedeemFee(t *testing.T) {
	k := initMeUSDKeeper(t, nil, nil, nil)

	index, err := k.RegisteredIndex(mocks.MeUSDDenom)
	require.NoError(t, err)

	balance, err := k.IndexBalances(mocks.MeUSDDenom)
	require.NoError(t, err)
	prices := meUSDIndexPricesAdjustedToBalance(t, balance)

	_, err = k.redeemFee(index, prices, sdk.NewCoin("inexistingAsset", sdkmath.ZeroInt()))
	require.ErrorIs(t, err, sdkerrors.ErrNotFound)

	// target_allocation = 0 -> fee = min_fee * coin_amount
	// fee = 0.01 * 10 = 0.1
	i, usdtAsset := index.AcceptedAsset(mocks.USDTBaseDenom)
	require.True(t, i >= 0)

	usdtAsset.TargetAllocation = sdk.ZeroDec()
	index.SetAcceptedAsset(usdtAsset)
	tenUSDT := sdk.NewCoin(mocks.USDTBaseDenom, sdkmath.NewInt(10_000000))

	fee, err := k.redeemFee(index, prices, tenUSDT)
	require.NoError(t, err)
	require.True(t, fee.Amount.Equal(sdkmath.NewInt(100000)))

	// redeem_fee = balanced_fee + delta_allocation * balanced_fee
	// redeem_fee = 0.2 + (0.276727736549164797) * 0.2 = 0.255345547309832959
	// fee = redeem_fee * coin_amount
	// fee = 0.255345547309832959 * 10 = 2.55345547309832959
	usdtAsset.TargetAllocation = sdk.MustNewDecFromStr("0.33")
	index.SetAcceptedAsset(usdtAsset)

	fee, err = k.redeemFee(index, prices, tenUSDT)
	require.NoError(t, err)
	require.True(t, fee.Amount.Equal(sdkmath.NewInt(2_553455)))
}
