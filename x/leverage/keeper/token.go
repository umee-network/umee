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
	return nil
}

// GetTokenSettings gets a token from the x/leverage module's KVStore.
func (k Keeper) GetTokenSettings(ctx sdk.Context, denom string) (types.Token, error) {
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

// GetReserveFactor gets the reserve factor for a given token.
func (k Keeper) GetReserveFactor(ctx sdk.Context, denom string) (sdk.Dec, error) {
	token, err := k.GetTokenSettings(ctx, denom)
	if err != nil {
		return sdk.ZeroDec(), err
	}

	return token.ReserveFactor, nil
}

// GetInterestBase gets the base interest rate for a given token.
func (k Keeper) GetInterestBase(ctx sdk.Context, denom string) (sdk.Dec, error) {
	token, err := k.GetTokenSettings(ctx, denom)
	if err != nil {
		return sdk.ZeroDec(), err
	}

	return token.BaseBorrowRate, nil
}

// GetInterestAtKink gets the interest rate at the "kink" in the
// utilization:interest graph for a given token.
func (k Keeper) GetInterestAtKink(ctx sdk.Context, denom string) (sdk.Dec, error) {
	token, err := k.GetTokenSettings(ctx, denom)
	if err != nil {
		return sdk.ZeroDec(), err
	}

	return token.KinkBorrowRate, nil
}

// GetCollateralWeight gets collateral weight of a given token.
func (k Keeper) GetCollateralWeight(ctx sdk.Context, denom string) (sdk.Dec, error) {
	token, err := k.GetTokenSettings(ctx, denom)
	if err != nil {
		return sdk.ZeroDec(), err
	}

	return token.CollateralWeight, nil
}

// GetLiquidationThreshold gets liquidation threshold of a given token.
func (k Keeper) GetLiquidationThreshold(ctx sdk.Context, denom string) (sdk.Dec, error) {
	token, err := k.GetTokenSettings(ctx, denom)
	if err != nil {
		return sdk.ZeroDec(), err
	}

	return token.LiquidationThreshold, nil
}

// GetLiquidationIncentive gets liquidation incentive of a given token.
func (k Keeper) GetLiquidationIncentive(ctx sdk.Context, denom string) (sdk.Dec, error) {
	token, err := k.GetTokenSettings(ctx, denom)
	if err != nil {
		return sdk.ZeroDec(), err
	}

	return token.LiquidationIncentive, nil
}

// AssertLendEnabled returns an error if a token does not exist or cannot be lent.
func (k Keeper) AssertLendEnabled(ctx sdk.Context, denom string) error {
	token, err := k.GetTokenSettings(ctx, denom)
	if err != nil {
		return err
	}
	if !token.EnableMsgLend {
		return sdkerrors.Wrap(types.ErrLendNotAllowed, denom)
	}

	return nil
}

// AssertBorrowEnabled returns an error if a token does not exist or cannot be borrowed.
func (k Keeper) AssertBorrowEnabled(ctx sdk.Context, denom string) error {
	token, err := k.GetTokenSettings(ctx, denom)
	if err != nil {
		return err
	}
	if !token.EnableMsgBorrow {
		return sdkerrors.Wrap(types.ErrBorrowNotAllowed, denom)
	}

	return nil
}

// AssertNotBlacklisted returns an error if a token does not exist or is blacklisted.
func (k Keeper) AssertNotBlacklisted(ctx sdk.Context, denom string) error {
	token, err := k.GetTokenSettings(ctx, denom)
	if err != nil {
		return err
	}
	if token.Blacklist {
		return sdkerrors.Wrap(types.ErrBlacklisted, denom)
	}

	return nil
}
