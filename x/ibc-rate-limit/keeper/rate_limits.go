package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v3/x/ibc-rate-limit/types"
)

// GetRateLimitsOfIBCDenoms returns rate limits of all registered ibc denoms.
func (k Keeper) GetRateLimitsOfIBCDenoms(ctx sdk.Context) ([]types.RateLimit, error) {
	var rateLimitsOfIBCDenoms []types.RateLimit

	prefix := types.KeyPrefixForIBCDenom
	store := ctx.KVStore(k.storeKey)

	iter := sdk.KVStorePrefixIterator(store, prefix)
	defer iter.Close()

	for ; iter.Valid(); iter.Next() {
		var rateLimitsOfIBCDenom types.RateLimit
		if err := rateLimitsOfIBCDenom.Unmarshal(iter.Value()); err != nil {
			return nil, err
		}

		rateLimitsOfIBCDenoms = append(rateLimitsOfIBCDenoms, rateLimitsOfIBCDenom)
	}

	return rateLimitsOfIBCDenoms, nil
}

// GetRateLimitsOfIBCDenom retunes the rate limits of ibc denom.
func (k Keeper) GetRateLimitsOfIBCDenom(ctx sdk.Context, ibcDenom string) (types.RateLimit, error) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(types.CreateKeyForRateLimitOfIBCDenom(ibcDenom))
	var rateLimitsOfIBCDenom types.RateLimit
	k.cdc.Unmarshal(bz, &rateLimitsOfIBCDenom)

	return rateLimitsOfIBCDenom, nil
}

// SetRateLimitsOfIBCDenoms save the rate limits of ibc denoms into store.
func (k Keeper) SetRateLimitsOfIBCDenoms(ctx sdk.Context, rateLimits []types.RateLimit) error {
	for _, rateLimitOfIBCDenom := range rateLimits {
		if err := k.SetRateLimitsOfIBCDenom(ctx, rateLimitOfIBCDenom); err != nil {
			return err
		}
	}

	return nil
}

// SetRateLimitsOfIBCDenom save the rate limits of ibc denom into store.
func (k Keeper) SetRateLimitsOfIBCDenom(ctx sdk.Context, rateLimitOfIBCDenom types.RateLimit) error {
	store := ctx.KVStore(k.storeKey)
	key := types.CreateKeyForRateLimitOfIBCDenom(rateLimitOfIBCDenom.IbcDenom)

	bz, err := k.cdc.Marshal(&rateLimitOfIBCDenom)
	if err != nil {
		return err
	}

	store.Set(key, bz)

	return nil
}
