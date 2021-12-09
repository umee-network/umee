package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/umee-network/umee/x/leverage/types"
)

// TokenPrice returns the USD value of a base token. Note, the token's denomination
// must be the base denomination, e.g. uumee. The x/oracle module must know of
// the base and display/symbol denominations for each exchange pair. E.g. it must
// know about the UMEE/USD exchange rate along with the uumee base denomination
// and the exponent.
func (k Keeper) TokenPrice(ctx sdk.Context, denom string) (sdk.Dec, error) {
	if !k.IsAcceptedToken(ctx, denom) {
		return sdk.ZeroDec(), sdkerrors.Wrap(types.ErrInvalidAsset, denom)
	}

	return k.oracleKeeper.GetExchangeRateBase(ctx, denom)
}

// TokenValue returns the total token value given a Coin. An error is
// returned if we cannot get the token's price or if it's not an accepted token.
func (k Keeper) TokenValue(ctx sdk.Context, coin sdk.Coin) (sdk.Dec, error) {
	p, err := k.TokenPrice(ctx, coin.Denom)
	if err != nil {
		return sdk.ZeroDec(), err
	}

	return p.Mul(coin.Amount.ToDec()), nil
}

// TotalTokenValue returns the total value of all supplied tokens. It is
// equivalent to calling GetTokenValue on each coin individually.
func (k Keeper) TotalTokenValue(ctx sdk.Context, coins sdk.Coins) (sdk.Dec, error) {
	total := sdk.ZeroDec()

	for _, c := range coins {
		v, err := k.TokenValue(ctx, c)
		if err != nil {
			return sdk.ZeroDec(), nil
		}

		total = total.Add(v)
	}

	return total, nil
}

// EquivalentValue returns the amount of a selected denom which would have equal
// USD value to a provided sdk.Coin
func (k Keeper) EquivalentTokenValue(ctx sdk.Context, fromCoin sdk.Coin, toDenom string) (sdk.Coin, error) {
	// get total USD price of input (from) denomination
	price, err := k.TokenPrice(ctx, fromCoin.Denom)
	if err != nil {
		return sdk.Coin{}, err
	}

	// return immediately on zero price
	if price.IsZero() {
		return sdk.NewCoin(toDenom, sdk.ZeroInt()), nil
	}

	// first derive USD value of new denom if amount was unchanged
	exchange, err := k.TokenPrice(ctx, toDenom)
	if err != nil {
		return sdk.Coin{}, err
	}
	if !exchange.IsPositive() {
		return sdk.Coin{}, sdkerrors.Wrap(types.ErrBadValue, exchange.String())
	}

	// then return the amount corrected by the price ratio
	return sdk.NewCoin(
		toDenom,
		fromCoin.Amount.ToDec().Mul(price).Quo(exchange).TruncateInt(),
	), nil
}
