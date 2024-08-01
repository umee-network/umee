package keeper

import (
	"fmt"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v6/util/genmap"
	"github.com/umee-network/umee/v6/x/oracle/types"
)

// prependUmeeIfUnique pushs `uumee` denom to the front of the list, if it is not yet included.
func prependUmeeIfUnique(voteTargets []string) []string {
	if genmap.Contains(types.UmeeDenom, voteTargets) {
		return voteTargets
	}
	rewardDenoms := make([]string, len(voteTargets)+1)
	rewardDenoms[0] = types.UmeeDenom
	copy(rewardDenoms[1:], voteTargets)
	return rewardDenoms
}

// RewardBallotWinners is executed at the end of every voting period, where we
// give out a portion of seigniorage reward(reward-weight) to the oracle voters
// that voted correctly.
func (k Keeper) RewardBallotWinners(
	ctx sdk.Context,
	votePeriod int64,
	rewardDistributionWindow int64,
	voteTargets []string,
	ballotWinners []types.Claim,
) {
	// sum weight of the claims
	var ballotPowerSum int64
	for _, winner := range ballotWinners {
		ballotPowerSum += winner.Weight
	}

	// early return - ballot was empty
	if ballotPowerSum == 0 {
		return
	}

	distributionRatio := sdkmath.LegacyNewDec(votePeriod).QuoInt64(rewardDistributionWindow)
	var periodRewards sdk.DecCoins
	rewardDenoms := prependUmeeIfUnique(voteTargets)
	for _, denom := range rewardDenoms {
		rewardPool := k.GetRewardPool(ctx, denom)

		// return if there's no rewards to give out
		if rewardPool.IsZero() {
			continue
		}

		periodRewards = periodRewards.Add(sdk.NewDecCoinFromDec(
			denom,
			sdkmath.LegacyNewDecFromInt(rewardPool.Amount).Mul(distributionRatio),
		))
	}

	// distribute rewards
	var distributedReward sdk.Coins

	for _, winner := range ballotWinners {
		receiverVal, err := k.StakingKeeper.Validator(ctx, winner.Validator)
		if err != nil {
			panic(fmt.Errorf("failed to get validator %s and error %w", winner.Validator.String(), err))
		}
		// in case absence of the validator, we just skip distribution
		if receiverVal == nil {
			continue
		}

		// reflects contribution
		rewardCoins, _ := periodRewards.MulDec(sdkmath.LegacyNewDec(winner.Weight).QuoInt64(ballotPowerSum)).TruncateDecimal()
		if rewardCoins.IsZero() {
			continue
		}

		err = k.distrKeeper.AllocateTokensToValidator(ctx, receiverVal, sdk.NewDecCoinsFromCoins(rewardCoins...))
		if err != nil {
			panic(fmt.Errorf("failed to allocate tokens to validator %w", err))
		}
		distributedReward = distributedReward.Add(rewardCoins...)
	}

	// move distributed reward to distribution module
	err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, k.distrName, distributedReward)
	if err != nil {
		panic(fmt.Errorf("failed to send coins to distribution module %w", err))
	}
}
