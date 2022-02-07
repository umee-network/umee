package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/umee-network/umee/x/leverage/types"
)

// ExchangeToken converts an sdk.Coin containing a base asset to its value as a
// uToken.
func (k Keeper) ExchangeToken(ctx sdk.Context, token sdk.Coin) (sdk.Coin, error) {
	if !token.IsValid() {
		return sdk.Coin{}, sdkerrors.Wrap(types.ErrInvalidAsset, token.String())
	}

	uTokenDenom := k.FromTokenToUTokenDenom(ctx, token.Denom)
	if uTokenDenom == "" {
		return sdk.Coin{}, sdkerrors.Wrap(types.ErrInvalidAsset, token.Denom)
	}

	exchangeRate := k.DeriveExchangeRate(ctx, token.Denom)

	uTokenAmount := token.Amount.ToDec().Quo(exchangeRate).TruncateInt()
	return sdk.NewCoin(uTokenDenom, uTokenAmount), nil
}

// ExchangeUToken converts an sdk.Coin containing a uToken to its value in a base
// token.
func (k Keeper) ExchangeUToken(ctx sdk.Context, uToken sdk.Coin) (sdk.Coin, error) {
	if !uToken.IsValid() {
		return sdk.Coin{}, sdkerrors.Wrap(types.ErrInvalidAsset, uToken.String())
	}

	tokenDenom := k.FromUTokenToTokenDenom(ctx, uToken.Denom)
	if tokenDenom == "" {
		return sdk.Coin{}, sdkerrors.Wrap(types.ErrInvalidAsset, uToken.Denom)
	}

	exchangeRate := k.DeriveExchangeRate(ctx, tokenDenom)

	tokenAmount := uToken.Amount.ToDec().Mul(exchangeRate).TruncateInt()
	return sdk.NewCoin(tokenDenom, tokenAmount), nil
}

// DeriveExchangeRate calculated the token:uToken exchange rate of a base token denom.
func (k Keeper) DeriveExchangeRate(ctx sdk.Context, denom string) sdk.Dec {
	// uToken exchange rate is equal to the token supply (including borrowed
	// tokens yet to be repaid and excluding tokens reserved) divided by total
	// uTokens in circulation.

	// Get relevant quantities
	moduleBalance := k.ModuleBalance(ctx, denom).ToDec()
	reserveAmount := k.GetReserveAmount(ctx, denom).ToDec()
	totalBorrowed := k.getAdjustedTotalBorrowed(ctx, denom).Mul(k.getInterestScalar(ctx, denom))
	uTokenSupply := k.TotalUTokenSupply(ctx, k.FromTokenToUTokenDenom(ctx, denom)).Amount

	// Derive effective token supply
	tokenSupply := moduleBalance.Add(totalBorrowed).Sub(reserveAmount)

	// Handle uToken supply == 0 case
	if !uTokenSupply.IsPositive() {
		return sdk.OneDec()
	}

	// Derive exchange rate
	return tokenSupply.QuoInt(uTokenSupply)
}
