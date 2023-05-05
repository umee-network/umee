package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v4/util/store"

	"github.com/umee-network/umee/v4/x/metoken"
)

// SetParams sets the x/metoken module's parameters.
func (k Keeper) SetParams(ctx sdk.Context, params metoken.Params) error {
	err := store.SetObject(k.KVStore(ctx), k.cdc, keyPrefixParams, &params, "params")
	return err
}

// GetParams gets the x/metoken module's parameters.
func (k Keeper) GetParams(ctx sdk.Context) metoken.Params {
	params := metoken.Params{}
	ok := store.GetObject(k.KVStore(ctx), k.cdc, keyPrefixParams, &params, "balance")

	if !ok {
		return metoken.Params{}
	}

	return params
}
