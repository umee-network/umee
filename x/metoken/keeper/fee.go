package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	"github.com/umee-network/umee/v6/x/metoken"
)

// swapFee to be charged to the user, given a specific Index configuration and asset amount.
func (k Keeper) swapFee(index metoken.Index, indexPrices metoken.IndexPrices, asset sdk.Coin) (
	sdk.Coin,
	error,
) {
	i, assetSettings := index.AcceptedAsset(asset.Denom)
	if i < 0 {
		return sdk.Coin{}, sdkerrors.ErrNotFound.Wrapf("asset %s is not accepted in the index", asset.Denom)
	}

	// charge max fee if we don't want the token in the index.
	if assetSettings.TargetAllocation.IsZero() {
		return sdk.NewCoin(asset.Denom, index.Fee.MaxFee.MulInt(asset.Amount).TruncateInt()), nil
	}

	currentAllocation, err := k.currentAllocation(index, indexPrices, asset.Denom)
	if err != nil {
		return sdk.Coin{}, err
	}

	// when current_allocation is zero, we incentivize the swap by charging only min_fee
	if currentAllocation.IsZero() {
		return sdk.NewCoin(asset.Denom, index.Fee.MinFee.MulInt(asset.Amount).TruncateInt()), nil
	}

	allocationDeviation := currentAllocation.Sub(assetSettings.TargetAllocation).Quo(assetSettings.TargetAllocation)
	fee := index.Fee.CalculateFee(allocationDeviation)
	return sdk.NewCoin(asset.Denom, fee.MulInt(asset.Amount).TruncateInt()), nil
}

// redeemFee to be charged to the user, given a specific Index configuration and asset amount.
func (k Keeper) redeemFee(index metoken.Index, indexPrices metoken.IndexPrices, asset sdk.Coin) (
	sdk.Coin,
	error,
) {
	i, assetSettings := index.AcceptedAsset(asset.Denom)
	if i < 0 {
		return sdk.Coin{}, sdkerrors.ErrNotFound.Wrapf("asset %s is not accepted in the index", asset.Denom)
	}

	// charge min fee if we don't want the token in the index.
	if assetSettings.TargetAllocation.IsZero() {
		return sdk.NewCoin(asset.Denom, index.Fee.MinFee.MulInt(asset.Amount).TruncateInt()), nil
	}

	allocationDeviation, err := k.redeemAllocationDeviation(
		index,
		indexPrices,
		asset.Denom,
		assetSettings.TargetAllocation,
	)
	if err != nil {
		return sdk.Coin{}, err
	}

	fee := index.Fee.CalculateFee(allocationDeviation)
	return sdk.NewCoin(asset.Denom, fee.MulInt(asset.Amount).TruncateInt()), nil
}

// currentAllocation returns a factor of the assetDenom supply in the index based on the USD price value.
func (k Keeper) currentAllocation(
	index metoken.Index,
	indexPrices metoken.IndexPrices,
	assetDenom string,
) (sdk.Dec, error) {
	balances, err := k.IndexBalances(index.Denom)
	if err != nil {
		return sdk.Dec{}, err
	}

	i, balance := balances.AssetBalance(assetDenom)
	if i < 0 {
		return sdk.Dec{}, sdkerrors.ErrNotFound.Wrapf("balance for denom %s not found", assetDenom)
	}

	// if asset wasn't supplied to the index yet, the allocation is zero
	if balance.AvailableSupply().IsZero() {
		return sdk.ZeroDec(), nil
	}

	// if no meToken in balance, the allocation is zero
	if !balances.MetokenSupply.IsPositive() {
		return sdk.ZeroDec(), nil
	}

	assetPrice, err := indexPrices.Price(assetDenom)
	if err != nil {
		return sdk.Dec{}, err
	}
	assetUSD, err := valueInUSD(balance.AvailableSupply(), assetPrice.Price, assetPrice.Exponent)
	if err != nil {
		return sdk.Dec{}, err
	}

	meTokenPrice, err := indexPrices.Price(index.Denom)
	if err != nil {
		return sdk.Dec{}, err
	}
	meTokenUSD, err := valueInUSD(balances.MetokenSupply.Amount, meTokenPrice.Price, meTokenPrice.Exponent)
	if err != nil {
		return sdk.Dec{}, err
	}

	return assetUSD.Quo(meTokenUSD), nil
}

// redeemAllocationDeviation returns the delta between the target allocation and current allocation of an asset in an
// index for a redemption. It returns negative value when the asset is oversupplied.
func (k Keeper) redeemAllocationDeviation(
	index metoken.Index,
	indexPrices metoken.IndexPrices,
	assetDenom string,
	targetAllocation sdk.Dec,
) (sdk.Dec, error) {
	currentAllocation, err := k.currentAllocation(index, indexPrices, assetDenom)
	if err != nil {
		return sdk.Dec{}, err
	}

	if currentAllocation.IsZero() {
		return targetAllocation, nil
	}

	return targetAllocation.Sub(currentAllocation).Quo(targetAllocation), nil
}
