package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v4/util/store"
	"github.com/umee-network/umee/v4/x/incentive"
)

// GetMaxUnbondings gets the maximum number of unbondings an account is allowed to have at one time.
func (k Keeper) GetMaxUnbondings(ctx sdk.Context) uint32 {
	return store.GetUint32(k.KVStore(ctx),
		keyPrefixParamMaxUnbondings, 0, "max unbondings")
}

// GetUnbondingDurationLong gets the duration in seconds of the long bonding tier.
func (k Keeper) GetUnbondingDurationLong(ctx sdk.Context) uint64 {
	return store.GetUint64(k.KVStore(ctx),
		keyPrefixParamUnbondingDurationLong, 0, "long unbonding duration")
}

// GetUnbondingDurationMiddle gets the duration in seconds of the middle bonding tier.
func (k Keeper) GetUnbondingDurationMiddle(ctx sdk.Context) uint64 {
	return store.GetUint64(k.KVStore(ctx),
		keyPrefixParamUnbondingDurationMiddle, 0, "middle unbonding duration")
}

// GetUnbondingDurationShort gets the duration in seconds of the short bonding tier.
func (k Keeper) GetUnbondingDurationShort(ctx sdk.Context) uint64 {
	return store.GetUint64(k.KVStore(ctx),
		keyPrefixParamUnbondingDurationShort, 0, "short unbonding duration")
}

// GetTierWeightShort gets the ratio of rewards received by the short tier of bonded assets. Ranges 0 - 1.
func (k Keeper) GetTierWeightMiddle(ctx sdk.Context) sdk.Dec {
	return store.GetDec(k.KVStore(ctx),
		keyPrefixParamTierWeightMiddle, sdk.ZeroDec(), "middle tier weight")
}

// GetTierWeightShort gets the ratio of rewards received by the middle tier of bonded assets. Ranges 0 - 1.
func (k Keeper) GetTierWeightShort(ctx sdk.Context) sdk.Dec {
	return store.GetDec(k.KVStore(ctx),
		keyPrefixParamTierWeightShort, sdk.ZeroDec(), "short tier weight")
}

// GetCommunityFundAddress retrieves the community fund address parameter. It is guaranteed to be
// either valid (by sdk.ValidateAddressFormat) or empty.
func (k Keeper) GetCommunityFundAddress(ctx sdk.Context) sdk.AccAddress {
	return store.GetAddress(k.KVStore(ctx),
		keyPrefixParamCommunityFundAddress, "community fund address")
}

// setParams validates and sets the incentive module parameters
func (k Keeper) setParams(ctx sdk.Context, params incentive.Params) error {
	kvs := k.KVStore(ctx)
	if err := params.Validate(); err != nil {
		return err
	}
	err := store.SetUint32(kvs, keyPrefixParamMaxUnbondings,
		params.MaxUnbondings, 0, "max unbondings")
	if err != nil {
		return err
	}
	err = store.SetUint64(kvs, keyPrefixParamUnbondingDurationLong,
		params.UnbondingDurationLong, 0, "long unbonding duration")
	if err != nil {
		return err
	}
	err = store.SetUint64(kvs, keyPrefixParamUnbondingDurationMiddle,
		params.UnbondingDurationMiddle, 0, "middle unbonding duration")
	if err != nil {
		return err
	}
	err = store.SetUint64(kvs, keyPrefixParamUnbondingDurationShort,
		params.UnbondingDurationShort, 0, "short unbonding duration")
	if err != nil {
		return err
	}
	err = store.SetDec(kvs, keyPrefixParamTierWeightMiddle,
		params.TierWeightMiddle, sdk.ZeroDec(), "middle tier weight")
	if err != nil {
		return err
	}
	err = store.SetDec(kvs, keyPrefixParamTierWeightShort,
		params.TierWeightShort, sdk.ZeroDec(), "short tier weight")
	if err != nil {
		return err
	}
	err = store.SetAddress(kvs, keyPrefixParamCommunityFundAddress,
		sdk.MustAccAddressFromBech32(params.CommunityFundAddress), "community fund address")
	if err != nil {
		return err
	}
	return nil
}

// GetIncentiveProgram gets an incentive program by ID, regardless of the program's status.
func (k Keeper) GetIncentiveProgram(ctx sdk.Context, id uint32) (incentive.IncentiveProgram, error) {
	keys := [][]byte{
		keyUpcomingIncentiveProgram(id),
		keyOngoingIncentiveProgram(id),
		keyCompletedIncentiveProgram(id),
	}

	store := ctx.KVStore(k.storeKey)

	// Looks for an incentive program with the specified ID in upcoming, ongoing, then completed program lists.
	for _, key := range keys {
		program := incentive.IncentiveProgram{}
		bz := store.Get(key)
		if len(bz) != 0 {
			err := k.cdc.Unmarshal(bz, &program)
			return program, err
		}
	}

	return incentive.IncentiveProgram{}, incentive.ErrNoProgramWithID
}

// GetNextProgramID gets the ID that will be assigned to the next incentive program passed by governance.
func (k Keeper) GetNextProgramID(ctx sdk.Context) uint32 {
	return store.GetUint32(k.KVStore(ctx), keyPrefixNextProgramID, 0, "next program ID")
}

// setNextProgramID sets the ID that will be assigned to the next incentive program passed by governance.
func (k Keeper) setNextProgramID(ctx sdk.Context, id uint32) error {
	prev := k.GetNextProgramID(ctx)
	if id < prev {
		return incentive.ErrDecreaseNextProgramID.Wrapf("%s to %s", id, prev)
	}
	return store.SetUint32(k.KVStore(ctx), keyPrefixNextProgramID, id, 0, "next program ID")
}

// GetLastRewardsTime gets the last unix time incentive rewards were computed globally by EndBlocker.
func (k Keeper) GetLastRewardsTime(ctx sdk.Context) uint64 {
	return store.GetUint64(k.KVStore(ctx), keyPrefixLastRewardsTime, 0, "last reward time")
}

// setLastRewardsTime sets the last unix time incentive rewards were computed globally by EndBlocker.
func (k Keeper) setLastRewardsTime(ctx sdk.Context, time uint64) error {
	prev := k.GetLastRewardsTime(ctx)
	if time < prev {
		return incentive.ErrDecreaseLastRewardTime.Wrapf("%s to %s", time, prev)
	}
	return store.SetUint64(k.KVStore(ctx), keyPrefixLastRewardsTime, time, 0, "last reward time")
}

// GetTotalBonded retrieves the total amount of uTokens of a given denom which are bonded to the incentive module
func (k Keeper) GetTotalBonded(ctx sdk.Context, denom string) sdk.Coin {
	key := keyTotalBonded(denom)
	amount := store.GetInt(k.KVStore(ctx), key, sdk.ZeroInt(), "total bonded "+denom)
	return sdk.NewCoin(denom, amount)
}

// setTotalBonded records the total amount of uTokens of a given denom which are bonded to the incentive module
func (k Keeper) setTotalBonded(ctx sdk.Context, uTokens sdk.Coin) error {
	key := keyTotalBonded(uTokens.Denom)
	return store.SetInt(k.KVStore(ctx), key, uTokens.Amount, sdk.ZeroInt(), "total bonded "+uTokens.Denom)
}

// GetBonded retrieves the amount of uTokens of a given denom which are bonded by a single account
func (k Keeper) GetBonded(ctx sdk.Context, denom string) sdk.Coin {
	key := keyTotalBonded(denom)
	amount := store.GetInt(k.KVStore(ctx), key, sdk.ZeroInt(), "bonded amount")
	return sdk.NewCoin(denom, amount)
}
