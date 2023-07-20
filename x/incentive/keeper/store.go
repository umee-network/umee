package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/umee-network/umee/v5/util/store"
	"github.com/umee-network/umee/v5/x/incentive"
)

func (k Keeper) GetParams(ctx sdk.Context) incentive.Params {
	params := store.GetValue[*incentive.Params](k.KVStore(ctx), keyPrefixParams, "params")
	if params == nil {
		// on missing module parameters, return defaults rather than panicking or returning zero values
		return incentive.DefaultParams()
	}
	return *params
}

// setParams validates and sets the incentive module parameters
func (k Keeper) setParams(ctx sdk.Context, params incentive.Params) error {
	if err := params.Validate(); err != nil {
		return err
	}
	return store.SetValue(k.KVStore(ctx), keyPrefixParams, &params, "params")
}

// incentiveProgramStatus gets an incentive program status by ID. If the program is not found,
// the status is NotExist. Error only if the program is found under multiple statuses.
func (k Keeper) incentiveProgramStatus(ctx sdk.Context, id uint32) (incentive.ProgramStatus, error) {
	if id == 0 {
		return incentive.ProgramStatusNotExist, incentive.ErrInvalidProgramID.Wrap("zero")
	}

	statuses := []incentive.ProgramStatus{
		incentive.ProgramStatusUpcoming,
		incentive.ProgramStatusOngoing,
		incentive.ProgramStatusCompleted,
	}

	var found int
	var status incentive.ProgramStatus
	kvStore := k.KVStore(ctx)

	// Looks for an incentive program in upcoming, ongoing, then completed program lists.
	for _, s := range statuses {
		if kvStore.Has(keyIncentiveProgram(id, s)) {
			found++
			status = s
		}
	}

	switch found {
	case 0:
		// If the program was not found in any of the three lists
		return incentive.ProgramStatusNotExist, nil
	case 1:
		// If the program existed
		return status, nil
	default:
		// If the program somehow existed in multiple statuses at once (should never happen)
		return incentive.ProgramStatusNotExist, incentive.ErrInvalidProgramStatus.Wrapf(
			"multiple statuses found for incentive program %d", id,
		)
	}
}

// getIncentiveProgram gets an incentive program by ID, regardless of the program's status.
// Returns the program's status if found, or an error if it does not exist.
func (k Keeper) getIncentiveProgram(ctx sdk.Context, id uint32) (
	incentive.IncentiveProgram, incentive.ProgramStatus, error,
) {
	program := incentive.IncentiveProgram{}
	if id == 0 {
		return program, incentive.ProgramStatusNotExist, incentive.ErrInvalidProgramID.Wrap("zero")
	}

	status, err := k.incentiveProgramStatus(ctx, id)
	if err != nil {
		return program, incentive.ProgramStatusNotExist, err
	}
	if status == incentive.ProgramStatusNotExist {
		return program, incentive.ProgramStatusNotExist, sdkerrors.ErrNotFound.Wrapf("program id %d", id)
	}

	if k.getObject(&ctx, keyIncentiveProgram(id, status), &program, "incentive program") {
		return program, incentive.ProgramStatusNotExist, sdkerrors.ErrNotFound.Wrapf("program id %d", id)
	}
	// Enforces that program ID matches where it was stored
	if program.ID != id {
		return program, incentive.ProgramStatusNotExist, incentive.ErrInvalidProgramID.Wrap("mismatch")
	}
	return program, status, nil
}

// setIncentiveProgram stores an incentive program with a given status.
// this can overwrite existing programs, but will not move them to a new status.
func (k Keeper) setIncentiveProgram(
	ctx sdk.Context, program incentive.IncentiveProgram, status incentive.ProgramStatus,
) error {
	if program.ID == 0 {
		return incentive.ErrInvalidProgramID.Wrap("zero")
	}
	s, err := k.incentiveProgramStatus(ctx, program.ID)
	if err != nil {
		return err
	}
	// error if program exists but not with intended status
	if s != incentive.ProgramStatusNotExist && s != status {
		return sdkerrors.ErrInvalidRequest.Wrapf("program %d already exists with status %d", program.ID, s)
	}

	key := keyIncentiveProgram(program.ID, status)
	return k.setObject(&ctx, key, &program, "incentive program")
}

// deleteIncentiveProgram deletes an incentive program. Returns an error if the program
// did not exist or was found with two different statuses.
func (k Keeper) deleteIncentiveProgram(ctx sdk.Context, id uint32) error {
	status, err := k.incentiveProgramStatus(ctx, id)
	if err != nil {
		return err
	}
	if status == incentive.ProgramStatusNotExist {
		return sdkerrors.ErrNotFound.Wrapf("program %d", id)
	}

	k.KVStore(ctx).Delete(keyIncentiveProgram(id, status))
	return nil
}

// getNextProgramID gets the ID that will be assigned to the next incentive program passed by governance.
func (k Keeper) getNextProgramID(ctx sdk.Context) uint32 {
	return store.GetInteger[uint32](k.KVStore(ctx), keyNextProgramID)
}

// setNextProgramID sets the ID that will be assigned to the next incentive program passed by governance.
func (k Keeper) setNextProgramID(ctx sdk.Context, id uint32) error {
	prev := k.getNextProgramID(ctx)
	if id < prev {
		return incentive.ErrDecreaseNextProgramID.Wrapf("%d to %d", id, prev)
	}
	store.SetInteger(k.KVStore(ctx), keyNextProgramID, id)
	return nil
}

// getLastRewardsTime gets the last unix time incentive rewards were computed globally by EndBlocker.
// panics if it would return a negative value.
func (k Keeper) GetLastRewardsTime(ctx sdk.Context) int64 {
	return store.GetInteger[int64](k.KVStore(ctx), keyLastRewardsTime)
}

// setLastRewardsTime sets the last unix time incentive rewards were computed globally by EndBlocker.
// does not accept negative unix time.
func (k Keeper) setLastRewardsTime(ctx sdk.Context, time int64) error {
	prev := k.GetLastRewardsTime(ctx)
	if time < 0 || time < prev {
		return incentive.ErrDecreaseLastRewardTime.Wrapf("%d to %d", time, prev)
	}
	store.SetInteger(k.KVStore(ctx), keyLastRewardsTime, time)
	return nil
}

// getTotalBonded retrieves the total amount of uTokens of a given denom which are bonded to the incentive module
func (k Keeper) getTotalBonded(ctx sdk.Context, denom string) sdk.Coin {
	key := keyTotalBonded(denom)
	amount := store.GetInt(k.KVStore(ctx), key, "total bonded")
	return sdk.NewCoin(denom, amount)
}

// GetBonded retrieves the amount of uTokens of a given denom which are bonded by an account
func (k Keeper) GetBonded(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	key := keyBondAmount(addr, denom)
	amount := store.GetInt(k.KVStore(ctx), key, "bonded amount")
	return sdk.NewCoin(denom, amount)
}

// setBonded sets the amount of uTokens of a given denom which are bonded by an account.
// Automatically updates TotalBonded as well.
//
// REQUIREMENT: This is the only function which is allowed to set TotalBonded.
func (k Keeper) setBonded(ctx sdk.Context, addr sdk.AccAddress, uToken sdk.Coin) error {
	// compute the change in bonded amount (can be negative when bond decreases)
	delta := uToken.Amount.Sub(k.GetBonded(ctx, addr, uToken.Denom).Amount)

	// Set bond amount
	key := keyBondAmount(addr, uToken.Denom)
	if err := store.SetInt(k.KVStore(ctx), key, uToken.Amount, "bonded amount"); err != nil {
		return err
	}

	// Update total bonded for this utoken denom using the computed change
	total := k.getTotalBonded(ctx, uToken.Denom)
	totalkey := keyTotalBonded(uToken.Denom)
	return store.SetInt(k.KVStore(ctx), totalkey, total.Amount.Add(delta), "total bonded")
}

// getUnbonding retrieves the amount of uTokens of a given denom which are unbonding by an account
func (k Keeper) getUnbondingAmount(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	key := keyUnbondAmount(addr, denom)
	amount := store.GetInt(k.KVStore(ctx), key, "unbonding amount")
	return sdk.NewCoin(denom, amount)
}

// getTotalUnbonding retrieves the total amount of uTokens of a given denom which are unbonding from
// the incentive module
func (k Keeper) getTotalUnbonding(ctx sdk.Context, denom string) sdk.Coin {
	key := keyTotalUnbonding(denom)
	amount := store.GetInt(k.KVStore(ctx), key, "total unbonding")
	return sdk.NewCoin(denom, amount)
}

// getUnbondings gets all unbondings currently associated with an account and bonded denom.
func (k Keeper) getUnbondings(ctx sdk.Context, addr sdk.AccAddress, denom string) []incentive.Unbonding {
	key := keyUnbondings(addr, denom)
	accUnbondings := incentive.AccountUnbondings{}
	if !k.getObject(&ctx, key, &accUnbondings, "account unbondings") {
		return []incentive.Unbonding{}
	}
	return accUnbondings.Unbondings
}

// setUnbondings stores the list of unbondings currently associated with an account and denom.
// It also updates the account's stored unbonding amounts, and thus the module's total unbonding as well.
// Any zero-amount unbondings are automatically filtered out before storage.
//
// REQUIREMENT: This is the only function which is allowed to set unbonding amounts and total unbonding.
func (k Keeper) setUnbondings(ctx sdk.Context, unbondings incentive.AccountUnbondings) error {
	kvStore := k.KVStore(ctx)
	addr, err := sdk.AccAddressFromBech32(unbondings.Account)
	if err != nil {
		// catches invalid and empty addresses
		return err
	}
	denom := unbondings.UToken

	// remove any zero-amount unbondings before setting
	nonzeroUnbondings := []incentive.Unbonding{}
	for _, u := range unbondings.Unbondings {
		if u.UToken.Amount.IsPositive() {
			nonzeroUnbondings = append(nonzeroUnbondings, u)
		}
	}
	unbondings.Unbondings = nonzeroUnbondings

	// compute the new total unbonding specific to this account and denom.
	newUnbonding := sdk.ZeroInt()
	for _, u := range unbondings.Unbondings {
		newUnbonding = newUnbonding.Add(u.UToken.Amount)
	}
	// compute the change in unbonding amount (can be negative when unbonding decreases)
	delta := newUnbonding.Sub(k.getUnbondingAmount(ctx, addr, denom).Amount)

	// Update unbonding amount
	amountKey := keyUnbondAmount(addr, denom)
	if err := store.SetInt(k.KVStore(ctx), amountKey, newUnbonding, "unbonding amount"); err != nil {
		return err
	}

	// Update total unbonding for this utoken denom using the computed change
	total := k.getTotalUnbonding(ctx, denom)
	totalkey := keyTotalUnbonding(denom)
	if err := store.SetInt(k.KVStore(ctx), totalkey, total.Amount.Add(delta), "total unbonding"); err != nil {
		return err
	}

	// set list of unbondings
	key := keyUnbondings(addr, unbondings.UToken)
	if len(unbondings.Unbondings) == 0 {
		// clear store on no unbondings remaining
		kvStore.Delete(key)
	}
	return store.SetValueCdc(kvStore, k.cdc, key, &unbondings, "account unbondings")
}

// getRewardAccumulator retrieves the reward accumulator of all reward tokens for a single bonded uToken -
// for example, how much UMEE, ATOM, etc (reward) would have been earned by 1 ATOM (bonded) since genesis.
func (k Keeper) getRewardAccumulator(ctx sdk.Context, bondDenom string) incentive.RewardAccumulator {
	key := keyRewardAccumulator(bondDenom)
	accumulator := incentive.RewardAccumulator{}
	if k.getObject(&ctx, key, &accumulator, "reward accumulator") {
		return accumulator
	}
	return incentive.NewRewardAccumulator(bondDenom, 0, sdk.NewDecCoins())
}

// setRewardAccumulator sets the full reward accumulator for a single bonded uToken.
func (k Keeper) setRewardAccumulator(ctx sdk.Context, accumulator incentive.RewardAccumulator) error {
	key := keyRewardAccumulator(accumulator.UToken)
	return k.setObject(&ctx, key, &accumulator, "reward accumulator")
}

// GetRewardTracker retrieves the reward tracker for a single bonded uToken on one account - this is the value of
// the reward accumulator for that specific denom the last time this account performed any action that required
// a reward tracker update (i.e. Bond, Claim, BeginUnbonding, or being Liquidated).
func (k Keeper) getRewardTracker(ctx sdk.Context, addr sdk.AccAddress, bondDenom string) incentive.RewardTracker {
	key := keyRewardTracker(addr, bondDenom)
	tracker := incentive.RewardTracker{}
	if k.getObject(&ctx, key, &tracker, "reward tracker") {
		return tracker
	}
	return incentive.NewRewardTracker(addr.String(), bondDenom, sdk.NewDecCoins())
}

// setRewardTracker sets the reward tracker for a single bonded uToken for an address.
// automatically clear the entry if rewards are all zero.
func (k Keeper) setRewardTracker(ctx sdk.Context, tracker incentive.RewardTracker) error {
	key := keyRewardTracker(sdk.MustAccAddressFromBech32(tracker.Account), tracker.UToken)
	if tracker.Rewards.IsZero() {
		k.clearRewardTracker(ctx, sdk.MustAccAddressFromBech32(tracker.Account), tracker.UToken)
		return nil
	}
	return k.setObject(&ctx, key, &tracker, "reward tracker")
}

// clearRewardTracker clears a reward tracker matching a specific account + bonded uToken from the store
func (k Keeper) clearRewardTracker(ctx sdk.Context, addr sdk.AccAddress, bondDenom string) {
	key := keyRewardTracker(addr, bondDenom)
	k.KVStore(ctx).Delete(key)
}
