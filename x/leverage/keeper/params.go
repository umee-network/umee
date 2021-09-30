package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/x/leverage/types"
)

// SetParams sets the x/leverage module's parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) {
	k.paramStore.SetParamSet(ctx, &params)
}

// GetParams gets the x/leverage module's parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params types.Params) {
	k.paramStore.GetParamSet(ctx, &params)
	return
}
