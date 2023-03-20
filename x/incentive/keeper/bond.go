package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v4/util/coin"
	"github.com/umee-network/umee/v4/x/incentive"
)

// getAllBonded gets the total amount of a uToken bonded to an account across all three unbonding tiers
func (k Keeper) getAllBonded(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	bonded := coin.Zero(denom)
	for _, tier := range []incentive.BondTier{incentive.BondTierLong, incentive.BondTierMiddle, incentive.BondTierShort} {
		bonded = bonded.Add(k.getBonded(ctx, addr, denom, tier))
	}
	return bonded
}

// accountBonds gets the total bonded and unbonding for a given account, as well as a list of ongoing unbondings,
// for a single bond tier and uToken denom. Using incentive.BondTierUnspecified returns all tiers instead.
// It reduces the returned bonded amount by completed unbondings without actually clearing those unbondings in state,
// so it is safe for queries and any parts of Msg functions which are not intended to alter state.
func (k Keeper) accountBonds(ctx sdk.Context, addr sdk.AccAddress, denom string, tier incentive.BondTier) (
	bonded sdk.Coin, unbonding sdk.Coin, unbondings []incentive.Unbonding,
) {
	bonded = coin.Zero(denom)
	unbonding = coin.Zero(denom)
	unbondings = []incentive.Unbonding{}

	time := k.getLastRewardsTime(ctx)

	// sum all bonded and unbonding tokens for this denom, for the specified tier or across all tiers
	for _, t := range []incentive.BondTier{incentive.BondTierLong, incentive.BondTierMiddle, incentive.BondTierShort} {
		if tier == incentive.BondTierUnspecified || t == tier {
			bonded = bonded.Add(k.getBonded(ctx, addr, denom, t))

			tierUnbondings := k.getUnbondings(ctx, addr, denom, t)
			for _, u := range tierUnbondings {
				if u.End > time {
					// this unbonding is still ongoing, and can be counted normally
					unbondings = append(unbondings, u)
					unbonding = unbonding.Add(u.Amount)
				} else {
					// this unbonding is completed, and will be quietly cleared by the next state-altering message
					// which is permitted to affect this account. For now, it is omitted from the returned unbonding
					// total and reduces the bonded total.
					bonded = bonded.Sub(u.Amount)
				}
			}
		}
	}

	// returned values reflect the real situation (after completed unbondings), not what is currently stored in state
	return bonded, unbonding, unbondings
}

// increaseBond increases the bonded uToken amount for an account at a given tier, and updates the module's total.
// it also initializes the account's reward tracker if the bonded amount was zero. This function should only be
// called during MsgBond.
func (k Keeper) increaseBond(ctx sdk.Context, addr sdk.AccAddress, tier incentive.BondTier, bond sdk.Coin) error {
	// get bonded amount before adding the currently bonding tokens
	bonded := k.getBonded(ctx, addr, bond.Denom, tier)
	// if bonded amount was zero, reward tracker must be initialized
	if bonded.IsZero() {
		if err := k.UpdateRewardTracker(ctx, addr, tier, bond.Denom); err != nil {
			return err
		}
	}
	// update bonded amount
	if err := k.setBonded(ctx, addr, bonded.Add(bond), tier); err != nil {
		return err
	}
	// update module total
	total := k.getTotalBonded(ctx, bond.Denom, tier)
	return k.setTotalBonded(ctx, total.Add(bond), tier)
}

// DecreaseBond decreases the bonded uToken amount for an account at a given tier, and updates the module's total.
// it also clears reward trackers for any bond tiers which have reached zero for the account, to save storage.
// This function must be called when unbondings are completed (not created) and when the leverage module forcefully
// liquidates bonded collateral.
func (k Keeper) DecreaseBond(ctx sdk.Context, addr sdk.AccAddress, tier incentive.BondTier, unbond sdk.Coin) error {
	// calculate and set new bonded amount after subtracting the currently unbonding tokens
	bonded := k.getBonded(ctx, addr, unbond.Denom, tier).Sub(unbond)
	if err := k.setBonded(ctx, addr, bonded, tier); err != nil {
		return err
	}
	// if the new bond for this account + tier + denom has reached zero, it is safe to stop tracking rewards
	if bonded.IsZero() {
		err := k.ClearRewardTracker(ctx, addr, tier, unbond.Denom)
		if err != nil {
			return err
		}
	}
	// update module total
	total := k.getTotalBonded(ctx, unbond.Denom, tier)
	return k.setTotalBonded(ctx, total.Sub(unbond), tier)
}