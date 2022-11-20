package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	store "github.com/umee-network/umee/v3/util/store"
	"github.com/umee-network/umee/v3/x/incentive"
)

// setParams sets the x/incentive module's parameters.
func (k Keeper) setParams(ctx sdk.Context, params incentive.Params) error {
	kvs := k.kvStore(ctx)
	if err := store.SetStoredUint32(kvs, incentive.KeyPrefixParamMaxUnbondings,
		params.MaxUnbondings, 1, "max unbondings",
	); err != nil {
		return err
	}
	if err := store.SetStoredUint64(kvs, incentive.KeyPrefixParamUnbondingDurationShort,
		params.UnbondingDurationShort, 1, "short unbonding duration",
	); err != nil {
		return err
	}
	if err := store.SetStoredUint64(kvs, incentive.KeyPrefixParamUnbondingDurationMiddle,
		params.UnbondingDurationMiddle, 1, "middle unbonding duration",
	); err != nil {
		return err
	}
	if err := store.SetStoredUint64(kvs, incentive.KeyPrefixParamUnbondingDurationLong,
		params.UnbondingDurationLong, 1, "long unbonding duration",
	); err != nil {
		return err
	}
	if err := store.SetStoredDec(kvs, incentive.KeyPrefixParamTierWeightShort,
		params.TierWeightShort, sdk.ZeroDec(), "short tier weight",
	); err != nil {
		return err
	}
	return store.SetStoredDec(kvs, incentive.KeyPrefixParamTierWeightMiddle,
		params.TierWeightMiddle, sdk.ZeroDec(), "middle tier weight")
}

// GetParams gets all of the x/incentive module's parameters at once.
func (k Keeper) GetParams(ctx sdk.Context) incentive.Params {
	kvs := k.kvStore(ctx)
	return incentive.Params{
		MaxUnbondings: store.GetStoredUint32(kvs,
			incentive.KeyPrefixParamMaxUnbondings, 1, "max unbondings"),
		UnbondingDurationShort: store.GetStoredUint64(kvs,
			incentive.KeyPrefixParamUnbondingDurationShort, 1, "short unbonding duration"),
		UnbondingDurationMiddle: store.GetStoredUint64(kvs,
			incentive.KeyPrefixParamUnbondingDurationMiddle, 1, "middle unbonding duration"),
		UnbondingDurationLong: store.GetStoredUint64(kvs,
			incentive.KeyPrefixParamUnbondingDurationLong, 1, "long unbonding duration"),
		TierWeightShort: store.GetStoredDec(kvs,
			incentive.KeyPrefixParamTierWeightShort, sdk.ZeroDec(), "short tier weight"),
		TierWeightMiddle: store.GetStoredDec(kvs,
			incentive.KeyPrefixParamTierWeightMiddle, sdk.ZeroDec(), "middle tier weight"),
	}
}

// MaxUnbondings gets the x/incentive module's MaxUnbondings parameter.
func (k Keeper) MaxUnbondings(ctx sdk.Context) uint32 {
	return store.GetStoredUint32(k.kvStore(ctx),
		incentive.KeyPrefixParamMaxUnbondings, 1, "max unbondings")
}

// UnbondingDurationShort gets the x/incentive module's UnbondingDurationShort parameter.
func (k Keeper) UnbondingDurationShort(ctx sdk.Context) uint64 {
	return store.GetStoredUint64(k.kvStore(ctx),
		incentive.KeyPrefixParamUnbondingDurationShort, 1, "short unbonding duration")
}

// UnbondingDurationMiddle gets the x/incentive module's UnbondingDurationMiddle parameter.
func (k Keeper) UnbondingDurationMiddle(ctx sdk.Context) uint64 {
	return store.GetStoredUint64(k.kvStore(ctx),
		incentive.KeyPrefixParamUnbondingDurationMiddle, 1, "middle unbonding duration")
}

// UnbondingDurationLong gets the x/incentive module's UnbondingDurationLong parameter.
func (k Keeper) UnbondingDurationLong(ctx sdk.Context) uint64 {
	return store.GetStoredUint64(k.kvStore(ctx),
		incentive.KeyPrefixParamUnbondingDurationLong, 1, "long unbonding duration")
}

// TierWeightShort gets the x/incentive module's TierWeightShort parameter.
func (k Keeper) TierWeightShort(ctx sdk.Context) sdk.Dec {
	return store.GetStoredDec(k.kvStore(ctx),
		incentive.KeyPrefixParamTierWeightShort, sdk.ZeroDec(), "short tier weight")
}

// TierWeightMiddle gets the x/incentive module's TierWeightMiddle parameter.
func (k Keeper) TierWeightMiddle(ctx sdk.Context) sdk.Dec {
	return store.GetStoredDec(k.kvStore(ctx),
		incentive.KeyPrefixParamTierWeightMiddle, sdk.ZeroDec(), "middle tier weight")
}
