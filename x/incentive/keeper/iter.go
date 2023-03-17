package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v4/util/store"
	"github.com/umee-network/umee/v4/x/incentive"
)

// getAllIncentivePrograms returns all incentive programs
// that have been passed by governance and have a particular status.
// The status of an incentive program is either Upcoming, Ongoing, or Completed.
func (k Keeper) getAllIncentivePrograms(ctx sdk.Context, status incentive.ProgramStatus,
) ([]incentive.IncentiveProgram, error) {
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
		return []incentive.IncentiveProgram{}, incentive.ErrInvalidProgramStatus
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

	err := store.Iterate(k.KVStore(ctx), prefix, iterator)
	return programs, err
}

// getPaginatedIncentivePrograms returns all incentive programs
// that have been passed by governance and have a particular status.
// The status of an incentive program is either Upcoming, Ongoing, or Completed.
// Accepts pagination parameters which specify the length of a page and which page to fetch.
func (k Keeper) getPaginatedIncentivePrograms(
	ctx sdk.Context, status incentive.ProgramStatus, page, limit uint64,
) ([]incentive.IncentiveProgram, error) {
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
		return []incentive.IncentiveProgram{}, incentive.ErrInvalidProgramStatus
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

	err := store.IteratePaginated(k.KVStore(ctx), prefix, uint(page), uint(limit), iterator)
	return programs, err
}
