package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	"github.com/umee-network/umee/x/leverage/types"
)

// GetPrice returns the USD value of an sdk.Coin containing base assets
func (k Keeper) GetPrice(ctx sdk.Context, coin sdk.Coin) (sdk.Dec, error) {
	if !k.IsAcceptedToken(ctx, coin.Denom) {
		return sdk.ZeroDec(), sdkerrors.Wrap(types.ErrInvalidAsset, coin.Denom)
	}
	// TODO #97: Use oracle module (as well as denom metadata from x/bank)
	return coin.Amount.ToDec(), nil
}

// GetTotalPrice returns the USD value of an sdk.Coins containing base assets
func (k Keeper) GetTotalPrice(ctx sdk.Context, coins sdk.Coins) (sdk.Dec, error) {
	value := sdk.ZeroDec()
	for _, coin := range coins {
		v, err := k.GetPrice(ctx, coin)
		if err != nil {
			return sdk.ZeroDec(), err
		}
		value = value.Add(v)
	}
	return value, nil
}

// EquivalentValue returns the amount of a selected denom which would have equal
// USD value to a provided sdk.Coin
func (k Keeper) EquivalentValue(ctx sdk.Context, coin sdk.Coin, toDenom string) (sdk.Coin, error) {
	value, err := k.GetPrice(ctx, coin)
	if err != nil {
		return sdk.Coin{}, err
	}

	// first derive USD value of new denom if amount was unchanged
	exchange, err := k.GetPrice(ctx, sdk.NewCoin(toDenom, coin.Amount))
	if err != nil {
		return sdk.Coin{}, err
	}
	if !exchange.IsPositive() {
		return sdk.Coin{}, sdkerrors.Wrap(types.ErrBadValue, exchange.String())
	}

	// then return the amount corrected by the price ratio
	return sdk.NewCoin(
		toDenom,
		coin.Amount.ToDec().Mul(value).Quo(exchange).TruncateInt(),
	), nil
}
