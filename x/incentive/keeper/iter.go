package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v4/util"
	"github.com/umee-network/umee/v4/util/store"
	"github.com/umee-network/umee/v4/x/incentive"
)

// GetAllIncentivePrograms returns all incentive programs
// that have been passed by governance and have a particular status.
// The status of an incentive program is either Upcoming, Ongoing, or Completed.
func (k Keeper) GetAllIncentivePrograms(ctx sdk.Context, status incentive.ProgramStatus) []incentive.IncentiveProgram {
	programs := []incentive.IncentiveProgram{}

	var prefix []byte
	switch status {
	case incentive.ProgramStatusUpcoming:
		prefix = keyPrefixUpcomingIncentiveProgram
	case incentive.ProgramStatusOngoing:
		prefix = keyPrefixOngoingIncentiveProgram
	case incentive.ProgramStatusCompleted:
		prefix = keyPrefixCompletedIncentiveProgram
	default:
		return []incentive.IncentiveProgram{}
	}

	iterator := func(_, val []byte) error {
		var p incentive.IncentiveProgram
		if err := p.Unmarshal(val); err != nil {
			// improperly marshaled IncentiveProgram should never happen
			return err
		}

		programs = append(programs, p)
		return nil
	}

	util.Panic(store.Iterate(k.KVStore(ctx), prefix, iterator))
	return programs
}

// iterateAccountBonds iterates over all bonded uTokens for an address by each individual
// uToken denom and tier
func (k Keeper) iterateAccountBonds(ctx sdk.Context, addr sdk.AccAddress,
	_ func(ctx sdk.Context, addr sdk.AccAddress, _ incentive.BondTier, _ sdk.Coin) error,
) error {
	prefix := keyBondAmountNoDenom(addr)

	iterator := func(key, val []byte) error {
		// TODO: implement (used for reward claiming and also queries)
		return incentive.ErrNotImplemented
	}

	return store.Iterate(k.KVStore(ctx), prefix, iterator)
}

// getFullRewardTracker combines all single-reward-denom reward trackers for a bonded uToken and tier
func (k Keeper) getFullRewardTracker(ctx sdk.Context, addr sdk.AccAddress, denom string, tier incentive.BondTier,
) sdk.DecCoins {
	prefix := keyRewardTrackerNoReward(addr, denom, tier)

	tracker := sdk.NewDecCoins()
	iterator := func(key, val []byte) error {
		// TODO: implement

		return incentive.ErrNotImplemented
	}

	util.Panic(store.Iterate(k.KVStore(ctx), prefix, iterator))
	return tracker
}

// getFullRewardAccumulator combines all single-reward-denom reward accumulators for a uToken denom and tier
func (k Keeper) getFullRewardAccumulator(ctx sdk.Context, denom string, tier incentive.BondTier) sdk.DecCoins {
	prefix := keyRewardAccumulatorNoReward(denom, tier)

	accumulator := sdk.NewDecCoins()
	iterator := func(key, val []byte) error {
		// TODO: implement

		return incentive.ErrNotImplemented
	}

	util.Panic(store.Iterate(k.KVStore(ctx), prefix, iterator))
	return accumulator
}

// GetPaginatedIncentivePrograms returns all incentive programs
// that have been passed by governance and have a particular status.
// The status of an incentive program is either Upcoming, Ongoing, or Completed.
// Accepts pagination parameters which specify the length of a page and which page to fetch.
func (k Keeper) GetPaginatedIncentivePrograms(
	ctx sdk.Context, status incentive.ProgramStatus, page, limit uint64,
) []incentive.IncentiveProgram {
	programs := []incentive.IncentiveProgram{}

	var prefix []byte
	switch status {
	case incentive.ProgramStatusUpcoming:
		prefix = keyPrefixUpcomingIncentiveProgram
	case incentive.ProgramStatusOngoing:
		prefix = keyPrefixOngoingIncentiveProgram
	case incentive.ProgramStatusCompleted:
		prefix = keyPrefixCompletedIncentiveProgram
	default:
		return []incentive.IncentiveProgram{}
	}

	iterator := func(_, val []byte) error {
		var p incentive.IncentiveProgram
		if err := p.Unmarshal(val); err != nil {
			// improperly marshaled IncentiveProgram should never happen
			return err
		}

		programs = append(programs, p)
		return nil
	}

	util.Panic(store.IteratePaginated(k.KVStore(ctx), prefix, uint(page), uint(limit), iterator))
	return programs
}
