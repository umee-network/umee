package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"

	oracletypes "github.com/umee-network/umee/v4/x/oracle/types"

	"github.com/umee-network/umee/v4/x/leverage/types"
)

var ten = sdk.MustNewDecFromStr("10")

// TokenBasePrice returns the USD value of a base token. Note, the token's denomination
// must be the base denomination, e.g. uumee. The x/oracle module must know of
// the base and display/symbol denominations for each exchange pair. E.g. it must
// know about the UMEE/USD exchange rate along with the uumee base denomination
// and the exponent. When error is nil, price is guaranteed to be positive.
func (k Keeper) TokenBasePrice(ctx sdk.Context, baseDenom string) (sdk.Dec, error) {
	t, err := k.GetTokenSettings(ctx, baseDenom)
	if err != nil {
		return sdk.ZeroDec(), err
	}

	if t.Blacklist {
		return sdk.ZeroDec(), types.ErrBlacklisted
	}

	price, err := k.oracleKeeper.GetExchangeRateBase(ctx, baseDenom)
	if err != nil {
		return sdk.ZeroDec(), sdkerrors.Wrap(err, "oracle")
	}

	if price.IsNil() || !price.IsPositive() {
		return sdk.ZeroDec(), sdkerrors.Wrap(types.ErrInvalidOraclePrice, baseDenom)
	}

	return price, nil
}

// TokenDefaultDenomPrice returns the USD value of a token's symbol denom, e.g. `UMEE` (rather than `uumee`).
// Note, the input denom must still be the base denomination, e.g. uumee. When error is nil, price is
// guaranteed to be positive. Also returns the token's exponent to reduce redundant registry reads.
// If the historic parameter is true, uses a median of recent prices instead of current price.
func (k Keeper) TokenDefaultDenomPrice(ctx sdk.Context, baseDenom string, historic bool) (sdk.Dec, uint32, error) {
	t, err := k.GetTokenSettings(ctx, baseDenom)
	if err != nil {
		return sdk.ZeroDec(), 0, err
	}

	if t.Blacklist {
		return sdk.ZeroDec(), t.Exponent, types.ErrBlacklisted
	}

	var price sdk.Dec

	if historic && t.HistoricMedians > 0 {
		// historic price
		var numStamps uint32
		price, numStamps, err = k.oracleKeeper.MedianOfHistoricMedians(ctx, t.SymbolDenom, uint64(t.HistoricMedians))
		if err == nil && numStamps < t.HistoricMedians {
			return sdk.ZeroDec(), t.Exponent, types.ErrNoHistoricMedians.Wrapf(
				"requested %d, got %d",
				t.HistoricMedians,
				numStamps,
			)
		}
	} else {
		// current price
		price, err = k.oracleKeeper.GetExchangeRate(ctx, t.SymbolDenom)
	}
	if err != nil {
		return sdk.ZeroDec(), t.Exponent, sdkerrors.Wrap(err, "oracle")
	}

	if price.IsNil() || !price.IsPositive() {
		return sdk.ZeroDec(), t.Exponent, sdkerrors.Wrap(types.ErrInvalidOraclePrice, baseDenom)
	}

	return price, t.Exponent, nil
}

// exponent multiplies an sdk.Dec by 10^n. n can be negative.
func exponent(input sdk.Dec, n int32) sdk.Dec {
	if n == 0 {
		return input
	}
	if n < 0 {
		quotient := ten.Power(uint64(n * -1))
		return input.Quo(quotient)
	}
	return input.Mul(ten.Power(uint64(n)))
}

// TokenValue returns the total token value given a Coin. An error is
// returned if we cannot get the token's price or if it's not an accepted token.
// Computation uses price of token's default denom to avoid rounding errors
// for exponent >= 18 tokens.
// If the historic parameter is true, uses medians of recent prices instead of current prices.
func (k Keeper) TokenValue(ctx sdk.Context, coin sdk.Coin, historic bool) (sdk.Dec, error) {
	p, exp, err := k.TokenDefaultDenomPrice(ctx, coin.Denom, historic)
	if err != nil {
		return sdk.ZeroDec(), err
	}
	return exponent(p.Mul(toDec(coin.Amount)), int32(exp)*-1), nil
}

// TotalTokenValue returns the total value of all supplied tokens. It is
// equivalent to the sum of TokenValue on each coin individually, except it
// ignores unregistered and blacklisted tokens instead of returning an error.
// If the historic parameter is true, uses medians of recent prices instead of current prices.
func (k Keeper) TotalTokenValue(ctx sdk.Context, coins sdk.Coins, historic bool) (sdk.Dec, error) {
	total := sdk.ZeroDec()

	accepted := k.filterAcceptedCoins(ctx, coins)

	for _, c := range accepted {
		v, err := k.TokenValue(ctx, c, historic)
		if err != nil {
			return sdk.ZeroDec(), err
		}

		total = total.Add(v)
	}

	return total, nil
}

// TokenWithValue creates a token of a given denom with an given USD value, using either current
// or historic prices. Returns an error on invalid price or denom.
func (k Keeper) TokenWithValue(ctx sdk.Context, denom string, value sdk.Dec, historic bool) (sdk.Coin, error) {
	// get token price (guaranteed positive if nil error) and exponent
	price, exp, err := k.TokenDefaultDenomPrice(ctx, denom, historic)
	if err != nil {
		return sdk.Coin{}, err
	}

	// amount = USD value * 10^exponent / symbol price
	amount := exponent(value, int32(exp)).Quo(price)
	return sdk.NewCoin(denom, amount.TruncateInt()), nil
}

// PriceRatio computes the ratio of the USD prices of two base tokens, as sdk.Dec(fromPrice/toPrice).
// Will return an error if either token price is not positive, and guarantees a positive output.
// Computation uses price of token's default denom to avoid rounding errors for exponent >= 18 tokens,
// but returns in terms of base tokens.
// If the historic parameter is true, uses medians of recent prices instead of current prices.
func (k Keeper) PriceRatio(ctx sdk.Context, fromDenom, toDenom string, historic bool) (sdk.Dec, error) {
	p1, e1, err := k.TokenDefaultDenomPrice(ctx, fromDenom, historic)
	if err != nil {
		return sdk.ZeroDec(), err
	}
	p2, e2, err := k.TokenDefaultDenomPrice(ctx, toDenom, historic)
	if err != nil {
		return sdk.ZeroDec(), err
	}
	// If tokens have different exponents, the symbol price ratio must be adjusted
	// to obtain the base token price ratio. If fromDenom has a higher exponent, then
	// the ratio p1/p2 must be adjusted lower.
	powerDifference := int32(e2) - int32(e1)
	// Price ratio > 1 if fromDenom is worth more than toDenom.
	return exponent(p1, powerDifference).Quo(p2), nil
}

// FundOracle transfers requested coins to the oracle module account, as
// long as the leverage module account has sufficient unreserved assets.
func (k Keeper) FundOracle(ctx sdk.Context, requested sdk.Coins) error {
	rewards := sdk.Coins{}

	// reduce rewards if they exceed unreserved module balance
	for _, coin := range requested {
		amountToTransfer := sdk.MinInt(coin.Amount, k.AvailableLiquidity(ctx, coin.Denom))

		if amountToTransfer.IsPositive() {
			rewards = rewards.Add(sdk.NewCoin(coin.Denom, amountToTransfer))
		}
	}

	// This action is not caused by a message so we need to make an event here
	k.Logger(ctx).Debug(
		"funded oracle",
		"amount", rewards,
	)
	if err := ctx.EventManager().EmitTypedEvent(&types.EventFundOracle{Assets: rewards}); err != nil {
		return err
	}

	// Send rewards
	if !rewards.IsZero() {
		return k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, oracletypes.ModuleName, rewards)
	}

	return nil
}
