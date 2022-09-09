package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/umee-network/umee/v3/x/leverage/types"
)

// ExchangeToken converts an sdk.Coin containing a base asset to its value as a
// uToken.
func (k Keeper) ExchangeToken(ctx sdk.Context, token sdk.Coin) (sdk.Coin, error) {
	if err := token.Validate(); err != nil {
		return sdk.Coin{}, err
	}

	uTokenDenom := types.ToUTokenDenom(token.Denom)
	if uTokenDenom == "" {
		return sdk.Coin{}, sdkerrors.Wrap(types.ErrUToken, token.Denom)
	}

	exchangeRate := k.DeriveExchangeRate(ctx, token.Denom)

	uTokenAmount := toDec(token.Amount).Quo(exchangeRate).TruncateInt()
	return sdk.NewCoin(uTokenDenom, uTokenAmount), nil
}

// ExchangeUToken converts an sdk.Coin containing a uToken to its value in a base
// token.
func (k Keeper) ExchangeUToken(ctx sdk.Context, uToken sdk.Coin) (sdk.Coin, error) {
	if err := uToken.Validate(); err != nil {
		return sdk.Coin{}, err
	}

	tokenDenom := types.ToTokenDenom(uToken.Denom)
	if tokenDenom == "" {
		return sdk.Coin{}, sdkerrors.Wrap(types.ErrNotUToken, uToken.Denom)
	}

	exchangeRate := k.DeriveExchangeRate(ctx, tokenDenom)

	tokenAmount := toDec(uToken.Amount).Mul(exchangeRate).TruncateInt()
	return sdk.NewCoin(tokenDenom, tokenAmount), nil
}

// ExchangeUTokens converts an sdk.Coins containing uTokens to their values in base
// tokens.
func (k Keeper) ExchangeUTokens(ctx sdk.Context, uTokens sdk.Coins) (sdk.Coins, error) {
	if err := uTokens.Validate(); err != nil {
		return sdk.Coins{}, err
	}

	tokens := sdk.Coins{}
	for _, coin := range uTokens {
		token, err := k.ExchangeUToken(ctx, coin)
		if err != nil {
			return sdk.Coins{}, err
		}
		tokens = tokens.Add(token)
	}

	return tokens, nil
}

// DeriveExchangeRate calculated the token:uToken exchange rate of a base token denom.
func (k Keeper) DeriveExchangeRate(ctx sdk.Context, denom string) sdk.Dec {
	// uToken exchange rate is equal to the token supply (including borrowed
	// tokens yet to be repaid and excluding tokens reserved) divided by total
	// uTokens in circulation.

	// Get relevant quantities
	moduleBalance := toDec(k.ModuleBalance(ctx, denom).Amount)
	reserveAmount := toDec(k.GetReserves(ctx, denom).Amount)
	totalBorrowed := k.getAdjustedTotalBorrowed(ctx, denom).Mul(k.getInterestScalar(ctx, denom))
	uTokenSupply := k.GetUTokenSupply(ctx, types.ToUTokenDenom(denom)).Amount

	// Derive effective token supply
	tokenSupply := moduleBalance.Add(totalBorrowed).Sub(reserveAmount)

	// Handle uToken supply == 0 case
	if !uTokenSupply.IsPositive() {
		return sdk.OneDec()
	}

	// Derive exchange rate
	return tokenSupply.QuoInt(uTokenSupply)
}
