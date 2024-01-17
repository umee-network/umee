package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/umee-network/umee/v6/x/metoken"
)

// swapFee to be charged to the user, given a specific Index configuration and asset amount.
// It returns fee in fraction, fee amount and error.
// The fee in fraction represents the percentage of the fee, while the fee amount is the actual
// fee applied to the asset amount.
func (k Keeper) swapFee(index metoken.Index, indexPrices metoken.IndexPrices, asset sdk.Coin) (
	sdkmath.LegacyDec,
	sdk.Coin,
	error,
) {
	assetSettings, i := index.AcceptedAsset(asset.Denom)
	if i < 0 {
		return sdkmath.LegacyDec{}, sdk.Coin{}, sdkerrors.ErrNotFound.
			Wrapf("asset %s is not accepted in the index", asset.Denom)
	}

	// charge max fee if we don't want the token in the index.
	if assetSettings.TargetAllocation.IsZero() {
		return index.Fee.MaxFee, sdk.NewCoin(asset.Denom, index.Fee.MaxFee.MulInt(asset.Amount).TruncateInt()), nil
	}

	currentAllocation, err := k.currentAllocation(index, indexPrices, asset.Denom)
	if err != nil {
		return sdkmath.LegacyDec{}, sdk.Coin{}, err
	}

	// when current_allocation is zero, we incentivize the swap by charging only min_fee
	if currentAllocation.IsZero() {
		return index.Fee.MinFee, sdk.NewCoin(asset.Denom, index.Fee.MinFee.MulInt(asset.Amount).TruncateInt()), nil
	}

	allocationDeviation := currentAllocation.Sub(assetSettings.TargetAllocation).Quo(assetSettings.TargetAllocation)
	fee := index.Fee.CalculateFee(allocationDeviation)
	return fee, sdk.NewCoin(asset.Denom, fee.MulInt(asset.Amount).TruncateInt()), nil
}

// redeemFee to be charged to the user, given a specific Index configuration and asset amount.
// It returns fee in fraction, fee amount and error.
// The fee in fraction indicates the fee percentage, while the fee amount is the computed
// fee based on the asset amount being redeemed.
func (k Keeper) redeemFee(index metoken.Index, indexPrices metoken.IndexPrices, asset sdk.Coin) (
	sdkmath.LegacyDec,
	sdk.Coin,
	error,
) {
	assetSettings, i := index.AcceptedAsset(asset.Denom)
	if i < 0 {
		return sdkmath.LegacyDec{}, sdk.Coin{}, sdkerrors.ErrNotFound.
			Wrapf("asset %s is not accepted in the index", asset.Denom)
	}

	// charge min fee if we don't want the token in the index.
	if assetSettings.TargetAllocation.IsZero() {
		return index.Fee.MinFee, sdk.NewCoin(asset.Denom, index.Fee.MinFee.MulInt(asset.Amount).TruncateInt()), nil
	}

	allocationDeviation, err := k.redeemAllocationDeviation(
		index,
		indexPrices,
		asset.Denom,
		assetSettings.TargetAllocation,
	)
	if err != nil {
		return sdkmath.LegacyDec{}, sdk.Coin{}, err
	}

	fee := index.Fee.CalculateFee(allocationDeviation)
	return fee, sdk.NewCoin(asset.Denom, fee.MulInt(asset.Amount).TruncateInt()), nil
}

// currentAllocation returns a factor of the assetDenom supply in the index based on the USD price value.
func (k Keeper) currentAllocation(
	index metoken.Index,
	indexPrices metoken.IndexPrices,
	assetDenom string,
) (sdkmath.LegacyDec, error) {
	balances, err := k.IndexBalances(index.Denom)
	if err != nil {
		return sdkmath.LegacyDec{}, err
	}

	balance, i := balances.AssetBalance(assetDenom)
	if i < 0 {
		return sdkmath.LegacyDec{}, sdkerrors.ErrNotFound.Wrapf("balance for denom %s not found", assetDenom)
	}

	// if asset wasn't supplied to the index yet, the allocation is zero
	if balance.AvailableSupply().IsZero() {
		return sdkmath.LegacyZeroDec(), nil
	}

	// if no meToken in balance, the allocation is zero
	if !balances.MetokenSupply.IsPositive() {
		return sdkmath.LegacyZeroDec(), nil
	}

	assetPrice, err := indexPrices.PriceByBaseDenom(assetDenom)
	if err != nil {
		return sdkmath.LegacyDec{}, err
	}
	assetUSD, err := valueInUSD(balance.AvailableSupply(), assetPrice.Price, assetPrice.Exponent)
	if err != nil {
		return sdkmath.LegacyDec{}, err
	}

	meTokenUSD, err := valueInUSD(balances.MetokenSupply.Amount, indexPrices.Price, indexPrices.Exponent)
	if err != nil {
		return sdkmath.LegacyDec{}, err
	}

	return assetUSD.Quo(meTokenUSD), nil
}

// redeemAllocationDeviation returns the delta between the target allocation and current allocation of an asset in an
// index for a redemption. It returns negative value when the asset is oversupplied.
func (k Keeper) redeemAllocationDeviation(
	index metoken.Index,
	indexPrices metoken.IndexPrices,
	assetDenom string,
	targetAllocation sdkmath.LegacyDec,
) (sdkmath.LegacyDec, error) {
	currentAllocation, err := k.currentAllocation(index, indexPrices, assetDenom)
	if err != nil {
		return sdkmath.LegacyDec{}, err
	}

	if currentAllocation.IsZero() {
		return targetAllocation, nil
	}

	return targetAllocation.Sub(currentAllocation).Quo(targetAllocation), nil
}
