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

	_, _, err = k.swapFee(index, prices, sdk.NewCoin("inexistingAsset", sdkmath.ZeroInt()))
	require.ErrorIs(t, err, sdkerrors.ErrNotFound)

	// target_allocation = 0 -> fee = max_fee * coin_amount
	// fee = 0.5 * 10 = 5
	usdtAsset, i := index.AcceptedAsset(mocks.USDTBaseDenom)
	require.True(t, i >= 0)

	usdtAsset.TargetAllocation = sdkmath.LegacyZeroDec()
	index.SetAcceptedAsset(usdtAsset)
	tenUSDT := sdk.NewCoin(mocks.USDTBaseDenom, sdkmath.NewInt(10_000000))

	feeFraction, feeAmount, err := k.swapFee(index, prices, tenUSDT)
	require.NoError(t, err)
	require.True(t, feeAmount.Amount.Equal(sdkmath.NewInt(5_000000)))
	require.True(t, feeFraction.Equal(sdkmath.LegacyMustNewDecFromStr("0.5")))

	// swap_fee = balanced_fee + delta_allocation * balanced_fee
	// swap_fee = 0.2 + (-0.276727736549164797) * 0.2 = 0.144654452690166976
	// fee = swap_fee * coin_amount
	// fee = 0.144654452690166976 * 10 = 1.44654452690166976
	usdtAsset.TargetAllocation = sdkmath.LegacyMustNewDecFromStr("0.33")
	index.SetAcceptedAsset(usdtAsset)

	feeFraction, feeAmount, err = k.swapFee(index, prices, tenUSDT)
	require.NoError(t, err)
	require.True(t, feeAmount.Amount.Equal(sdkmath.NewInt(1_446544)))
	require.True(t, feeFraction.Equal(sdkmath.LegacyMustNewDecFromStr("0.144654452690166976")))
}

func TestRedeemFee(t *testing.T) {
	k := initMeUSDKeeper(t, nil, nil, nil)

	index, err := k.RegisteredIndex(mocks.MeUSDDenom)
	require.NoError(t, err)

	balance, err := k.IndexBalances(mocks.MeUSDDenom)
	require.NoError(t, err)
	prices := meUSDIndexPricesAdjustedToBalance(t, balance)

	_, _, err = k.redeemFee(index, prices, sdk.NewCoin("inexistingAsset", sdkmath.ZeroInt()))
	require.ErrorIs(t, err, sdkerrors.ErrNotFound)

	// target_allocation = 0 -> fee = min_fee * coin_amount
	// fee = 0.01 * 10 = 0.1
	usdtAsset, i := index.AcceptedAsset(mocks.USDTBaseDenom)
	require.True(t, i >= 0)

	usdtAsset.TargetAllocation = sdkmath.LegacyZeroDec()
	index.SetAcceptedAsset(usdtAsset)
	tenUSDT := sdk.NewCoin(mocks.USDTBaseDenom, sdkmath.NewInt(10_000000))

	feeFraction, feeAmount, err := k.redeemFee(index, prices, tenUSDT)
	require.NoError(t, err)
	require.True(t, feeAmount.Amount.Equal(sdkmath.NewInt(100000)))
	require.True(t, feeFraction.Equal(sdkmath.LegacyMustNewDecFromStr("0.01")))

	// redeem_fee = balanced_fee + delta_allocation * balanced_fee
	// redeem_fee = 0.2 + (0.276727736549164797) * 0.2 = 0.255345547309833024
	// fee = redeem_fee * coin_amount
	// fee = 0.255345547309833024 * 10 = 2.55345547309833024
	usdtAsset.TargetAllocation = sdkmath.LegacyMustNewDecFromStr("0.33")
	index.SetAcceptedAsset(usdtAsset)

	feeFraction, feeAmount, err = k.redeemFee(index, prices, tenUSDT)
	require.NoError(t, err)
	require.True(t, feeAmount.Amount.Equal(sdkmath.NewInt(2_553455)))
	require.True(t, feeFraction.Equal(sdkmath.LegacyMustNewDecFromStr("0.255345547309833024")))
}
