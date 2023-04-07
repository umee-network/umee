package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v4/x/incentive"
)

// addUnbonding creates an unbonding and adds it to the account's current unbondings in the store.
// Assumes the validity of the unbonding has already been checked. Also updates unbonding amounts
// indirectly by calling setUnbondings.
func (k Keeper) addUnbonding(ctx sdk.Context, addr sdk.AccAddress, uToken sdk.Coin) error {
	if err := k.decreaseBond(ctx, addr, uToken); err != nil {
		return err
	}
	unbondingDuration := k.GetParams(ctx).UnbondingDuration
	if unbondingDuration == 0 {
		// For unbonding duration zero, return after decreasing bonded amount
		// without creating an unbonding struct
		return nil
	}
	currentTime := k.GetLastRewardsTime(ctx)
	unbonding := incentive.Unbonding{
		Amount: uToken,
		// Start and end time are stored based on current parameters, and
		// the stored end time does not change even if the module's unbonding
		// duration parameter is changed. The unbonding will still end early
		// if that parameter is reduced though.
		Start: currentTime,
		End:   currentTime + unbondingDuration,
	}
	unbondings := incentive.AccountUnbondings{
		Account:    addr.String(),
		Denom:      uToken.Denom,
		Unbondings: append(k.getUnbondings(ctx, addr, uToken.Denom), unbonding),
	}
	return k.setUnbondings(ctx, unbondings)
}

// finishUnbondings finishes any unbondings on an account which have reached their end time
func (k Keeper) finishUnbondings(ctx sdk.Context, addr sdk.AccAddress) error {
	time := k.GetLastRewardsTime(ctx)
	duration := k.GetParams(ctx).UnbondingDuration
	bondedDenoms := k.getAllBondDenoms(ctx, addr)

	for _, denom := range bondedDenoms {
		storedUnbondings := k.getUnbondings(ctx, addr, denom)
		newUnbondings := incentive.NewAccountUnbondings(
			addr.String(),
			denom,
			[]incentive.Unbonding{},
		)
		for _, u := range storedUnbondings {
			if u.End > time && u.Start+duration > time {
				// If unbonding has passed neither its original end time nor its dynamic end time based on parameters
				// the unbonding is still ongoing, and can be counted normally.
				// This logic allows a reduction in unbonding duration param to speed up existing unbondings.
				newUnbondings.Unbondings = append(newUnbondings.Unbondings, u)
			}
			// Otherwise, this unbonding is completed, and will be cleared.
		}
		// Set new unbondings list, which contains only unbondings which are still ongoing
		if err := k.setUnbondings(ctx, newUnbondings); err != nil {
			return err
		}
	}
	return nil
}

// emergencyUnbond
func (k Keeper) emergencyUnbond(ctx sdk.Context, addr sdk.AccAddress, uToken sdk.Coin) error {
	penaltyAmount := k.GetParams(ctx).EmergencyUnbondFee.MulInt(uToken.Amount).TruncateInt()

	// instant unbonding penalty is donated to the leverage module as uTokens which are immediately
	// burned. leverage reserved amount increases by token equivalent.
	if err := k.leverageKeeper.DonateCollateral(ctx, addr, sdk.NewCoin(uToken.Denom, penaltyAmount)); err != nil {
		return err
	}

	//
	// TODO: reduce unbondings first, then bonded amount, to reach required instant unbond amount
	//

	// bonded, unbonding, unbondings := k.BondSummary(ctx, addr, uToken.Denom)
	// remainingToUnbond := uToken.Amount

	/*
		for _, u := range unbondings {
			// iterate through unbondings in progress
			if u.Amount.Denom == uToken.Denom {

			}
		}
	*/

	return incentive.ErrNotImplemented
}

// ForceSetCollateral is used by leverage module liquidation hooks to immediately unbond collateral
// which is bonded to or unbonding from an account. The uToken it accepts as input is the amount of
// collateral which the liquidated borrower is left with - bonds and unbondings must be removed
// until they do not total to more than this amount.
func (k Keeper) ForceSetCollateral(ctx sdk.Context, addr sdk.AccAddress, newCollateral sdk.Coin) error {
	// first finishes any in-progress unbondings and claims rewards
	if _, err := k.UpdateAccount(ctx, addr); err != nil {
		return err
	}
	// then detects if bonded or unbonding collateral needs to be forcefully reduced
	bonded, _, _ := k.BondSummary(ctx, addr, newCollateral.Denom)
	if bonded.Amount.GT(newCollateral.Amount) {
		//
		// TODO: reduce unbondings first, then bonded amount, until bonded = newCollateral
		//
		return incentive.ErrNotImplemented
	}
	// no reduction was required
	return nil
}
