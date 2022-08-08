package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// filterCoins returns the subset of an sdk.Coins that meet a given condition
func (k Keeper) filterCoins(coins sdk.Coins, accept func(sdk.Coin) bool) sdk.Coins {
	filtered := sdk.Coins{}
	for _, c := range coins {
		if accept(c) {
			filtered = append(filtered, c)
		}
	}
	return filtered
}

// filterAcceptedCoins returns the subset of an sdk.Coins that are accepted, non-blacklisted tokens
func (k Keeper) filterAcceptedCoins(ctx sdk.Context, coins sdk.Coins) sdk.Coins {
	return k.filterCoins(
		coins,
		func(c sdk.Coin) bool {
			return k.validateAcceptedAsset(ctx, c) == nil
		},
	)
}
