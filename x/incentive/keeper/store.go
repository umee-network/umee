package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/umee-network/umee/v4/util/store"
	"github.com/umee-network/umee/v4/x/incentive"
)

// getMaxUnbondings gets the maximum number of unbondings an account is allowed to have at one time.
func (k Keeper) getMaxUnbondings(ctx sdk.Context) uint32 {
	return store.GetUint32(k.KVStore(ctx),
		keyPrefixParamMaxUnbondings, "max unbondings")
}

// getUnbondingDurationLong gets the duration in seconds of the long bonding tier.
func (k Keeper) getUnbondingDurationLong(ctx sdk.Context) uint64 {
	return store.GetUint64(k.KVStore(ctx),
		keyPrefixParamUnbondingDurationLong, "long unbonding duration")
}

// getUnbondingDurationMiddle gets the duration in seconds of the middle bonding tier.
func (k Keeper) getUnbondingDurationMiddle(ctx sdk.Context) uint64 {
	return store.GetUint64(k.KVStore(ctx),
		keyPrefixParamUnbondingDurationMiddle, "middle unbonding duration")
}

// getUnbondingDurationShort gets the duration in seconds of the short bonding tier.
func (k Keeper) getUnbondingDurationShort(ctx sdk.Context) uint64 {
	return store.GetUint64(k.KVStore(ctx),
		keyPrefixParamUnbondingDurationShort, "short unbonding duration")
}

// GetTierWeightShort gets the ratio of rewards received by the short tier of bonded assets. Ranges 0 - 1.
func (k Keeper) getTierWeightMiddle(ctx sdk.Context) sdk.Dec {
	return store.GetDec(k.KVStore(ctx),
		keyPrefixParamTierWeightMiddle, "middle tier weight")
}

// getTierWeightShort gets the ratio of rewards received by the middle tier of bonded assets. Ranges 0 - 1.
func (k Keeper) getTierWeightShort(ctx sdk.Context) sdk.Dec {
	return store.GetDec(k.KVStore(ctx),
		keyPrefixParamTierWeightShort, "short tier weight")
}

// getCommunityFundAddress retrieves the community fund address parameter. It is guaranteed to be
// either valid (by sdk.ValidateAddressFormat) or empty.
func (k Keeper) getCommunityFundAddress(ctx sdk.Context) sdk.AccAddress {
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

// getIncentiveProgram gets an incentive program by ID, regardless of the program's status.
// Returns the program's status if found, or an error if it does not exist.
func (k Keeper) getIncentiveProgram(ctx sdk.Context, id uint32) (
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

// setIncentiveProgram stores an incentive program in either the upcoming, ongoing, or completed program lists.
// does not validate the incentive program struct or the validity of its status change (e.g. upcoming -> complete)
func (k Keeper) setIncentiveProgram(ctx sdk.Context,
	program incentive.IncentiveProgram, status incentive.ProgramStatus,
) error {
	keys := [][]byte{
		keyIncentiveProgram(program.ID, incentive.ProgramStatusUpcoming),
		keyIncentiveProgram(program.ID, incentive.ProgramStatusOngoing),
		keyIncentiveProgram(program.ID, incentive.ProgramStatusCompleted),
	}

	kvStore := k.KVStore(ctx)
	for _, key := range keys {
		// always clear the program from the status it was prevously stored under
		kvStore.Delete(key)
	}

	key := keyIncentiveProgram(program.ID, status)
	bz, err := k.cdc.Marshal(&program)
	if err != nil {
		return err
	}
	kvStore.Set(key, bz)
	return nil
}

// getNextProgramID gets the ID that will be assigned to the next incentive program passed by governance.
func (k Keeper) getNextProgramID(ctx sdk.Context) uint32 {
	return store.GetUint32(k.KVStore(ctx), keyPrefixNextProgramID, "next program ID")
}

// setNextProgramID sets the ID that will be assigned to the next incentive program passed by governance.
func (k Keeper) setNextProgramID(ctx sdk.Context, id uint32) error {
	prev := k.getNextProgramID(ctx)
	if id < prev {
		return incentive.ErrDecreaseNextProgramID.Wrapf("%d to %d", id, prev)
	}
	return store.SetUint32(k.KVStore(ctx), keyPrefixNextProgramID, id, "next program ID")
}

// getLastRewardsTime gets the last unix time incentive rewards were computed globally by EndBlocker.
func (k Keeper) getLastRewardsTime(ctx sdk.Context) uint64 {
	return store.GetUint64(k.KVStore(ctx), keyPrefixLastRewardsTime, "last reward time")
}

// setLastRewardsTime sets the last unix time incentive rewards were computed globally by EndBlocker.
func (k Keeper) setLastRewardsTime(ctx sdk.Context, time uint64) error {
	prev := k.getLastRewardsTime(ctx)
	if time < prev {
		return incentive.ErrDecreaseLastRewardTime.Wrapf("%d to %d", time, prev)
	}
	return store.SetUint64(k.KVStore(ctx), keyPrefixLastRewardsTime, time, "last reward time")
}

// getTotalBonded retrieves the total amount of uTokens of a given denom which are bonded to the incentive module
func (k Keeper) getTotalBonded(ctx sdk.Context, denom string, tier incentive.BondTier) sdk.Coin {
	key := keyTotalBonded(denom, tier)
	amount := store.GetInt(k.KVStore(ctx), key, "total bonded")
	return sdk.NewCoin(denom, amount)
}

// getTotalUnbonding retrieves the total amount of uTokens of a given denom which are unbonding from
// the incentive module
func (k Keeper) getTotalUnbonding(ctx sdk.Context, denom string, tier incentive.BondTier) sdk.Coin {
	key := keyTotalUnbonding(denom, tier)
	amount := store.GetInt(k.KVStore(ctx), key, "total unbonding")
	return sdk.NewCoin(denom, amount)
}

// getBonded retrieves the amount of uTokens of a given denom which are bonded to a single tier by an account
func (k Keeper) getBonded(ctx sdk.Context, addr sdk.AccAddress, denom string, tier incentive.BondTier) sdk.Coin {
	key := keyBondAmount(addr, denom, tier)
	amount := store.GetInt(k.KVStore(ctx), key, "bonded amount")
	return sdk.NewCoin(denom, amount)
}

// setBonded sets the amount of uTokens of a given denom which are bonded to a single tier by an account.
// Automatically updates TotalBonded as well.
//
// REQUIREMENT: This is the only function which is allowed to set TotalBonded.
func (k Keeper) setBonded(ctx sdk.Context,
	addr sdk.AccAddress, uToken sdk.Coin, tier incentive.BondTier,
) error {
	// compute the change in bonded amount (can be negative when bond decreases)
	delta := uToken.Amount.Sub(k.getBonded(ctx, addr, uToken.Denom, tier).Amount)

	// Set bond amount
	key := keyBondAmount(addr, uToken.Denom, tier)
	if err := store.SetInt(k.KVStore(ctx), key, uToken.Amount, "bonded amount"); err != nil {
		return err
	}

	// Update total bonded for this utoken denom using the computed change
	total := k.getTotalBonded(ctx, uToken.Denom, tier)
	totalkey := keyTotalBonded(uToken.Denom, tier)
	return store.SetInt(k.KVStore(ctx), totalkey, total.Amount.Add(delta), "total bonded")
}

// getUnbonding retrieves the amount of uTokens of a given denom which are unbonding from a single tier by an account
func (k Keeper) getUnbondingAmount(ctx sdk.Context, addr sdk.AccAddress, denom string, tier incentive.BondTier,
) sdk.Coin {
	key := keyUnbondAmount(addr, denom, tier)
	amount := store.GetInt(k.KVStore(ctx), key, "unbonding amount")
	return sdk.NewCoin(denom, amount)
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

// setRewardTracker sets the reward tracker of a reward token for a single bonded uToken and tier for an address.
func (k Keeper) setRewardTracker(ctx sdk.Context,
	addr sdk.AccAddress, bondDenom string, reward sdk.DecCoin, tier incentive.BondTier,
) error {
	key := keyRewardTracker(addr, bondDenom, reward.Denom, tier)
	return store.SetDec(k.KVStore(ctx), key, reward.Amount, "reward tracker")
}

// getUnbondings gets all unbondings currently associated with an account, bonded denom, and tier.
func (k Keeper) getUnbondings(ctx sdk.Context, addr sdk.AccAddress, denom string, tier incentive.BondTier,
) []incentive.Unbonding {
	key := keyUnbondings(addr, denom, tier)
	kvStore := k.KVStore(ctx)

	accUnbondings := incentive.AccountUnbondings{}
	bz := kvStore.Get(key)
	if len(bz) == 0 {
		return []incentive.Unbonding{}
	}

	k.cdc.MustUnmarshal(bz, &accUnbondings)
	return accUnbondings.Unbondings
}

// setUnbondings stores the list of unbondings currently associated with an account, denom, and tier.
// It also updates the account's stored unbonding amounts, and thus the module's total unbonding as well.
//
// REQUIREMENT: This is the only function which is allowed to set unbonding amounts and total unbonding.
func (k Keeper) setUnbondings(ctx sdk.Context, unbondings incentive.AccountUnbondings) error {
	kvStore := k.KVStore(ctx)
	addr, err := sdk.AccAddressFromBech32(unbondings.Account)
	if err != nil {
		// catches invalid and empty addresses
		return err
	}
	tier, err := bondTier(unbondings.Tier)
	if err != nil {
		// catches invalid or unspecified tier
		return err
	}
	denom := unbondings.Denom

	// compute the new total unbonding specific to this account, denom, and tier
	newUnbonding := sdk.ZeroInt()
	for _, u := range unbondings.Unbondings {
		newUnbonding = newUnbonding.Add(u.Amount.Amount)
	}
	// compute the change in unbonding amount (can be negative when unbonding decreases)
	delta := newUnbonding.Sub(k.getUnbondingAmount(ctx, addr, denom, tier).Amount)

	// Update unbonding amount
	amountKey := keyUnbondAmount(addr, denom, tier)
	if err := store.SetInt(k.KVStore(ctx), amountKey, newUnbonding, "unbonding amount"); err != nil {
		return err
	}

	// Update total unbonding for this utoken denom using the computed change
	total := k.getTotalUnbonding(ctx, denom, tier)
	totalkey := keyTotalUnbonding(denom, tier)
	if err := store.SetInt(k.KVStore(ctx), totalkey, total.Amount.Add(delta), "total unbonding"); err != nil {
		return err
	}

	// set list of unbondings
	key := keyUnbondings(addr, unbondings.Denom, tier)
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
