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
	return nil
}

// GetTokenSettings gets a token from the x/leverage module's KVStore.
func (k Keeper) GetTokenSettings(ctx sdk.Context, denom string) (types.Token, error) {
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

// SaveOrUpdateTokenSettingsToRegistry adds new tokens or updates the new tokens settings to registry.
func (k Keeper) SaveOrUpdateTokenSettingsToRegistry(
	ctx sdk.Context, authority string, tokens []types.Token, registeredTokenDenoms map[string]bool, update bool,
) error {
	for _, token := range tokens {
		if err := token.Validate(); err != nil {
			return err
		}
	}

	for _, token := range tokens {
		if update {
			if _, ok := registeredTokenDenoms[token.BaseDenom]; !ok {
				return types.ErrNotRegisteredToken.Wrapf("token %s is not registered", token.BaseDenom)
			}
		} else {
			if _, ok := registeredTokenDenoms[token.BaseDenom]; ok {
				return types.ErrDuplicateToken.Wrapf("token %s is already registered", token.BaseDenom)
			}
		}
	}

	for _, token := range tokens {
		if err := k.SetTokenSettings(ctx, token); err != nil {
			return err
		}
	}

	return nil
}
