package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	store "github.com/umee-network/umee/v3/util/store"
	"github.com/umee-network/umee/v3/x/incentive/types"
)

// setParams sets the x/incentive module's parameters.
func (k Keeper) setParams(ctx sdk.Context, params types.Params) error {
	kvs := k.kvStore(ctx)
	store.SetStoredUint32(kvs, types.KeyPrefixParamMaxUnbondings,
		params.MaxUnbondings, 1, "max unbondings")
	store.SetStoredUint64(kvs, types.KeyPrefixParamUnbondingDurationShort,
		params.UnbondingDurationShort, 1, "short unbonding duration")
	store.SetStoredUint64(kvs, types.KeyPrefixParamUnbondingDurationMiddle,
		params.UnbondingDurationMiddle, 1, "middle unbonding duration")
	store.SetStoredUint64(kvs, types.KeyPrefixParamUnbondingDurationLong,
		params.UnbondingDurationLong, 1, "long unbonding duration")
	store.SetStoredDec(kvs, types.KeyPrefixParamTierWeightShort,
		params.TierWeightShort, sdk.ZeroDec(), "short tier weight")
	store.SetStoredDec(kvs, types.KeyPrefixParamTierWeightMiddle,
		params.TierWeightMiddle, sdk.ZeroDec(), "middle tier weight")
	return nil
}

// GetParams gets all of the x/incentive module's parameters at once.
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	kvs := k.kvStore(ctx)
	return types.Params{
		MaxUnbondings: store.GetStoredUint32(kvs,
			types.KeyPrefixParamMaxUnbondings, 1, "max unbondings"),
		UnbondingDurationShort: store.GetStoredUint64(kvs,
			types.KeyPrefixParamUnbondingDurationShort, 1, "short unbonding duration"),
		UnbondingDurationMiddle: store.GetStoredUint64(kvs,
			types.KeyPrefixParamUnbondingDurationMiddle, 1, "middle unbonding duration"),
		UnbondingDurationLong: store.GetStoredUint64(kvs,
			types.KeyPrefixParamUnbondingDurationLong, 1, "long unbonding duration"),
		TierWeightShort: store.GetStoredDec(kvs,
			types.KeyPrefixParamTierWeightShort, sdk.ZeroDec(), "short tier weight"),
		TierWeightMiddle: store.GetStoredDec(kvs,
			types.KeyPrefixParamTierWeightMiddle, sdk.ZeroDec(), "middle tier weight"),
	}
}

// MaxUnbondings gets the x/incentive module's MaxUnbondings parameter.
func (k Keeper) MaxUnbondings(ctx sdk.Context) uint32 {
	return store.GetStoredUint32(k.kvStore(ctx),
		types.KeyPrefixParamMaxUnbondings, 1, "max unbondings")
}

// UnbondingDurationShort gets the x/incentive module's UnbondingDurationShort parameter.
func (k Keeper) UnbondingDurationShort(ctx sdk.Context) uint64 {
	return store.GetStoredUint64(k.kvStore(ctx),
		types.KeyPrefixParamUnbondingDurationShort, 1, "short unbonding duration")
}

// UnbondingDurationMiddle gets the x/incentive module's UnbondingDurationMiddle parameter.
func (k Keeper) UnbondingDurationMiddle(ctx sdk.Context) uint64 {
	return store.GetStoredUint64(k.kvStore(ctx),
		types.KeyPrefixParamUnbondingDurationMiddle, 1, "middle unbonding duration")
}

// UnbondingDurationLong gets the x/incentive module's UnbondingDurationLong parameter.
func (k Keeper) UnbondingDurationLong(ctx sdk.Context) uint64 {
	return store.GetStoredUint64(k.kvStore(ctx),
		types.KeyPrefixParamUnbondingDurationLong, 1, "long unbonding duration")
}

// TierWeightShort gets the x/incentive module's TierWeightShort parameter.
func (k Keeper) TierWeightShort(ctx sdk.Context) sdk.Dec {
	return store.GetStoredDec(k.kvStore(ctx),
		types.KeyPrefixParamTierWeightShort, sdk.ZeroDec(), "short tier weight")
}

// TierWeightMiddle gets the x/incentive module's TierWeightMiddle parameter.
func (k Keeper) TierWeightMiddle(ctx sdk.Context) sdk.Dec {
	return store.GetStoredDec(k.kvStore(ctx),
		types.KeyPrefixParamTierWeightMiddle, sdk.ZeroDec(), "middle tier weight")
}
