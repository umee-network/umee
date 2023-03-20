package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v4/util/coin"
	"github.com/umee-network/umee/v4/x/incentive"
)

var bondTiers = []incentive.BondTier{incentive.BondTierLong, incentive.BondTierMiddle, incentive.BondTierShort}

// bondTier converts from the uint32 used in message types to the enumeration, returning an error
// if it is not valid. Does not allow incentive.BondTierUnspecified
func bondTier(n uint32) (incentive.BondTier, error) {
	if n == 0 || n > uint32(incentive.BondTierLong) {
		return incentive.BondTierUnspecified, incentive.ErrInvalidTier.Wrapf("%d", n)
	}
	return incentive.BondTier(n), nil
}

// getAllBonded gets the total amount of a uToken bonded to an account across all three unbonding tiers.
// This does not include tokens currently unbonding.
func (k Keeper) getAllBonded(ctx sdk.Context, addr sdk.AccAddress, denom string) sdk.Coin {
	bonded := coin.Zero(denom)
	for _, tier := range bondTiers {
		bonded = bonded.Add(k.getBonded(ctx, addr, denom, tier))
	}
	return bonded
}

// accountBonds gets the total bonded and unbonding for a given account, as well as a list of ongoing unbondings,
// for a single bond tier and uToken denom. Using incentive.BondTierUnspecified returns all tiers instead.
// It ignores completed unbondings without actually clearing those unbondings in state, so it is safe for
// queries and any parts of Msg functions which are not intended to alter state.
func (k Keeper) accountBonds(ctx sdk.Context, addr sdk.AccAddress, denom string, tier incentive.BondTier) (
	bonded sdk.Coin, unbonding sdk.Coin, unbondings []incentive.Unbonding,
) {
	bonded = coin.Zero(denom)
	unbonding = coin.Zero(denom)
	unbondings = []incentive.Unbonding{}

	time := k.getLastRewardsTime(ctx)

	// sum all bonded and unbonding tokens for this denom, for the specified tier or across all tiers
	for _, t := range bondTiers {
		if tier == incentive.BondTierUnspecified || t == tier {
			bonded = bonded.Add(k.getBonded(ctx, addr, denom, t))

			tierUnbondings := k.getUnbondings(ctx, addr, denom, t)
			for _, u := range tierUnbondings {
				if u.End > time {
					// this unbonding is still ongoing, and can be counted normally
					unbondings = append(unbondings, u)
					unbonding = unbonding.Add(u.Amount)
				}
				// Otherwise, this unbonding is completed, and will be quietly cleared by the next state-altering message
				// which is permitted to affect this account. For now, it is omitted from the returned unbonding total.
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
	// update bonded amount (also increases TotalBonded)
	return k.setBonded(ctx, addr, bonded.Add(bond), tier)
}

// decreaseBond decreases the bonded uToken amount for an account at a given tier, and updates the module's total.
// it also clears reward trackers for any bond tiers which have reached zero for the account, to save storage.
// This function must be called during MsgBeginUnbonding and when the leverage module forcefully liquidates bonded
// (but not already unbonding) collateral.
func (k Keeper) decreaseBond(ctx sdk.Context, addr sdk.AccAddress, tier incentive.BondTier, unbond sdk.Coin) error {
	// calculate new bonded amount after subtracting the currently unbonding tokens
	bonded := k.getBonded(ctx, addr, unbond.Denom, tier).Sub(unbond)

	// if the new bond for this account + tier + denom has reached zero, it is safe to stop tracking rewards
	if bonded.IsZero() {
		err := k.clearRewardTracker(ctx, addr, tier, unbond.Denom)
		if err != nil {
			return err
		}
	}

	// update bonded amount (also decreases TotalBonded)
	return k.setBonded(ctx, addr, bonded, tier)
}
