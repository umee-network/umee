package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v4/util/store"
	"github.com/umee-network/umee/v4/x/incentive"
)

// GetMaxUnbondings gets the maximum number of unbondings an account is allowed to have at one time.
func (k Keeper) GetMaxUnbondings(ctx sdk.Context) uint32 {
	return store.GetUint32(k.KVStore(ctx), incentive.KeyPrefixParamMaxUnbondings, 0)
}

// GetUnbondingDurationLong gets the duration in seconds of the long bonding tier.
func (k Keeper) GetUnbondingDurationLong(ctx sdk.Context) uint64 {
	return store.GetUint64(k.KVStore(ctx), incentive.KeyPrefixParamUnbondingDurationLong, 0)
}

// GetUnbondingDurationMiddle gets the duration in seconds of the middle bonding tier.
func (k Keeper) GetUnbondingDurationMiddle(ctx sdk.Context) uint64 {
	return store.GetUint64(k.KVStore(ctx), incentive.KeyPrefixParamUnbondingDurationMiddle, 0)
}

// GetUnbondingDurationShort gets the duration in seconds of the short bonding tier.
func (k Keeper) GetUnbondingDurationShort(ctx sdk.Context) uint64 {
	return store.GetUint64(k.KVStore(ctx), incentive.KeyPrefixParamUnbondingDurationShort, 0)
}

// GetTierWeightShort gets the ratio of rewards received by the short tier of bonded assets. Ranges 0 - 1.
func (k Keeper) GetTierWeightMiddle(ctx sdk.Context) sdk.Dec {
	return store.GetDec(k.KVStore(ctx), incentive.KeyPrefixParamTierWeightMiddle, sdk.ZeroDec())
}

// GetTierWeightShort gets the ratio of rewards received by the middle tier of bonded assets. Ranges 0 - 1.
func (k Keeper) GetTierWeightShort(ctx sdk.Context) sdk.Dec {
	return store.GetDec(k.KVStore(ctx), incentive.KeyPrefixParamTierWeightShort, sdk.ZeroDec())
}

// GetCommunityFundAddress retrieves the community fund address parameter. It is guaranteed to be
// either valid (by sdk.ValidateAddressFormat) or empty.
func (k Keeper) GetCommunityFundAddress(ctx sdk.Context) sdk.AccAddress {
	return store.GetAddress(k.KVStore(ctx), incentive.KeyPrefixParamCommunityFundAddress)
}

// SetParams validates and sets the incentive module parameters
func (k Keeper) SetParams(ctx sdk.Context, params incentive.Params) error {
	kvs := k.KVStore(ctx)
	if err := params.Validate(); err != nil {
		return err
	}
	err := store.SetUint32(kvs, incentive.KeyPrefixParamMaxUnbondings, params.MaxUnbondings, 0)
	if err != nil {
		return err
	}
	err = store.SetUint64(kvs, incentive.KeyPrefixParamUnbondingDurationLong, params.UnbondingDurationLong, 0)
	if err != nil {
		return err
	}
	err = store.SetUint64(kvs, incentive.KeyPrefixParamUnbondingDurationMiddle, params.UnbondingDurationMiddle, 0)
	if err != nil {
		return err
	}
	err = store.SetUint64(kvs, incentive.KeyPrefixParamUnbondingDurationShort, params.UnbondingDurationShort, 0)
	if err != nil {
		return err
	}
	err = store.SetDec(kvs, incentive.KeyPrefixParamTierWeightMiddle, params.TierWeightMiddle, sdk.ZeroDec())
	if err != nil {
		return err
	}
	err = store.SetDec(kvs, incentive.KeyPrefixParamTierWeightShort, params.TierWeightShort, sdk.ZeroDec())
	if err != nil {
		return err
	}
	err = store.SetAddress(kvs, incentive.KeyPrefixParamCommunityFundAddress,
		sdk.MustAccAddressFromBech32(params.CommunityFundAddress))
	if err != nil {
		return err
	}
	return nil
}
