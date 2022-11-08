package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v3/x/incentive/types"
)

// SetParams sets the x/incentive module's parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	// k.paramSpace.SetParamSet(ctx, &params)
	return types.ErrNotImplemented
}

// GetParams gets the x/incentive module's parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	// k.paramSpace.GetParamSet(ctx, &params)
	return types.Params{}
}
