package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/umee-network/umee/v4/util/store"
	"github.com/umee-network/umee/v4/x/incentive"
)

// GetMaxUnbondings gets the maximum number of unbondings an account is allowed to have at one time.
func (k Keeper) GetMaxUnbondings(ctx sdk.Context) uint32 {
	return store.GetUint32(k.KVStore(ctx),
		keyPrefixParamMaxUnbondings, "max unbondings")
}

// GetUnbondingDurationLong gets the duration in seconds of the long bonding tier.
func (k Keeper) GetUnbondingDurationLong(ctx sdk.Context) uint64 {
	return store.GetUint64(k.KVStore(ctx),
		keyPrefixParamUnbondingDurationLong, "long unbonding duration")
}

// GetUnbondingDurationMiddle gets the duration in seconds of the middle bonding tier.
func (k Keeper) GetUnbondingDurationMiddle(ctx sdk.Context) uint64 {
	return store.GetUint64(k.KVStore(ctx),
		keyPrefixParamUnbondingDurationMiddle, "middle unbonding duration")
}

// GetUnbondingDurationShort gets the duration in seconds of the short bonding tier.
func (k Keeper) GetUnbondingDurationShort(ctx sdk.Context) uint64 {
	return store.GetUint64(k.KVStore(ctx),
		keyPrefixParamUnbondingDurationShort, "short unbonding duration")
}

// GetTierWeightShort gets the ratio of rewards received by the short tier of bonded assets. Ranges 0 - 1.
func (k Keeper) GetTierWeightMiddle(ctx sdk.Context) sdk.Dec {
	return store.GetDec(k.KVStore(ctx),
		keyPrefixParamTierWeightMiddle, "middle tier weight")
}

// GetTierWeightShort gets the ratio of rewards received by the middle tier of bonded assets. Ranges 0 - 1.
func (k Keeper) GetTierWeightShort(ctx sdk.Context) sdk.Dec {
	return store.GetDec(k.KVStore(ctx),
		keyPrefixParamTierWeightShort, "short tier weight")
}

// GetCommunityFundAddress retrieves the community fund address parameter. It is guaranteed to be
// either valid (by sdk.ValidateAddressFormat) or empty.
func (k Keeper) GetCommunityFundAddress(ctx sdk.Context) sdk.AccAddress {
	return store.GetAddress(k.KVStore(ctx),
		keyPrefixParamCommunityFundAddress, "community fund address")
}

// SetParams validates and sets the incentive module parameters
func (k Keeper) SetParams(ctx sdk.Context, params incentive.Params) error {
	kvs := k.KVStore(ctx)
	if err := params.Validate(); err != nil {
		return err
	}
	err := store.SetUint32(kvs, keyPrefixParamMaxUnbondings,
		params.MaxUnbondings, "max unbondings")
	if err != nil {
		return err
	}
	err = store.SetUint64(kvs, keyPrefixParamUnbondingDurationLong,
		params.UnbondingDurationLong, "long unbonding duration")
	if err != nil {
		return err
	}
	err = store.SetUint64(kvs, keyPrefixParamUnbondingDurationMiddle,
		params.UnbondingDurationMiddle, "middle unbonding duration")
	if err != nil {
		return err
	}
	err = store.SetUint64(kvs, keyPrefixParamUnbondingDurationShort,
		params.UnbondingDurationShort, "short unbonding duration")
	if err != nil {
		return err
	}
	err = store.SetDec(kvs, keyPrefixParamTierWeightMiddle,
		params.TierWeightMiddle, "middle tier weight")
	if err != nil {
		return err
	}
	err = store.SetDec(kvs, keyPrefixParamTierWeightShort,
		params.TierWeightShort, "short tier weight")
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
// Returns the program's status if found, or an error if it does not exist.
func (k Keeper) GetIncentiveProgram(ctx sdk.Context, id uint32) (
	incentive.IncentiveProgram, incentive.ProgramStatus, error,
) {
	statuses := []incentive.ProgramStatus{
		incentive.ProgramStatusUpcoming,
		incentive.ProgramStatusOngoing,
		incentive.ProgramStatusCompleted,
	}

	kvStore := k.KVStore(ctx)

	// Looks for an incentive program with the specified ID in upcoming, ongoing, then completed program lists.
	for _, status := range statuses {
		program := incentive.IncentiveProgram{}
		bz := kvStore.Get(keyIncentiveProgram(id, status))
		if len(bz) != 0 {
			err := k.cdc.Unmarshal(bz, &program)
			return program, status, err
		}
	}

	return incentive.IncentiveProgram{}, 0, sdkerrors.ErrNotFound
}

// SetIncentiveProgram stores an incentive program in either the upcoming, ongoing, or completed program lists.
// does not validate the incentive program struct or the validity of its status change (e.g. upcoming -> complete)
func (k Keeper) SetIncentiveProgram(ctx sdk.Context,
	program incentive.IncentiveProgram, status incentive.ProgramStatus,
) error {
	keys := [][]byte{
		keyIncentiveProgram(program.Id, incentive.ProgramStatusUpcoming),
		keyIncentiveProgram(program.Id, incentive.ProgramStatusOngoing),
		keyIncentiveProgram(program.Id, incentive.ProgramStatusCompleted),
	}

	kvStore := k.KVStore(ctx)
	for _, key := range keys {
		// always clear the program from the status it was prevously stored under
		kvStore.Delete(key)
	}

	key := keyIncentiveProgram(program.Id, status)
	bz, err := k.cdc.Marshal(&program)
	if err != nil {
		return err
	}
	kvStore.Set(key, bz)
	return nil
}

// GetNextProgramID gets the ID that will be assigned to the next incentive program passed by governance.
func (k Keeper) GetNextProgramID(ctx sdk.Context) uint32 {
	return store.GetUint32(k.KVStore(ctx), keyPrefixNextProgramID, "next program ID")
}

// SetNextProgramID sets the ID that will be assigned to the next incentive program passed by governance.
func (k Keeper) SetNextProgramID(ctx sdk.Context, id uint32) error {
	prev := k.GetNextProgramID(ctx)
	if id < prev {
		return incentive.ErrDecreaseNextProgramID.Wrapf("%d to %d", id, prev)
	}
	return store.SetUint32(k.KVStore(ctx), keyPrefixNextProgramID, id, "next program ID")
}

// GetLastRewardsTime gets the last unix time incentive rewards were computed globally by EndBlocker.
func (k Keeper) GetLastRewardsTime(ctx sdk.Context) uint64 {
	return store.GetUint64(k.KVStore(ctx), keyPrefixLastRewardsTime, "last reward time")
}

// SetLastRewardsTime sets the last unix time incentive rewards were computed globally by EndBlocker.
func (k Keeper) SetLastRewardsTime(ctx sdk.Context, time uint64) error {
	prev := k.GetLastRewardsTime(ctx)
	if time < prev {
		return incentive.ErrDecreaseLastRewardTime.Wrapf("%d to %d", time, prev)
	}
	return store.SetUint64(k.KVStore(ctx), keyPrefixLastRewardsTime, time, "last reward time")
}

// GetTotalBonded retrieves the total amount of uTokens of a given denom which are bonded to the incentive module
func (k Keeper) GetTotalBonded(ctx sdk.Context, denom string, tier incentive.BondTier) sdk.Coin {
	key := keyTotalBonded(denom, tier)
	amount := store.GetInt(k.KVStore(ctx), key, "total bonded")
	return sdk.NewCoin(denom, amount)
}

// SetTotalBonded records the total amount of uTokens of a given denom which are bonded to the incentive module
func (k Keeper) SetTotalBonded(ctx sdk.Context, uTokens sdk.Coin, tier incentive.BondTier) error {
	key := keyTotalBonded(uTokens.Denom, tier)
	return store.SetInt(k.KVStore(ctx), key, uTokens.Amount, "total bonded")
}

// GetBonded retrieves the amount of uTokens of a given denom which are bonded to a single tier by an account
func (k Keeper) GetBonded(ctx sdk.Context, addr sdk.AccAddress, denom string, tier incentive.BondTier) sdk.Coin {
	key := keyBondAmount(addr, denom, tier)
	amount := store.GetInt(k.KVStore(ctx), key, "bonded amount")
	return sdk.NewCoin(denom, amount)
}

// SetBonded sets the amount of uTokens of a given denom which are bonded to a single tier by an account
func (k Keeper) SetBonded(ctx sdk.Context,
	addr sdk.AccAddress, uToken sdk.Coin, tier incentive.BondTier,
) error {
	key := keyBondAmount(addr, uToken.Denom, tier)
	return store.SetInt(k.KVStore(ctx), key, uToken.Amount, "bonded amount")
}

// GetRewardAccumulator retrieves the reward accumulator of a reward token for a single bonded uToken and tier -
// for example, how much UMEE (reward) would have been earned by 1 ATOM bonded to the middle tier since genesis.
func (k Keeper) GetRewardAccumulator(ctx sdk.Context, bondDenom, rewardDenom string, tier incentive.BondTier,
) sdk.DecCoin {
	key := keyRewardAccumulator(bondDenom, rewardDenom, tier)
	amount := store.GetDec(k.KVStore(ctx), key, "reward accumulator")
	return sdk.NewDecCoinFromDec(rewardDenom, amount)
}

// SetRewardAccumulator sets the reward accumulator of a reward token for a single bonded uToken and tier.
func (k Keeper) SetRewardAccumulator(ctx sdk.Context,
	bondDenom string, reward sdk.DecCoin, tier incentive.BondTier,
) error {
	key := keyRewardAccumulator(bondDenom, reward.Denom, tier)
	return store.SetDec(k.KVStore(ctx), key, reward.Amount, "reward accumulator")
}

// GetRewardTracker retrieves the reward tracker of a reward token for a single bonded uToken and tier on one account -
// this is the value of the reward accumulator for those specific denoms and tier the last time this account performed
// and action that requires a reward tracker update (i.e. Bond, Claim, BeginUnbonding, or being Liquidated).
func (k Keeper) GetRewardTracker(ctx sdk.Context,
	addr sdk.AccAddress, bondDenom, rewardDenom string, tier incentive.BondTier,
) sdk.DecCoin {
	key := keyRewardTracker(addr, bondDenom, rewardDenom, tier)
	amount := store.GetDec(k.KVStore(ctx), key, "reward tracker")
	return sdk.NewDecCoinFromDec(rewardDenom, amount)
}

// SetRewardTracker sets the reward tracker of a reward token for a single bonded uToken and tier for an address.
func (k Keeper) SetRewardTracker(ctx sdk.Context,
	addr sdk.AccAddress, bondDenom string, reward sdk.DecCoin, tier incentive.BondTier,
) error {
	key := keyRewardTracker(addr, bondDenom, reward.Denom, tier)
	return store.SetDec(k.KVStore(ctx), key, reward.Amount, "reward tracker")
}

// GetUnbondings gets all unbondings currently associated with an account.
func (k Keeper) GetUnbondings(ctx sdk.Context, addr sdk.AccAddress) []incentive.Unbonding {
	key := keyUnbondings(addr)
	kvStore := k.KVStore(ctx)

	accUnbondings := incentive.AccountUnbondings{}
	bz := kvStore.Get(key)
	if len(bz) == 0 {
		return []incentive.Unbonding{}
	}

	k.cdc.MustUnmarshal(bz, &accUnbondings)
	return accUnbondings.Unbondings
}

// SetUnbondings stores the full list of unbondings currently associated with an account.
func (k Keeper) SetUnbondings(ctx sdk.Context, unbondings incentive.AccountUnbondings) error {
	kvStore := k.KVStore(ctx)
	addr, err := sdk.AccAddressFromBech32(unbondings.Account)
	if err != nil {
		// catches invalid and empty addresses
		return err
	}
	key := keyUnbondings(addr)
	if len(unbondings.Unbondings) == 0 {
		// clear store on no unbondings remaining
		kvStore.Delete(key)
	}
	bz, err := k.cdc.Marshal(&unbondings)
	if err != nil {
		return err
	}
	kvStore.Set(key, bz)
	return nil
}
