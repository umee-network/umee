package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v5/x/leverage/types"
)

// afterTokenRegistered notifies any modules which have registered TokenHooks of
// a new token in the leverage module registry
func (k *Keeper) afterTokenRegistered(ctx sdk.Context, t types.Token) {
	for _, h := range k.tokenHooks {
		h.AfterTokenRegistered(ctx, t)
	}
}

// afterRegisteredTokenRemoved notifies any modules which have registered TokenHooks of
// a deleted token in the leverage module registry
func (k *Keeper) afterRegisteredTokenRemoved(ctx sdk.Context, t types.Token) {
	for _, h := range k.tokenHooks {
		h.AfterRegisteredTokenRemoved(ctx, t)
	}
}

// bondedCollateral returns how much of an account's collateral is bonded to external modules.
// this amount of collateral is not allowed to be decollateralized or withdrawn. Note that if multiple
// modules register bond hooks, the amount returned is the maximum (not the sum) of a user's bond
// amounts with each module.
//
// Tokens which are in an unbonding period should be considered as bonded for the purposes of this function,
// as they cannot be withdrawn until unbonding is completed.
//
// e.g. If addr has 45 u/uumee bonded to incentive module and 23 u/uumee bonded to a mystery module,
// bondedCollateral(addr,"u/uumee") returns 45.
func (k *Keeper) bondedCollateral(ctx sdk.Context, addr sdk.AccAddress, uDenom string) sdk.Coin {
	bondedAmount := sdk.ZeroInt()
	for _, h := range k.bondHooks {
		bondedAmount = sdk.MaxInt(bondedAmount, h.GetBonded(ctx, addr, uDenom))
	}
	return sdk.NewCoin(uDenom, bondedAmount)
}

// reduceBondTo instantly unbonds an account's collateral from any modules which has registered bond hooks,
// until bonded (plus unbonding) amount is equal to or less than a given uToken amount. This is used during
// liquidations.
//
// If multiple modules have registered bondHooks, applies this effect to each module independently of the others.
func (k Keeper) reduceBondTo(ctx sdk.Context, addr sdk.AccAddress, collateral sdk.Coin) error {
	for _, h := range k.bondHooks {
		err := h.ForceUnbondTo(ctx, addr, collateral)
		if err != nil {
			return err
		}
	}
	return nil
}
