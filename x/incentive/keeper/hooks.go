package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	leveragetypes "github.com/umee-network/umee/v4/x/leverage/types"
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
	bonded, unbonding, _ := h.k.BondSummary(ctx, addr, uDenom)
	return bonded.Amount.Add(unbonding.Amount)
}

// ForceUnondTo instantly unbonds uTokens until an account's bonded amount of a given uToken
// is no greater than a certain amount.
func (h BondHooks) ForceUnondTo(ctx sdk.Context, addr sdk.AccAddress, uToken sdk.Coin) error {
	return h.k.ForceUnondTo(ctx, addr, uToken)
}
