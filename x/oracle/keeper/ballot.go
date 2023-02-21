package keeper

import (
	"sort"

	"github.com/umee-network/umee/v4/x/oracle/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// OrganizeBallotByDenom collects all oracle votes for the current vote period,
// categorized by the votes' denom parameter.
func (k Keeper) OrganizeBallotByDenom(
	ctx sdk.Context,
	validatorClaimMap map[string]types.Claim,
) []types.BallotDenom {
	votes := map[string]types.ExchangeRateBallot{}

	// collect aggregate votes
	aggregateHandler := func(voterAddr sdk.ValAddress, vote types.AggregateExchangeRateVote) bool {
		// organize ballot only for the active validators
		claim, ok := validatorClaimMap[vote.Voter]
		if ok {
			for _, tuple := range vote.ExchangeRateTuples {
				votes[tuple.Denom] = append(
					votes[tuple.Denom],
					types.NewVoteForTally(tuple.ExchangeRate, tuple.Denom, voterAddr, claim.Power),
				)
			}
		}
		return false
	}

	k.IterateAggregateExchangeRateVotes(ctx, aggregateHandler)

	// sort created ballots
	for denom, ballot := range votes {
		sort.Sort(ballot)
		votes[denom] = ballot
	}
	return types.BallotMapToSlice(votes)
}

// ClearVotes clears all tallied prevotes and votes from the store.
func (k Keeper) ClearVotes(ctx sdk.Context, votePeriod uint64) {
	currentHeight := uint64(ctx.BlockHeight())
	k.IterateAggregateExchangeRatePrevotes(
		ctx,
		func(voterAddr sdk.ValAddress, aggPrevote types.AggregateExchangeRatePrevote) bool {
			if currentHeight > (aggPrevote.SubmitBlock + votePeriod) {
				k.DeleteAggregateExchangeRatePrevote(ctx, voterAddr)
			}

			return false
		},
	)

	// clear all aggregate votes
	k.IterateAggregateExchangeRateVotes(
		ctx,
		func(voterAddr sdk.ValAddress, _ types.AggregateExchangeRateVote) bool {
			k.DeleteAggregateExchangeRateVote(ctx, voterAddr)
			return false
		},
	)
}
