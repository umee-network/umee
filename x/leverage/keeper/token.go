package keeper

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/umee-network/umee/v2/x/leverage/types"
)

// FromUTokenToTokenDenom strips the uToken prefix ("u/") from an input denom.
// An empty string is returned if the prefix is not present.
func (k Keeper) FromUTokenToTokenDenom(ctx sdk.Context, uTokenDenom string) string {
	if strings.HasPrefix(uTokenDenom, types.UTokenPrefix) {
		return strings.TrimPrefix(uTokenDenom, types.UTokenPrefix)
	}
	return ""
}

// FromTokenToUTokenDenom adds the uToken prefix ("u/") to an input denom.
// An empty string is returned if the input token denom already has the prefix.
func (k Keeper) FromTokenToUTokenDenom(ctx sdk.Context, tokenDenom string) string {
	if strings.HasPrefix(tokenDenom, types.UTokenPrefix) {
		return ""
	}
	return types.UTokenPrefix + tokenDenom
}

// IsAcceptedToken returns true if a given (non-UToken) token denom is an
// existing, non-blacklisted asset type.
func (k Keeper) IsAcceptedToken(ctx sdk.Context, tokenDenom string) bool {
	t, err := k.GetTokenSettings(ctx, tokenDenom)
	return err == nil && !t.Blacklist
}

// IsAcceptedUToken returns true if a given uToken denom is associated with
// an accepted base asset type.
func (k Keeper) IsAcceptedUToken(ctx sdk.Context, uTokenDenom string) bool {
	tokenDenom := k.FromUTokenToTokenDenom(ctx, uTokenDenom)
	return k.IsAcceptedToken(ctx, tokenDenom)
}

// SetTokenSettings stores a Token into the x/leverage module's KVStore.
func (k Keeper) SetTokenSettings(ctx sdk.Context, token types.Token) error {
	if err := token.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	tokenKey := types.CreateRegisteredTokenKey(token.BaseDenom)

	bz, err := k.cdc.Marshal(&token)
	if err != nil {
		panic(fmt.Errorf("failed to encode token settings: %w", err))
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
		return token, sdkerrors.Wrap(types.ErrInvalidAsset, denom)
	}

	err := k.cdc.Unmarshal(bz, &token)
	return token, err
}
