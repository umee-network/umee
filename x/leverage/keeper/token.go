package keeper

import (
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v5/util/store"
	"github.com/umee-network/umee/v5/x/leverage/types"
)

// CleanTokenRegistry deletes all blacklisted tokens in the leverage registry
// whose uToken supplies are zero. Called automatically on registry update.
func (k Keeper) CleanTokenRegistry(ctx sdk.Context) error {
	tokens := k.GetAllRegisteredTokens(ctx)
	for _, t := range tokens {
		if t.Blacklist {
			uDenom := types.ToUTokenDenom(t.BaseDenom)
			uSupply := k.GetUTokenSupply(ctx, uDenom)
			if uSupply.IsZero() {
				err := k.deleteTokenSettings(ctx, t)
				if err != nil {
					return err
				}
			}
		}
	}
	return nil
}

// deleteTokenSettings deletes a Token in the x/leverage module's KVStore.
// it should only be called by CleanTokenRegistry.
func (k Keeper) deleteTokenSettings(ctx sdk.Context, token types.Token) error {
	store := ctx.KVStore(k.storeKey)
	tokenKey := types.KeyRegisteredToken(token.BaseDenom)
	store.Delete(tokenKey)
	// call token hooks on deleted (not just blacklisted) token
	k.afterRegisteredTokenRemoved(ctx, token)
	return nil
}

// SetTokenSettings stores a Token into the x/leverage module's KVStore.
func (k Keeper) SetTokenSettings(ctx sdk.Context, token types.Token) error {
	if err := token.Validate(); err != nil {
		return err
	}

	store := ctx.KVStore(k.storeKey)
	tokenKey := types.KeyRegisteredToken(token.BaseDenom)

	bz, err := k.cdc.Marshal(&token)
	if err != nil {
		return err
	}

	k.afterTokenRegistered(ctx, token)
	store.Set(tokenKey, bz)
	return nil
}

// GetTokenSettings gets a token from the x/leverage module's KVStore.
func (k Keeper) GetTokenSettings(ctx sdk.Context, denom string) (types.Token, error) {
	store := ctx.KVStore(k.storeKey)
	tokenKey := types.KeyRegisteredToken(denom)

	token := types.Token{}
	bz := store.Get(tokenKey)
	if len(bz) == 0 {
		return token, types.ErrNotRegisteredToken.Wrap(denom)
	}

	err := k.cdc.Unmarshal(bz, &token)
	return token, err
}

// SetSpecialAssetPair stores a SpecialAssetPair into the x/leverage module's KVStore.
// Deletes any existing special pairs between the assets instead if given zero
// collateral weight.
func (k Keeper) SetSpecialAssetPair(ctx sdk.Context, pair types.SpecialAssetPair) error {
	if err := pair.Validate(); err != nil {
		return err
	}
	if !pair.CollateralWeight.IsPositive() {
		k.deleteSpecialAssetPair(ctx, pair.Collateral, pair.Borrow)
		return nil
	}

	key := types.KeySpecialAssetPair(pair.Collateral, pair.Borrow)
	return store.SetValue(ctx.KVStore(k.storeKey), key, &pair, "leverage-special-asset")
}

// deleteSpecialAssetPair removes a SpecialAssetPair from the x/leverage module's KVStore.
func (k Keeper) deleteSpecialAssetPair(ctx sdk.Context, collateralDenom, borrowDenom string) {
	key := types.KeySpecialAssetPair(collateralDenom, borrowDenom)
	ctx.KVStore(k.storeKey).Delete(key)
}

// SaveOrUpdateTokenSettingsToRegistry adds new tokens or updates the new tokens settings to registry.
// It requires maps of the currently registered base and symbol denoms, so it can prevent duplicates of either.
func (k Keeper) SaveOrUpdateTokenSettingsToRegistry(
	ctx sdk.Context, tokens []types.Token, regdTkDenoms, regdSymDenoms map[string]bool, update bool,
) error {
	for _, token := range tokens {
		if err := token.Validate(); err != nil {
			return err
		}
	}

	for _, token := range tokens {
		if update {
			if _, ok := regdTkDenoms[token.BaseDenom]; !ok {
				return types.ErrNotRegisteredToken.Wrapf("token %s is not registered", token.BaseDenom)
			}
		} else {
			if _, ok := regdTkDenoms[token.BaseDenom]; ok {
				return types.ErrDuplicateToken.Wrapf("token %s is already registered", token.BaseDenom)
			}

			if _, ok := regdSymDenoms[strings.ToUpper(token.SymbolDenom)]; ok {
				return types.ErrDuplicateToken.Wrapf("symbol denom %s is already registered", token.SymbolDenom)
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
