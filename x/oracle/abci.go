package oracle

import (
	"fmt"
	"strings"
	"time"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v6/x/oracle/keeper"
	"github.com/umee-network/umee/v6/x/oracle/types"
)

// EndBlocker is called at the end of every block
func EndBlocker(ctx sdk.Context, k keeper.Keeper) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	params := k.GetParams(ctx)
	if k.IsPeriodLastBlock(ctx, params.VotePeriod) {
		if err := CalcPrices(ctx, params, k); err != nil {
			return err
		}
	}

	// Slash oracle providers who missed voting over the threshold and
	// reset miss counters of all validators at the last block of slash window
	if k.IsPeriodLastBlock(ctx, params.SlashWindow) {
		k.SlashAndResetMissCounters(ctx)
	}

	k.PruneAllPrices(ctx)

	return nil
}

func CalcPrices(ctx sdk.Context, params types.Params, k keeper.Keeper) error {
	// Build claim map over all validators in active set
	validatorClaimMap := make(map[string]types.Claim)
	powerReduction := k.StakingKeeper.PowerReduction(ctx)
	// Calculate total validator power
	var totalBondedPower int64
	for _, v := range k.StakingKeeper.GetBondedValidatorsByPower(ctx) {
		addr := v.GetOperator()
		power := v.GetConsensusPower(powerReduction)
		totalBondedPower += power
		validatorClaimMap[addr.String()] = types.NewClaim(power, 0, 0, addr)
	}

	// voteTargets defines the symbol (ticker) denoms that we require votes on
	voteTargets := make(map[string]bool, 0)
	voteTargetDenoms := make([]string, 0)
	for _, v := range params.AcceptList {
		voteTargets[v.SymbolDenom] = true // unique symbol denoms <Note: we are allowing duplicate symbol denoms>
		voteTargetDenoms = append(voteTargetDenoms, v.BaseDenom)
	}

	// NOTE: it filters out inactive or jailed validators
	// ballotDenomSlice is oracle votes of the symbol denoms, those are stored by AggregateExchangeRateVote
	ballotDenomSlice := k.OrganizeBallotByDenom(ctx, validatorClaimMap)
	threshold := k.VoteThreshold(ctx).MulInt64(types.MaxVoteThresholdMultiplier).TruncateInt64()

	// Iterate through ballots and update exchange rates; drop if not enough votes have been achieved.
	for _, ballotDenom := range ballotDenomSlice {
		// Calculate the portion of votes received as an integer, scaled up using the
		// same multiplier as the `threshold` computed above
		support := ballotDenom.Ballot.Power() * types.MaxVoteThresholdMultiplier / totalBondedPower
		if support < threshold {
			// ctx.Logger().Info("Ballot voting power is under vote threshold, dropping ballot", "denom", ballotDenom)
			for _, v := range ballotDenom.Ballot {
				fmt.Print("mark C")
				fmt.Println("not enough support has been achieved for an asset")
				fmt.Println("denom:", v.Denom)
				fmt.Println("power in ballot:", v.Power)
				fmt.Println("support:", support)
				fmt.Println("threshold:", threshold)

				// get votes for each of the major 4 validators
				val1, _ := sdk.ValAddressFromBech32("umeevaloper1hd4zwgmu2uax5pp43mqdnn09720qmvuqjenkr4")
				val2, _ := sdk.ValAddressFromBech32("umeevaloper1aswcvpju7ff2gturf2r6whx25p0j0apu5kuvg0")
				val3, _ := sdk.ValAddressFromBech32("umeevaloper16952kyr062e7l0a7ssxfzmur05tjmknvh7kax3")
				vote, err := k.GetAggregateExchangeRateVote(ctx, val1)
				if err != nil || len(vote.ExchangeRateTuples) == 0 {
					fmt.Println("val 1 did not vote!")
				} else {
					fmt.Println("val 1 voted!")
				}
				vote, err = k.GetAggregateExchangeRateVote(ctx, val2)
				if err != nil || len(vote.ExchangeRateTuples) == 0 {
					fmt.Println("val 2 did not vote!")
				} else {
					fmt.Println("val 2 voted!")
				}
				vote, err = k.GetAggregateExchangeRateVote(ctx, val3)
				if err != nil || len(vote.ExchangeRateTuples) == 0 {
					fmt.Println("val 3 did not vote!")
				} else {
					fmt.Println("val 3 voted!")
				}

			}

			continue
		}

		fmt.Println("mark D")
		fmt.Println("ballot made it through!")
		fmt.Println("denom:", ballotDenom.Denom)
		fmt.Println("support:", support)
		fmt.Println("threshold:", threshold)

		denom := strings.ToUpper(ballotDenom.Denom)
		// Get weighted median of exchange rates
		exchangeRate, err := Tally(ballotDenom.Ballot, params.RewardBand, validatorClaimMap)
		if err != nil {
			return err
		}
		// save the exchange rate to store with denom and timestamp
		k.SetExchangeRate(ctx, denom, exchangeRate)

		if k.IsPeriodLastBlock(ctx, params.HistoricStampPeriod) {
			k.AddHistoricPrice(ctx, denom, exchangeRate)
		}

		// Calculate and stamp median/median deviation if median stamp period has passed
		if k.IsPeriodLastBlock(ctx, params.MedianStampPeriod) {
			if err = k.CalcAndSetHistoricMedian(ctx, denom); err != nil {
				return err
			}
		}
	}

	// update miss counting & slashing
	voteTargetsLen := len(voteTargets)
	claimSlice := types.ClaimMapToSlice(validatorClaimMap)
	for _, claim := range claimSlice {
		// Skip valid voters
		// in MsgAggregateExchangeRateVote we filter tokens from the AcceptList.
		if int(claim.TokensVoted) == voteTargetsLen {
			continue
		}

		if claim.Validator.String() == "umeevaloper13s987upqqy3cqkvy4gtclwej2kljg39qyajv7m" {
			fmt.Println("mark A")
			fmt.Println(claim.Validator.String())
			fmt.Println(claim.TokensVoted)
			fmt.Println(voteTargetsLen)
		}

		// Increase miss counter
		k.SetMissCounter(ctx, claim.Validator, k.GetMissCounter(ctx, claim.Validator)+1)
	}

	// Distribute rewards to ballot winners
	k.RewardBallotWinners(
		ctx,
		int64(params.VotePeriod),
		int64(params.RewardDistributionWindow),
		voteTargetDenoms,
		claimSlice,
	)

	k.ClearVotes(ctx, params.VotePeriod)
	return nil
}

// Tally calculates and returns the median. It sets the set of voters to be
// rewarded, i.e. voted within a reasonable spread from the weighted median to
// the store. Note, the ballot is sorted by ExchangeRate.
func Tally(
	ballot types.ExchangeRateBallot,
	rewardBand sdk.Dec,
	validatorClaimMap map[string]types.Claim,
) (sdk.Dec, error) {
	weightedMedian, err := ballot.WeightedMedian()
	if err != nil {
		return sdk.ZeroDec(), err
	}
	standardDeviation, err := ballot.StandardDeviation()
	if err != nil {
		return sdk.ZeroDec(), err
	}

	// rewardSpread is the MAX((weightedMedian * (rewardBand/2)), standardDeviation)
	rewardSpread := weightedMedian.Mul(rewardBand.QuoInt64(2))
	rewardSpread = sdk.MaxDec(rewardSpread, standardDeviation)

	for _, tallyVote := range ballot {
		// Filter ballot winners. For voters, we filter out the tally vote iff:
		// (weightedMedian - rewardSpread) <= ExchangeRate <= (weightedMedian + rewardSpread)
		if (tallyVote.ExchangeRate.GTE(weightedMedian.Sub(rewardSpread)) &&
			tallyVote.ExchangeRate.LTE(weightedMedian.Add(rewardSpread))) ||
			!tallyVote.ExchangeRate.IsPositive() {

			key := tallyVote.Voter.String()
			claim := validatorClaimMap[key]

			claim.Weight += tallyVote.Power
			claim.TokensVoted++
			validatorClaimMap[key] = claim
		} else {
			if tallyVote.Voter.String() == "umeevaloper13s987upqqy3cqkvy4gtclwej2kljg39qyajv7m" {
				fmt.Print("mark B")
				fmt.Println(tallyVote.Voter.String())
				fmt.Println(tallyVote.ExchangeRate)
				fmt.Println(weightedMedian.Sub(rewardSpread))
				fmt.Println(weightedMedian.Add(rewardSpread))
			}
		}
	}

	return weightedMedian, nil
}
