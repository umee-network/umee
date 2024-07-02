package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v6/util/sdkutil"
	"github.com/umee-network/umee/v6/x/oracle/types"
)

// SlashAndResetMissCounters iterates over all the current missed counters and
// calculates the "valid vote rate" as:
// (votePeriodsPerWindow - missCounter)/votePeriodsPerWindow.
//
// If the valid vote rate is below the minValidPerWindow, the validator will be
// slashed and jailed.
func (k Keeper) SlashAndResetMissCounters(ctx sdk.Context) {
	var (
		height               = ctx.BlockHeight()
		distributionHeight   = height - sdk.ValidatorUpdateDelay - 1
		slashWindow          = int64(k.SlashWindow(ctx))
		votePeriod           = int64(k.VotePeriod(ctx))
		votePeriodsPerWindow = sdkmath.LegacyNewDec(slashWindow).QuoInt64(votePeriod).TruncateInt64()
		minValidPerWindow    = k.MinValidPerWindow(ctx)
		slashFraction        = k.SlashFraction(ctx)
		powerReduction       = k.StakingKeeper.PowerReduction(ctx)
	)

	k.IterateMissCounters(ctx, func(operator sdk.ValAddress, missCounter uint64) bool {
		diff := sdkmath.NewInt(votePeriodsPerWindow - int64(missCounter))
		validVoteRate := sdkmath.LegacyNewDecFromInt(diff).QuoInt64(votePeriodsPerWindow)

		// Slash and jail the validator if their valid vote rate is smaller than the
		// minimum threshold.
		if validVoteRate.LT(minValidPerWindow) {
			validator, err := k.StakingKeeper.Validator(ctx, operator)
			if err != nil {
				panic(err)
			}
			if validator.IsBonded() && !validator.IsJailed() {
				consAddr, err := validator.GetConsAddr()
				if err != nil {
					panic(err)
				}

				_, err = k.StakingKeeper.Slash(
					ctx,
					consAddr,
					distributionHeight,
					validator.GetConsensusPower(powerReduction), slashFraction,
				)
				if err != nil {
					panic(err)
				}
				err = k.StakingKeeper.Jail(ctx, consAddr)
				if err != nil {
					panic(err)
				}

				sdkutil.Emit(&ctx, &types.EventSlash{
					Validator: sdk.ConsAddress(consAddr).String(),
					Factor:    slashFraction,
					Reason:    "voting_rate",
					Jailed:    true,
				})
			}
		}

		k.DeleteMissCounter(ctx, operator)
		return false
	})
}
