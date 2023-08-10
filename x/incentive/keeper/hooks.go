package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	leveragetypes "github.com/umee-network/umee/v6/x/leverage/types"
)

// BondHooks defines a structure around the x/incentive Keeper that implements various
// BondHooks interface defined by other modules such as x/leverage.
type BondHooks struct {
	k Keeper
}

var _ leveragetypes.BondHooks = BondHooks{}

// BondHooks returns a new Hooks instance that wraps the x/incentive keeper.
func (k Keeper) BondHooks() BondHooks {
	return BondHooks{k}
}

// GetBonded gets sum of bonded and unbonding uTokens of a given denom for an account.
func (h BondHooks) GetBonded(ctx sdk.Context, addr sdk.AccAddress, uDenom string) sdkmath.Int {
	return h.k.restrictedCollateral(ctx, addr, uDenom).Amount
}

// ForceUnbondTo instantly unbonds uTokens until an account's bonded amount of a given uToken
// is no greater than a certain amount.
func (h BondHooks) ForceUnbondTo(ctx sdk.Context, addr sdk.AccAddress, uToken sdk.Coin) error {
	if err := uToken.Validate(); err != nil {
		return err
	}
	// ensure rewards and unbondings are up to date when using liquidation hooks
	if _, err := h.k.UpdateAccount(ctx, addr); err != nil {
		return err
	}
	return h.k.reduceBondTo(ctx, addr, uToken)
}
