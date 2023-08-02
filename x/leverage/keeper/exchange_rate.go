package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v5/x/leverage/types"
)

// Token2uTokenRate returns a uToken amount user would receive when supplying the token.
// Returns error if the token is not a Token.
func (k Keeper) Token2uTokenRate(ctx sdk.Context, token sdk.Coin) (sdk.Coin, error) {
	if err := token.Validate(); err != nil {
		return sdk.Coin{}, err
	}

	uTokenDenom := types.ToUTokenDenom(token.Denom)
	if uTokenDenom == "" {
		return sdk.Coin{}, types.ErrUToken.Wrap(token.Denom)
	}

	exchangeRate := k.DeriveExchangeRate(ctx, token.Denom)
	uTokenAmount := toDec(token.Amount).Quo(exchangeRate).TruncateInt()
	return sdk.NewCoin(uTokenDenom, uTokenAmount), nil
}

// UToken2TokenRate returns a token amount user would receive when supplying the uToken.
// Returns error if the uToken is not a uToken.
func (k Keeper) UToken2TokenRate(ctx sdk.Context, uToken sdk.Coin) (sdk.Coin, error) {
	if err := uToken.Validate(); err != nil {
		return sdk.Coin{}, err
	}

	tokenDenom := types.ToTokenDenom(uToken.Denom)
	if tokenDenom == "" {
		return sdk.Coin{}, types.ErrNotUToken.Wrap(uToken.Denom)
	}

	exchangeRate := k.DeriveExchangeRate(ctx, tokenDenom)

	tokenAmount := toDec(uToken.Amount).Mul(exchangeRate).TruncateInt()
	return sdk.NewCoin(tokenDenom, tokenAmount), nil
}

// Tokens2uTokensRate converts an sdk.Coins containing uTokens to their values in base
// tokens.
func (k Keeper) Tokens2uTokensRate(ctx sdk.Context, uTokens sdk.Coins) (sdk.Coins, error) {
	if err := uTokens.Validate(); err != nil {
		return sdk.Coins{}, err
	}

	tokens := sdk.Coins{}
	for _, coin := range uTokens {
		token, err := k.UToken2TokenRate(ctx, coin)
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
