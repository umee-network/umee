package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v4/x/uibc"
)

// SetParams sets the x/uibc module's parameters.
func (k Keeper) SetParams(ctx sdk.Context, params uibc.Params) error {
	store := ctx.KVStore(k.storeKey)
	bz, err := k.cdc.Marshal(&params)
	if err != nil {
		return err
	}
	store.Set(uibc.KeyPrefixParams, bz)

	return nil
}

// GetParams gets the x/uibc module's parameters.
func (k Keeper) GetParams(ctx sdk.Context) (params uibc.Params) {
	store := ctx.KVStore(k.storeKey)
	bz := store.Get(uibc.KeyPrefixParams)
	if bz == nil {
		return params
	}

	k.cdc.MustUnmarshal(bz, &params)
	return params
}
