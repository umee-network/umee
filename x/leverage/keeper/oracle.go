package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	oracletypes "github.com/umee-network/umee/v2/x/oracle/types"

	"github.com/umee-network/umee/v2/x/leverage/types"
)

// TokenPrice returns the USD value of a base token. Note, the token's denomination
// must be the base denomination, e.g. uumee. The x/oracle module must know of
// the base and display/symbol denominations for each exchange pair. E.g. it must
// know about the UMEE/USD exchange rate along with the uumee base denomination
// and the exponent. When error is nil, price is guaranteed to be positive.
func (k Keeper) TokenPrice(ctx sdk.Context, denom string) (sdk.Dec, error) {
	t, err := k.GetTokenSettings(ctx, denom)
	if err != nil {
		return sdk.ZeroDec(), err
	}

	if t.Blacklist {
		return sdk.ZeroDec(), types.ErrBlacklisted
	}

	price, err := k.oracleKeeper.GetExchangeRateBase(ctx, denom)
	if err != nil {
		return sdk.ZeroDec(), sdkerrors.Wrap(err, "oracle")
	}

	if price.IsNil() || !price.IsPositive() {
		return sdk.ZeroDec(), sdkerrors.Wrap(types.ErrInvalidOraclePrice, denom)
	}

	return price, nil
}

// TokenValue returns the total token value given a Coin. An error is
// returned if we cannot get the token's price or if it's not an accepted token.
func (k Keeper) TokenValue(ctx sdk.Context, coin sdk.Coin) (sdk.Dec, error) {
	p, err := k.TokenPrice(ctx, coin.Denom)
	if err != nil {
		return sdk.ZeroDec(), err
	}
	return p.Mul(toDec(coin.Amount)), nil
}

// TotalTokenValue returns the total value of all supplied tokens. It is
// equivalent to the sum of TokenValue on each coin individually, except it
// ignores unregistered and blacklisted tokens instead of returning an error.
func (k Keeper) TotalTokenValue(ctx sdk.Context, coins sdk.Coins) (sdk.Dec, error) {
	total := sdk.ZeroDec()

	accepted := k.filterAcceptedCoins(ctx, coins)

	for _, c := range accepted {
		v, err := k.TokenValue(ctx, c)
		if err != nil {
			return sdk.ZeroDec(), err
		}

		total = total.Add(v)
	}

	return total, nil
}

// PriceRatio computed the ratio of the USD prices of two tokens, as sdk.Dec(fromPrice/toPrice).
// Will return an error if either token price is not positive, and guarantees a positive output.
func (k Keeper) PriceRatio(ctx sdk.Context, fromDenom, toDenom string) (sdk.Dec, error) {
	p1, err := k.TokenPrice(ctx, fromDenom)
	if err != nil {
		return sdk.ZeroDec(), err
	}
	p2, err := k.TokenPrice(ctx, toDenom)
	if err != nil {
		return sdk.ZeroDec(), err
	}
	// Price ratio > 1 if fromDenom is worth more than toDenom.
	return p1.Quo(p2), nil
}

// FundOracle transfers requested coins to the oracle module account, as
// long as the leverage module account has sufficient unreserved assets.
func (k Keeper) FundOracle(ctx sdk.Context, requested sdk.Coins) error {
	rewards := sdk.Coins{}

	// reduce rewards if they exceed unreserved module balance
	for _, coin := range requested {
		reserved := k.GetReserveAmount(ctx, coin.Denom)
		balance := k.ModuleBalance(ctx, coin.Denom)

		amountToTransfer := sdk.MinInt(coin.Amount, balance.Sub(reserved))

		if amountToTransfer.IsPositive() {
			rewards = rewards.Add(sdk.NewCoin(coin.Denom, amountToTransfer))
		}
	}

	// Because this action is not caused by a message, logging and
	// events are here instead of msg_server.go
	k.Logger(ctx).Debug(
		"funded oracle",
		"amount", rewards.String(),
	)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeFundOracle,
			sdk.NewAttribute(sdk.AttributeKeyAmount, rewards.String()),
		),
	})

	// Send rewards
	if !rewards.IsZero() {
		return k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, oracletypes.ModuleName, rewards)
	}

	return nil
}
