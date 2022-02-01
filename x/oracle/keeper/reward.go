package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/x/oracle/types"
)

// RewardBallotWinners is executed at the end of every voting period, where we
// give out a portion of seigniorage reward(reward-weight) to the oracle voters
// that voted correctly.
func (k Keeper) RewardBallotWinners(
	ctx sdk.Context,
	votePeriod int64,
	rewardDistributionWindow int64,
	voteTargets []string,
	ballotWinners map[string]types.Claim,
) {
	rewardDenoms := make([]string, len(voteTargets)+1)
	rewardDenoms[0] = types.UmeeDenom

	i := 1
	for _, denom := range voteTargets {
		rewardDenoms[i] = denom
		i++
	}

	// sum weight of the claims
	var ballotPowerSum int64
	for _, winner := range ballotWinners {
		ballotPowerSum += winner.Weight
	}

	// return if the ballot is empty
	if ballotPowerSum == 0 {
		return
	}

	distributionRatio := sdk.NewDec(votePeriod).QuoInt64(rewardDistributionWindow)

	var periodRewards sdk.DecCoins
	for _, denom := range rewardDenoms {
		rewardPool := k.GetRewardPool(ctx, denom)

		// return if there's no rewards to give out
		if rewardPool.IsZero() {
			continue
		}

		periodRewards = periodRewards.Add(sdk.NewDecCoinFromDec(
			denom,
			sdk.NewDecFromInt(rewardPool.Amount).Mul(distributionRatio),
		))
	}

	// distribute rewards
	var distributedReward sdk.Coins
	for _, winner := range ballotWinners {
		receiverVal := k.StakingKeeper.Validator(ctx, winner.Recipient)
		// in case absence of the validator, we just skip distribution
		if receiverVal == nil {
			continue
		}

		// reflects contribution
		rewardCoins, _ := periodRewards.MulDec(sdk.NewDec(winner.Weight).QuoInt64(ballotPowerSum)).TruncateDecimal()
		if rewardCoins.IsZero() {
			continue
		}

		k.distrKeeper.AllocateTokensToValidator(ctx, receiverVal, sdk.NewDecCoinsFromCoins(rewardCoins...))
		distributedReward = distributedReward.Add(rewardCoins...)
	}

	// move distributed reward to distribution module
	err := k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, k.distrName, distributedReward)
	if err != nil {
		panic(fmt.Sprintf("failed to send coins to distribution module %s", err.Error()))
	}
}
