package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v3/x/ibctransfer"
)

// SetParams sets the x/ibc-rate-limit module's parameters.
func (k Keeper) SetParams(ctx sdk.Context, params ibctransfer.Params) {
	k.paramSpace.SetParamSet(ctx, &params)
}

// GetParams gets the x/ibc-rate-limit module's parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params ibctransfer.Params) {
	k.paramSpace.GetParamSet(ctx, &params)
	return
}
