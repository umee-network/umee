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
	if k.getUnbondingDuration(ctx) == 0 {
		// For unbonding duration zero, return after decreasing bonded amount
		// without creating an unbonding struct
		return nil
	}
	unbonding := incentive.Unbonding{
		Amount: uToken,
		End:    k.getLastRewardsTime(ctx) + k.getUnbondingDuration(ctx),
	}
	unbondings := incentive.AccountUnbondings{
		Account:    addr.String(),
		Denom:      uToken.Denom,
		Unbondings: append(k.getUnbondings(ctx, addr, uToken.Denom), unbonding),
	}
	return k.setUnbondings(ctx, unbondings)
}
