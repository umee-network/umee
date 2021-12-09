package keeper

import (
	"fmt"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/umee-network/umee/x/leverage/types"
)

// FromUTokenToTokenDenom strips the uToken prefix ("u/") from an input denom.
// An empty string is returned if the prefix is not present or if the resulting
// token denom is not an accepted asset type.
func (k Keeper) FromUTokenToTokenDenom(ctx sdk.Context, uTokenDenom string) string {
	if strings.HasPrefix(uTokenDenom, types.UTokenPrefix) {
		tokenDenom := strings.TrimPrefix(uTokenDenom, types.UTokenPrefix)
		if k.IsAcceptedToken(ctx, tokenDenom) {
			return tokenDenom
		}
	}
	return ""
}

// FromTokenToUTokenDenom adds the uToken prefix ("u/") to an input denom.
// An empty string is returned if the input token denom is not an accepted asset type.
func (k Keeper) FromTokenToUTokenDenom(ctx sdk.Context, tokenDenom string) string {
	if k.IsAcceptedToken(ctx, tokenDenom) {
		return types.UTokenPrefix + tokenDenom
	}
	return ""
}

// IsAcceptedToken returns true if a given (non-UToken) token denom is an
// accepted asset type.
func (k Keeper) IsAcceptedToken(ctx sdk.Context, tokenDenom string) bool {
	store := ctx.KVStore(k.storeKey)
	key := types.CreateRegisteredTokenKey(tokenDenom)

	return store.Has(key)
}

// IsAcceptedUToken returns true if a given uToken denom is associated with
// an accepted base asset type.
func (k Keeper) IsAcceptedUToken(ctx sdk.Context, uTokenDenom string) bool {
	tokenDenom := k.FromUTokenToTokenDenom(ctx, uTokenDenom)
	return k.IsAcceptedToken(ctx, tokenDenom)
}

// SetRegisteredToken stores a token into the x/leverage module's KVStore.
func (k Keeper) SetRegisteredToken(ctx sdk.Context, token types.Token) {
	store := ctx.KVStore(k.storeKey)
	tokenKey := types.CreateRegisteredTokenKey(token.BaseDenom)

	bz, err := token.Marshal()
	if err != nil {
		panic(fmt.Sprintf("failed to encode token: %s", err))
	}

	// For tokens not previously registered, sets tokens:uToken to 1.0
	err = k.InitializeExchangeRate(ctx, token.BaseDenom)
	if err != nil {
		panic(err)
	}

	store.Set(tokenKey, bz)
}

// DeleteRegisteredTokens deletes all registered tokens from the x/leverage
// module's KVStore.
func (k Keeper) DeleteRegisteredTokens(ctx sdk.Context) {
	store := ctx.KVStore(k.storeKey)

	iter := sdk.KVStorePrefixIterator(store, types.KeyPrefixRegisteredToken)
	defer iter.Close()

	var keys [][]byte
	for ; iter.Valid(); iter.Next() {
		keys = append(keys, iter.Key())
	}

	for _, k := range keys {
		store.Delete(k)
	}
}

// GetAllRegisteredTokens returns all the registered tokens from the x/leverage
// module's KVStore.
func (k Keeper) GetAllRegisteredTokens(ctx sdk.Context) ([]types.Token, error) {
	store := ctx.KVStore(k.storeKey)

	iter := sdk.KVStorePrefixIterator(store, types.KeyPrefixRegisteredToken)
	defer iter.Close()

	var tokens []types.Token
	for ; iter.Valid(); iter.Next() {
		var t types.Token
		if err := t.Unmarshal(iter.Value()); err != nil {
			return nil, err
		}

		tokens = append(tokens, t)
	}

	return tokens, nil
}

// GetRegisteredToken gets a token from the x/leverage module's KVStore.
func (k Keeper) GetRegisteredToken(ctx sdk.Context, denom string) (types.Token, error) {
	store := ctx.KVStore(k.storeKey)
	tokenKey := types.CreateRegisteredTokenKey(denom)

	token := types.Token{}
	bz := store.Get(tokenKey)
	if len(bz) == 0 {
		return token, sdkerrors.Wrap(types.ErrInvalidAsset, denom)
	}

	err := token.Unmarshal(bz)
	return token, err
}

// GetReserveFactor gets the reserve factor for a given token.
func (k Keeper) GetReserveFactor(ctx sdk.Context, denom string) (sdk.Dec, error) {
	token, err := k.GetRegisteredToken(ctx, denom)
	if err != nil {
		return sdk.ZeroDec(), err
	}

	return token.ReserveFactor, nil
}

// GetInterestBase gets the base interest rate for a given token.
func (k Keeper) GetInterestBase(ctx sdk.Context, denom string) (sdk.Dec, error) {
	token, err := k.GetRegisteredToken(ctx, denom)
	if err != nil {
		return sdk.ZeroDec(), err
	}

	return token.BaseBorrowRate, nil
}

// GetInterestMax gets the maximum interest rate for a given token.
func (k Keeper) GetInterestMax(ctx sdk.Context, denom string) (sdk.Dec, error) {
	token, err := k.GetRegisteredToken(ctx, denom)
	if err != nil {
		return sdk.ZeroDec(), err
	}

	return token.MaxBorrowRate, nil
}

// GetInterestAtKink gets the interest rate at the "kink" in the
// utilization:interest graph for a given token.
func (k Keeper) GetInterestAtKink(ctx sdk.Context, denom string) (sdk.Dec, error) {
	token, err := k.GetRegisteredToken(ctx, denom)
	if err != nil {
		return sdk.ZeroDec(), err
	}

	return token.KinkBorrowRate, nil
}

// GetInterestKinkUtilization gets the utilization at the "kink" in the
// utilization:interest graph for a given token.
func (k Keeper) GetInterestKinkUtilization(ctx sdk.Context, denom string) (sdk.Dec, error) {
	token, err := k.GetRegisteredToken(ctx, denom)
	if err != nil {
		return sdk.ZeroDec(), err
	}

	return token.KinkUtilizationRate, nil
}

// GetCollateralWeight gets collateral weight of a given token.
func (k Keeper) GetCollateralWeight(ctx sdk.Context, denom string) (sdk.Dec, error) {
	token, err := k.GetRegisteredToken(ctx, denom)
	if err != nil {
		return sdk.ZeroDec(), err
	}

	return token.CollateralWeight, nil
}

// GetLiquidationIncentive gets liquidation incentive of a given token.
func (k Keeper) GetLiquidationIncentive(ctx sdk.Context, denom string) (sdk.Dec, error) {
	token, err := k.GetRegisteredToken(ctx, denom)
	if err != nil {
		return sdk.ZeroDec(), err
	}

	return token.LiquidationIncentive, nil
}
