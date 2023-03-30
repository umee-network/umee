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
// on that account for each collateral uToken denom.
func (k Keeper) updateAccount(_ sdk.Context, _ sdk.AccAddress) (sdk.Coins, error) {
	// TODO
	return sdk.Coins{}, incentive.ErrNotImplemented
}
