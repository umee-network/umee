package keeper

import (
	"github.com/umee-network/umee/v6/util/store"

	"github.com/umee-network/umee/v6/x/metoken"
)

// SetParams sets the x/metoken module's parameters.
func (k Keeper) SetParams(params metoken.Params) error {
	return store.SetValue(k.store, keyPrefixParams, &params, "params")
}

// GetParams gets the x/metoken module's parameters.
func (k Keeper) GetParams() metoken.Params {
	params := store.GetValue[*metoken.Params](k.store, keyPrefixParams, "params")

	if params == nil {
		return metoken.Params{}
	}

	return *params
}
