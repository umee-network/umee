package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v6/util/keys"
	"github.com/umee-network/umee/v6/util/store"
	"github.com/umee-network/umee/v6/x/incentive"
)

// getAllIncentivePrograms returns all incentive programs
// that have been passed by governance and have a particular status.
// The status of an incentive program is either Upcoming, Ongoing, or Completed.
func (k Keeper) getAllIncentivePrograms(
	ctx sdk.Context, status incentive.ProgramStatus,
) ([]incentive.IncentiveProgram, error) {
	var prefix []byte
	switch status {
	case incentive.ProgramStatusUpcoming:
		prefix = keyPrefixUpcomingIncentiveProgram
	case incentive.ProgramStatusOngoing:
		prefix = keyPrefixOngoingIncentiveProgram
	case incentive.ProgramStatusCompleted:
		prefix = keyPrefixCompletedIncentiveProgram
	default:
		return []incentive.IncentiveProgram{}, incentive.ErrInvalidProgramStatus
	}

	return store.LoadAll[*incentive.IncentiveProgram](k.KVStore(ctx), prefix)
}

// getIncentiveProgramsByStatusAndDenom returns all incentive programs with a specified status and uToken denom.
func (k Keeper) getIncentiveProgramsByStatusAndDenom(
	ctx sdk.Context,
	status incentive.ProgramStatus,
	denom string,
) ([]incentive.IncentiveProgram, error) {
	var result []incentive.IncentiveProgram
	programs, err := k.getAllIncentivePrograms(ctx, status)
	if err != nil {
		return nil, err
	}

	for _, program := range programs {
		if program.UToken == denom {
			result = append(result, program)
		}
	}

	return result, nil

}

func (k Keeper) HasOngoingProgramsForDenom(ctx sdk.Context, denom string) (bool, error) {
	programs, err := k.getIncentiveProgramsByStatusAndDenom(ctx, incentive.ProgramStatusOngoing, denom)
	if err != nil {
		return false, err
	}

	return len(programs) > 0, nil
}

// getAllBondDenoms gets all uToken denoms for which an account has nonzero bonded amounts.
// useful for setting up queries which look at all of an account's bonds or unbondings.
func (k Keeper) getAllBondDenoms(ctx sdk.Context, addr sdk.AccAddress) ([]string, error) {
	prefix := keyBondAmountNoDenom(addr)
	bonds := []string{}

	iterator := func(key, _ []byte) error {
		_, denom, _, err := keys.ExtractAddressAndString(len(keyPrefixBondAmount), key)
		if err != nil {
			return err
		}
		bonds = append(bonds, denom)
		return nil
	}

	err := store.Iterate(k.KVStore(ctx), prefix, iterator)
	return bonds, err
}

// getAllBonds gets all bonds for all accounts (used during export genesis)
func (k Keeper) getAllBonds(ctx sdk.Context) ([]incentive.Bond, error) {
	prefix := keyPrefixBondAmount
	bonds := []incentive.Bond{}

	iterator := func(key, val []byte) error {
		addr, denom, _, err := keys.ExtractAddressAndString(len(keyPrefixBondAmount), key)
		if err != nil {
			return err
		}
		amount := store.Int(val, "bond amount")
		bonds = append(bonds, incentive.NewBond(
			addr.String(),
			sdk.NewCoin(denom, amount),
		))

		return nil
	}

	err := store.Iterate(k.KVStore(ctx), prefix, iterator)
	return bonds, err
}

// getAllRewardTrackers gets all reward trackers for all accounts (used during export genesis)
func (k Keeper) getAllRewardTrackers(ctx sdk.Context) ([]incentive.RewardTracker, error) {
	return store.LoadAll[*incentive.RewardTracker](k.KVStore(ctx), keyPrefixRewardTracker)
}

// getAllRewardAccumulators gets all reward accumulators for all uTokens (used during export genesis)
func (k Keeper) getAllRewardAccumulators(ctx sdk.Context) ([]incentive.RewardAccumulator, error) {
	return store.LoadAll[*incentive.RewardAccumulator](k.KVStore(ctx), keyPrefixRewardAccumulator)
}

// getAllAccountUnbondings gets all account unbondings for all accounts (used during export genesis)
func (k Keeper) getAllAccountUnbondings(ctx sdk.Context) ([]incentive.AccountUnbondings, error) {
	return store.LoadAll[*incentive.AccountUnbondings](k.KVStore(ctx), keyPrefixUnbondings)
}

// getAllTotalUnbonding gets total unbonding for all uTokens (used for a query)
func (k Keeper) getAllTotalUnbonding(ctx sdk.Context) sdk.Coins {
	return store.SumCoins(k.prefixStore(ctx, keyPrefixTotalUnbonding), keys.ToStr)
}

// getAllTotalBonded gets total bonded for all uTokens (used for a query)
func (k Keeper) getAllTotalBonded(ctx sdk.Context) sdk.Coins {
	return store.SumCoins(k.prefixStore(ctx, keyPrefixTotalBonded), keys.ToStr)
}
