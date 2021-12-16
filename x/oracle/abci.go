package oracle

import (
	"time"

	"github.com/umee-network/umee/x/oracle/keeper"
	"github.com/umee-network/umee/x/oracle/types"

	"github.com/cosmos/cosmos-sdk/telemetry"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// IsPeriodLastBlock returns true if we are at the last block of the period
func IsPeriodLastBlock(ctx sdk.Context, blocksPerPeriod uint64) bool {
	return (uint64(ctx.BlockHeight())+1)%blocksPerPeriod == 0
}

// EndBlocker is called at the end of every block
func EndBlocker(ctx sdk.Context, k keeper.Keeper) error {
	defer telemetry.ModuleMeasureSince(types.ModuleName, time.Now(), telemetry.MetricKeyEndBlocker)

	params := k.GetParams(ctx)
	if IsPeriodLastBlock(ctx, params.VotePeriod) {
		// Build claim map over all validators in active set
		validatorClaimMap := make(map[string]types.Claim)

		maxValidators := k.StakingKeeper.MaxValidators(ctx)
		iterator := k.StakingKeeper.ValidatorsPowerStoreIterator(ctx)
		defer iterator.Close()

		powerReduction := k.StakingKeeper.PowerReduction(ctx)

		i := 0
		for ; iterator.Valid() && i < int(maxValidators); iterator.Next() {
			validator := k.StakingKeeper.Validator(ctx, iterator.Value())

			// Exclude not bonded validator
			if validator.IsBonded() {
				valAddr := validator.GetOperator()
				validatorClaimMap[valAddr.String()] = types.NewClaim(validator.GetConsensusPower(powerReduction), 0, 0, valAddr)
				i++
			}
		}

		var (
			// voteTargets defines the symbol (ticker) denoms that we require votes on
			voteTargets      []string
			voteTargetDenoms []string
		)
		for _, v := range params.AcceptList {
			voteTargets = append(voteTargets, v.SymbolDenom)
			voteTargetDenoms = append(voteTargetDenoms, v.BaseDenom)
		}

		// Clear all exchange rates
		k.IterateExchangeRates(ctx, func(denom string, _ sdk.Dec) (stop bool) {
			k.DeleteExchangeRate(ctx, denom)
			return false
		})

		// Organize votes to ballot by denom
		// NOTE: **Filter out inactive or jailed validators**
		// NOTE: **Make abstain votes to have zero vote power**
		voteMap := k.OrganizeBallotByDenom(ctx, validatorClaimMap)

		ballotDenomSlice := types.BallotMapToSlice(voteMap)

		// Iterate through ballots and update exchange rates; drop if not enough votes have been achieved.
		for _, ballotDenom := range ballotDenomSlice {
			// Get weighted median of exchange rates
			exchangeRate, err := Tally(ctx, ballotDenom.Ballot, params.RewardBand, validatorClaimMap)
			if err != nil {
				return err
			}

			// Set the exchange rate, emit ABCI event
			k.SetExchangeRateWithEvent(ctx, ballotDenom.Denom, exchangeRate)
		}

		// update miss counting & slashing
		voteTargetsLen := len(voteTargets)
		claimSlice := types.ClaimMapToSlice(validatorClaimMap)
		for _, claim := range claimSlice {
			// Skip abstain & valid voters
			if int(claim.WinCount) == voteTargetsLen {
				continue
			}

			// Increase miss counter
			k.SetMissCounter(ctx, claim.Recipient, k.GetMissCounter(ctx, claim.Recipient)+1)
		}

		// Distribute rewards to ballot winners
		k.RewardBallotWinners(
			ctx,
			int64(params.VotePeriod),
			int64(params.RewardDistributionWindow),
			voteTargetDenoms,
			validatorClaimMap,
		)

		// Clear the ballot
		k.ClearBallots(ctx, params.VotePeriod)

		// Update vote targets
		k.ApplyAcceptList(ctx, params.AcceptList, voteTargets)
	}

	// Slash oracle providers who missed voting over the threshold and
	// reset miss counters of all validators at the last block of slash window
	if IsPeriodLastBlock(ctx, params.SlashWindow) {
		k.SlashAndResetMissCounters(ctx)
	}

	return nil
}

// Tally calculates the median and returns it. Sets the set of voters to be rewarded, i.e. voted within
// a reasonable spread from the weighted median to the store
func Tally(ctx sdk.Context,
	pb types.ExchangeRateBallot,
	rewardBand sdk.Dec,
	validatorClaimMap map[string]types.Claim,
) (sdk.Dec, error) {
	weightedMedian := pb.WeightedMedian()
	standardDeviation, err := pb.StandardDeviation()
	if err != nil {
		return sdk.ZeroDec(), err
	}

	rewardSpread := weightedMedian.Mul(rewardBand.QuoInt64(2))

	if standardDeviation.GT(rewardSpread) {
		rewardSpread = standardDeviation
	}

	for _, vote := range pb {
		// Filter ballot winners & abstain voters
		if (vote.ExchangeRate.GTE(weightedMedian.Sub(rewardSpread)) &&
			vote.ExchangeRate.LTE(weightedMedian.Add(rewardSpread))) ||
			!vote.ExchangeRate.IsPositive() {

			key := vote.Voter.String()
			claim := validatorClaimMap[key]
			claim.Weight += vote.Power
			claim.WinCount++
			validatorClaimMap[key] = claim
		}
	}

	return weightedMedian, nil
}
