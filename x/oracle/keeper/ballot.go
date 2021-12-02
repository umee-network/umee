package keeper

import (
	"sort"

	"github.com/umee-network/umee/x/oracle/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// OrganizeBallotByDenom collects all oracle votes for the current vote period,
// categorized by the votes' denom parameter.
func (k Keeper) OrganizeBallotByDenom(
	ctx sdk.Context,
	validatorClaimMap map[string]types.Claim,
) (votes map[string]types.ExchangeRateBallot) {

	votes = map[string]types.ExchangeRateBallot{}

	// collect aggregate votes
	aggregateHandler := func(voterAddr sdk.ValAddress, vote types.AggregateExchangeRateVote) bool {
		// organize ballot only for the active validators
		claim, ok := validatorClaimMap[vote.Voter]
		if ok {
			power := claim.Power

			for _, tuple := range vote.ExchangeRateTuples {
				tmpPower := power
				if !tuple.ExchangeRate.IsPositive() {
					// make the power of abstain vote zero
					tmpPower = 0
				}

				votes[tuple.Denom] = append(
					votes[tuple.Denom],
					types.NewVoteForTally(tuple.ExchangeRate, tuple.Denom, voterAddr, tmpPower),
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

	return votes
}

// ClearBallots clears all tallied prevotes and votes from the store.
func (k Keeper) ClearBallots(ctx sdk.Context, votePeriod uint64) {
	// clear all aggregate prevotes
	k.IterateAggregateExchangeRatePrevotes(
		ctx,
		func(voterAddr sdk.ValAddress, aggPrevote types.AggregateExchangeRatePrevote) bool {
			if ctx.BlockHeight() > int64(aggPrevote.SubmitBlock+votePeriod) {
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

// ApplyWhitelist updates vote targets denom list and set tobin tax with params
// whitelist.
func (k Keeper) ApplyWhitelist(
	ctx sdk.Context,
	whitelist types.DenomList,
	voteTargets map[string]sdk.Dec,
) {
	// check is there any update in whitelist params
	updateRequired := false
	if len(voteTargets) != len(whitelist) {
		updateRequired = true
	} else {
		for _, item := range whitelist {
			tobinTax, ok := voteTargets[item.Name]
			if !ok || !tobinTax.Equal(item.TobinTax) {
				updateRequired = true
				break
			}
		}
	}

	if updateRequired {
		k.ClearTobinTaxes(ctx)

		for _, item := range whitelist {
			k.SetTobinTax(ctx, item.Name, item.TobinTax)

			// register metadata to bank module
			if _, ok := k.bankKeeper.GetDenomMetaData(ctx, item.Name); !ok {
				base := item.Name
				display := base[1:]

				k.bankKeeper.SetDenomMetaData(
					ctx,
					banktypes.Metadata{
						Description: "The national currency of the United States",
						DenomUnits: []*banktypes.DenomUnit{
							{Denom: base, Exponent: uint32(0), Aliases: []string{}},
							{Denom: display, Exponent: uint32(6), Aliases: []string{}},
						},
						Base:    base,
						Display: display,
						Name:    display,
						Symbol:  display,
						// TODO: The Gravity bridge requires that name, symbol and display
						// are all equal. However, it is not currently clear if these assets
						// will be bridged across to Ethereum. If not, we can uncomment below.
						//
						// ref: https://github.com/umee-network/umee/issues/225
						//
						// Name:    fmt.Sprintf("%s United States Dollar", strings.ToUpper(display)),
						// Symbol:  fmt.Sprintf("%sUSD", strings.ToUpper(display[:len(display)-1])),
					},
				)
			}
		}
	}
}
