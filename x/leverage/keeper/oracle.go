package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/umee-network/umee/x/leverage/types"
)

// Price returns the USD value of a base token. Note, the token's denomination
// must be the base denomination, e.g. uumee. The x/oracle module must know of
// the base and display/symbol denominations for each exchange pair. E.g. it must
// know about the UMEE/USD exchange rate along with the uumee base denomination
// and the exponent.
func (k Keeper) Price(ctx sdk.Context, denom string) (sdk.Dec, error) {
	if !k.IsAcceptedToken(ctx, denom) {
		return sdk.ZeroDec(), sdkerrors.Wrap(types.ErrInvalidAsset, denom)
	}

	return k.oracleKeeper.GetExchangeRateBase(ctx, denom)
}

// TotalPrice returns the total USD value of a set of Coins.
func (k Keeper) TotalPrice(ctx sdk.Context, denoms []string) (sdk.Dec, error) {
	price := sdk.ZeroDec()

	for _, denom := range denoms {
		p, err := k.Price(ctx, denom)
		if err != nil {
			return sdk.ZeroDec(), err
		}

		price = price.Add(p)
	}

	return price, nil
}

// EquivalentValue returns the amount of a selected denom which would have equal
// USD value to a provided sdk.Coin
func (k Keeper) EquivalentValue(ctx sdk.Context, fromCoin sdk.Coin, toDenom string) (sdk.Coin, error) {
	// get total USD price of input (from) denomination
	price, err := k.Price(ctx, fromCoin.Denom)
	if err != nil {
		return sdk.Coin{}, err
	}

	// return immediately on zero price
	if price.IsZero() {
		return sdk.NewCoin(toDenom, sdk.ZeroInt()), nil
	}

	// first derive USD value of new denom if amount was unchanged
	exchange, err := k.Price(ctx, toDenom)
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
