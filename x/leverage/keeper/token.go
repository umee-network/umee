package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/x/leverage/types"
)

// FromUTokenToTokenDenom returns the associated token denom for the given uToken
// denom. If the uToken denom does not exist, we assume the association is
// invalid and we return an empty string.
func (k Keeper) FromUTokenToTokenDenom(ctx sdk.Context, uTokenDenom string) string {
	store := ctx.KVStore(k.storeKey)
	key := types.CreateUTokenDenomKey(uTokenDenom)

	bz := store.Get(key)
	if len(bz) == 0 {
		return ""
	}

	return string(bz)
}

// FromTokenToUTokenDenom returns the associated uToken denom for the given token
// denom. If the token denom does not exist, we assume the association is invalid
// and we return an empty string.
func (k Keeper) FromTokenToUTokenDenom(ctx sdk.Context, tokenDenom string) string {
	store := ctx.KVStore(k.storeKey)
	key := types.CreateTokenDenomKey(tokenDenom)

	bz := store.Get(key)
	if len(bz) == 0 {
		return ""
	}

	return string(bz)
}

// IsAcceptedToken returns true if a given (non-UToken) token denom is an
// accepted asset type.
func (k Keeper) IsAcceptedToken(ctx sdk.Context, tokenDenom string) bool {
	store := ctx.KVStore(k.storeKey)
	key := types.CreateTokenDenomKey(tokenDenom)

	return store.Has(key)
}

// IsAcceptedUToken returns true if a given uToken denom is an accepted asset
// type.
func (k Keeper) IsAcceptedUToken(ctx sdk.Context, uTokenDenom string) bool {
	store := ctx.KVStore(k.storeKey)
	key := types.CreateUTokenDenomKey(uTokenDenom)

	return store.Has(key)
}

// SetTokenDenom stores the token denom along with the uToken denom association.
func (k Keeper) SetTokenDenom(ctx sdk.Context, tokenDenom string) {
	uTokenDenom := types.UTokenFromTokenDenom(tokenDenom)
	store := ctx.KVStore(k.storeKey)

	tokenKey := types.CreateTokenDenomKey(tokenDenom)
	store.Set(tokenKey, []byte(uTokenDenom))

	uTokenKey := types.CreateUTokenDenomKey(uTokenDenom)
	store.Set(uTokenKey, []byte(tokenDenom))
}

// SetRegisteredToken stores a token into the x/leverage module's KVStore.
func (k Keeper) SetRegisteredToken(ctx sdk.Context, token types.Token) {
	store := ctx.KVStore(k.storeKey)
	tokenKey := types.CreateRegisteredTokenKey(token.BaseDenom)

	bz, err := token.Marshal()
	if err != nil {
		panic(fmt.Sprintf("failed to encode token: %s", err))
	}

	store.Set(tokenKey, bz)
}
