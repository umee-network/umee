package oracle

import (
	"strings"
	"time"

	sdkmath "cosmossdk.io/math"
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
	validators, err := k.StakingKeeper.GetBondedValidatorsByPower(ctx)
	if err != nil {
		return err
	}
	for _, v := range validators {
		addr := v.GetOperator()
		power := v.GetConsensusPower(powerReduction)
		totalBondedPower += power
		valAddr, err := sdk.ValAddressFromBech32(addr)
		if err != nil {
			return err
		}
		validatorClaimMap[addr] = types.NewClaim(power, 0, 0, valAddr)
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
			ctx.Logger().Info("Ballot voting power is under vote threshold, dropping ballot", "denom", ballotDenom)
			continue
		}

		denom := strings.ToUpper(ballotDenom.Denom)
		// Get weighted median of exchange rates
		exchangeRate, err := Tally(ballotDenom.Ballot, params.RewardBand, validatorClaimMap)
		if err != nil {
			return err
		}
		// save the exchange rate to store with denom and timestamp
		k.SetExchangeRate(ctx, denom, exchangeRate)
	}

	if k.IsPeriodLastBlock(ctx, params.HistoricStampPeriod) {
		k.IterateExchangeRates(ctx, func(denom string, exgRate sdkmath.LegacyDec, _ time.Time) (stop bool) {
			k.AddHistoricPrice(ctx, denom, exgRate)
			return false
		})
	}
	// Calculate and stamp median/median deviation if median stamp period has passed
	if k.IsPeriodLastBlock(ctx, params.MedianStampPeriod) {
		var err error
		k.IterateExchangeRates(ctx, func(denom string, _ sdkmath.LegacyDec, _ time.Time) (stop bool) {
			err = k.CalcAndSetHistoricMedian(ctx, denom)
			return err != nil
		})
		if err != nil {
			return err
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
	rewardBand sdkmath.LegacyDec,
	validatorClaimMap map[string]types.Claim,
) (sdkmath.LegacyDec, error) {
	weightedMedian, err := ballot.WeightedMedian()
	if err != nil {
		return sdkmath.LegacyZeroDec(), err
	}
	standardDeviation, err := ballot.StandardDeviation()
	if err != nil {
		return sdkmath.LegacyZeroDec(), err
	}

	// rewardSpread is the MAX((weightedMedian * (rewardBand/2)), standardDeviation)
	rewardSpread := weightedMedian.Mul(rewardBand.QuoInt64(2))
	rewardSpread = sdkmath.LegacyMaxDec(rewardSpread, standardDeviation)

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
		}
	}

	return weightedMedian, nil
}
