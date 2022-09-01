package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/umee-network/umee/v3/x/leverage/types"
)

// SetTokenSettings stores a Token into the x/leverage module's KVStore.
func (k Keeper) SetTokenSettings(ctx sdk.Context, token types.Token) error {
	if err := token.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	tokenKey := types.CreateRegisteredTokenKey(token.BaseDenom)

	bz, err := k.cdc.Marshal(&token)
	if err != nil {
		return err
	}

	k.hooks.AfterTokenRegistered(ctx, token)
	store.Set(tokenKey, bz)
	k.tokenRegCache.Add(token.BaseDenom, token)
	ctx.GasMeter().ConsumeGas(gasCacheUpdate, "cache update")
	return nil
}

// GetTokenSettings gets a token from the x/leverage module's KVStore.
func (k Keeper) GetTokenSettings(ctx sdk.Context, denom string) (types.Token, error) {
	ctx.GasMeter().ConsumeGas(gasCacheAccess, "cache access")
	if v, ok := k.tokenRegCache.Get(denom); ok {
		return v.(types.Token), nil
	}
	store := ctx.KVStore(k.storeKey)
	tokenKey := types.CreateRegisteredTokenKey(denom)

	token := types.Token{}
	bz := store.Get(tokenKey)
	if len(bz) == 0 {
		return token, sdkerrors.Wrap(types.ErrNotRegisteredToken, denom)
	}

	err := k.cdc.Unmarshal(bz, &token)
	return token, err
}
