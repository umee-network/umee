package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v6/util/store"
	"github.com/umee-network/umee/v6/x/leverage/types"
)

// SetParams sets the x/leverage module's parameters.
func (k Keeper) SetParams(ctx sdk.Context, params types.Params) error {
	return store.SetValue(ctx.KVStore(k.storeKey), types.KeyParams, &params, "leverage params")
}

// GetParams gets the x/leverage module's parameters.
func (k Keeper) GetParams(ctx sdk.Context) types.Params {
	params := store.GetValue[*types.Params](ctx.KVStore(k.storeKey), types.KeyParams, "leverage params")
	if params == nil {
		panic("params not initialized")
	}
	return *params
}
