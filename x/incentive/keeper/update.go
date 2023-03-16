package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v4/x/incentive"
)

// updateAccount finishes any unbondings associated with an account which have ended and claims any pending rewards.
// It returns the amount of rewards claimed.
//
// REQUIREMENT: This function must be called during any message or hook which creates an unbonding or updates
// bonded amounts. Leverage hooks which decrease borrower collateral must also call this before acting.
// This ensures that between any two consecutive claims by a single account, bonded amounts were constant
// on that account for each bond tier and collateral uToken denom.
func (k Keeper) updateAccount(ctx sdk.Context, addr sdk.AccAddress) (sdk.Coins, error) {
	/*
		current := k.GetUnbondings(ctx, addr)
		remaining := incentive.AccountUnbondings{
			Account:    addr.String(),
			Unbondings: []incentive.Unbonding{},
		}

		// first clears any completed unbondings
		unixTime := k.GetLastRewardsTime(ctx)
		for _, u := range current {
			if u.End > unixTime {
				// unbondings which have yet to end will be kept
				remaining.Unbondings = append(remaining.Unbondings, u)
			} else {
				// completed unbondings disappear and decrease the account's bonded amounts
				tier := incentive.BondTier(u.Tier)
				if err := k.decreaseBond(ctx, addr, tier, u.Amount); err != nil {
					return sdk.NewCoins(), err
				}
				// TODO: also decrease unbonding amounts
			}
		}

		// store the new list of unbondings
		if err := k.SetUnbondings(ctx, remaining); err != nil {
			return sdk.NewCoins(), err
		}
	*/

	// then claims rewards for all bonded (but not currently unbonding) uTokens
	//
	// TODO: need to subtract unbonding amounts, as this currently does not account for that
	//
	rewards := sdk.NewCoins()
	if err := k.iterateAccountBonds(ctx, addr,
		func(ctx sdk.Context, addr sdk.AccAddress, tier incentive.BondTier, uToken sdk.Coin) error {
			reward, err := k.claimReward(ctx, addr, tier, uToken)
			if err != nil {
				return err
			}
			rewards = reward.Add(reward...)
			return nil
		},
	); err != nil {
		return sdk.NewCoins(), err
	}

	return rewards, nil
}
