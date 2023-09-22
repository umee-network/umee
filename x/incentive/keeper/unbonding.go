package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/umee-network/umee/v6/x/incentive"
)

// UnbondForModule attempts to unbond tokens bonded from a module account. In addition to the regular error
// return, also returns a boolean which indicates whether the error was recoverable.
// A recoverable = true error means UnbondForModule was aborted without harming state.
func (k Keeper) UnbondForModule(ctx sdk.Context, forModule string, coin sdk.Coin) (bool, error) {
	moduleAddr := authtypes.NewModuleAddress(forModule)
	// get current account state for the requested uToken denom only
	bonded, _, unbondings := k.BondSummary(ctx, moduleAddr, coin.Denom)

	maxUnbondings := int(k.GetParams(ctx).MaxUnbondings)
	if maxUnbondings > 0 && len(unbondings) >= maxUnbondings {
		// reject concurrent unbondings that would exceed max unbondings - zero is unlimited
		return true, incentive.ErrMaxUnbondings.Wrapf("%d", len(unbondings))
	}

	// reject unbondings greater than maximum available amount
	if coin.Amount.GT(bonded.Amount) {
		return true, incentive.ErrInsufficientBonded.Wrapf(
			"bonded: %s, requested: %s",
			bonded,
			coin,
		)
	}

	// clear completed unbondings and claim all rewards
	// this must happen before unbonding is created, as rewards are for the previously bonded amount
	_, err := k.UpdateAccount(ctx, moduleAddr)
	if err != nil {
		return false, err
	}

	// start the unbonding
	return false, k.addUnbonding(ctx, moduleAddr, coin)
}

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
		UToken: uToken,
		// Start and end time are stored based on current parameters, and
		// the stored end time does not change even if the module's unbonding
		// duration parameter is changed. The unbonding will still end early
		// if that parameter is reduced though.
		Start: currentTime,
		End:   currentTime + unbondingDuration,
	}
	unbondings := incentive.AccountUnbondings{
		Account:    addr.String(),
		UToken:     uToken.Denom,
		Unbondings: append(k.getUnbondings(ctx, addr, uToken.Denom), unbonding),
	}
	return k.setUnbondings(ctx, unbondings)
}

// cleanupUnbondings finishes any unbondings on an account which have reached their end time
func (k Keeper) cleanupUnbondings(ctx sdk.Context, addr sdk.AccAddress) error {
	time := k.GetLastRewardsTime(ctx)
	duration := k.GetParams(ctx).UnbondingDuration
	bondedDenoms, err := k.getAllBondDenoms(ctx, addr)
	if err != nil {
		return err
	}

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

// reduceBondTo is used by MsgEmergencyUnbond and by liquidation hooks to immediately unbond collateral
// which is bonded to or unbonding from an account. The uToken it accepts as input is the amount of
// collateral which the borrower is left with: bonds and unbondings must be removed until they do not
// total to more than this amount.
func (k Keeper) reduceBondTo(ctx sdk.Context, addr sdk.AccAddress, newCollateral sdk.Coin) error {
	// detect if bonded or unbonding collateral needs to be forcefully reduced
	bonded, unbonding, unbondings := k.BondSummary(ctx, addr, newCollateral.Denom)
	if bonded.Amount.Add(unbonding.Amount).LTE(newCollateral.Amount) {
		// nothing needs to happen
		return nil
	}
	if bonded.Amount.GTE(newCollateral.Amount) {
		// if remaining collateral is less than or equal to bonded amount,
		// all in-progress unbondings and potentially some bonded tokens
		// must be instantly unbonded.
		if err := k.setBonded(ctx, addr, newCollateral); err != nil {
			return err
		}
		// set new (empty) list of unbondings, which clears it from store
		au := incentive.AccountUnbondings{
			Account:    addr.String(),
			UToken:     newCollateral.Denom,
			Unbondings: []incentive.Unbonding{},
		}
		return k.setUnbondings(ctx, au)
	}
	// if we have not returned yet, the only some in-progress unbondings will be
	// instantly unbonded.
	amountToUnbond := bonded.Amount.Add(unbonding.Amount).Sub(newCollateral.Amount)
	for i, u := range unbondings {
		// for ongoing unbondings, starting with the oldest
		specificReduction := sdk.MinInt(amountToUnbond, u.UToken.Amount)
		// reduce the in-progress unbonding amount, and the remaining instant unbond
		unbondings[i].UToken.Amount = u.UToken.Amount.Sub(specificReduction)
		amountToUnbond = amountToUnbond.Sub(specificReduction)
		// if no more unbondings need to be reduced, break out of the loop early
		if amountToUnbond.IsZero() {
			break
		}
	}
	// set new (reduced but not empty) list of unbondings
	au := incentive.AccountUnbondings{
		Account:    addr.String(),
		UToken:     newCollateral.Denom,
		Unbondings: unbondings,
	}
	return k.setUnbondings(ctx, au)
}
