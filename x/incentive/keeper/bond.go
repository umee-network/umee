package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v4/util/coin"
	"github.com/umee-network/umee/v4/x/incentive"
)

// BondSummary gets the total bonded and unbonding for a given account, as well as a list of ongoing unbondings,
// for a single uToken denom. It ignores completed unbondings without actually clearing those unbondings in state,
// so it is safe for use by queries and any parts of Msg functions which are not intended to alter state.
func (k Keeper) BondSummary(ctx sdk.Context, addr sdk.AccAddress, denom string) (
	bonded sdk.Coin, unbonding sdk.Coin, unbondings []incentive.Unbonding,
) {
	bonded = coin.Zero(denom)
	unbonding = coin.Zero(denom)
	unbondings = []incentive.Unbonding{}

	time := k.getLastRewardsTime(ctx)
	duration := k.getUnbondingDuration(ctx)

	// sum all bonded and unbonding tokens for this denom
	bonded = bonded.Add(k.getBonded(ctx, addr, denom))

	storedUnbondings := k.getUnbondings(ctx, addr, denom)
	for _, u := range storedUnbondings {
		if u.End > time && u.Start+duration > time {
			// If unbonding has passed neither its original end time nor its dynamic end time based on parameters
			// the unbonding is still ongoing, and can be counted normally.
			// This logic allows a reduction in unbonding duration param to speed up existing unbondings.
			unbondings = append(unbondings, u)
			unbonding = unbonding.Add(u.Amount)
		}
		// Otherwise, this unbonding is completed, and will be quietly cleared by the next state-altering message
		// which is permitted to affect this account. For now, it is omitted from the returned unbonding total.
	}

	// returned values reflect the real situation (after completed unbondings), not what is currently stored in state
	return bonded, unbonding, unbondings
}

// increaseBond increases the bonded uToken amount for an account, and updates the module's total.
// it also initializes the account's reward tracker if the bonded amount was zero. This function should only be
// called during MsgBond.
func (k Keeper) increaseBond(ctx sdk.Context, addr sdk.AccAddress, bond sdk.Coin) error {
	// get bonded amount before adding the currently bonding tokens
	bonded := k.getBonded(ctx, addr, bond.Denom)
	// if bonded amount was zero, reward tracker must be initialized
	if bonded.IsZero() {
		if err := k.UpdateRewardTracker(ctx, addr, bond.Denom); err != nil {
			return err
		}
	}
	// update bonded amount (also increases TotalBonded)
	return k.setBonded(ctx, addr, bonded.Add(bond))
}

// decreaseBond decreases the bonded uToken amount for an account, and updates the module's total.
// it also clears reward trackers if bond amount reached zero for the account, to save storage.
// This function must be called during MsgBeginUnbonding and when the leverage module forcefully liquidates bonded
// (but not already unbonding) collateral.
func (k Keeper) decreaseBond(ctx sdk.Context, addr sdk.AccAddress, unbond sdk.Coin) error {
	// calculate new bonded amount after subtracting the currently unbonding tokens
	bonded := k.getBonded(ctx, addr, unbond.Denom).Sub(unbond)

	// if the new bond for this account + denom has reached zero, it is safe to stop tracking rewards
	if bonded.IsZero() {
		err := k.clearRewardTracker(ctx, addr, unbond.Denom)
		if err != nil {
			return err
		}
	}

	// update bonded amount (also decreases TotalBonded)
	return k.setBonded(ctx, addr, bonded)
}
