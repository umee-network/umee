package keeper

import (
	"strings"

	"cosmossdk.io/errors"
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v6/util/coin"
	"github.com/umee-network/umee/v6/util/sdkutil"
	"github.com/umee-network/umee/v6/x/auction"
	"github.com/umee-network/umee/v6/x/leverage/types"
	oracletypes "github.com/umee-network/umee/v6/x/oracle/types"
)

// TODO: parameterize this
const MaxSpotPriceAge = 180 // 180 seconds = 3 minutes

var ten = sdk.MustNewDecFromStr("10")

// TokenPrice returns the USD value of a token's symbol denom, e.g. `UMEE` (rather than `uumee`).
// Note, the input denom must still be the base denomination, e.g. uumee. When error is nil, price is
// guaranteed to be positive. Also returns the token's exponent to reduce redundant registry reads.
func (k Keeper) TokenPrice(ctx sdk.Context, baseDenom string, mode types.PriceMode) (sdk.Dec, uint32, error) {
	t, err := k.GetTokenSettings(ctx, baseDenom)
	if err != nil {
		return sdk.ZeroDec(), 0, err
	}
	if t.Blacklist {
		return sdk.ZeroDec(), t.Exponent, types.ErrBlacklisted
	}

	// if a token is exempt from historic pricing, all price modes ignore historic prices
	// and use spot prices instead, sometimes also allowing expired prices.
	if t.HistoricMedians == 0 {
		mode = mode.IgnoreHistoric()
	}

	var price, historicPrice sdk.Dec
	var spotPrice oracletypes.ExchangeRate
	if mode != types.PriceModeHistoric {
		// spot price is required for modes other than historic
		spotPrice, err = k.oracleKeeper.GetExchangeRate(ctx, t.SymbolDenom)
		if err != nil {
			return sdk.ZeroDec(), t.Exponent, errors.Wrap(err, "oracle")
		}
		if !mode.AllowsExpired() {
			// with the exception of account summary queries, require spot prices to be recent
			moduleTime := k.getLastInterestTime(ctx)
			priceTime := spotPrice.Timestamp.Unix()
			priceAge := moduleTime - priceTime
			if priceAge < 0 || priceAge > MaxSpotPriceAge {
				return sdk.ZeroDec(), t.Exponent, types.ErrExpiredOraclePrice.Wrapf(
					"price: %d, module: %d", priceTime, moduleTime)
			}
		}
	}

	if mode != types.PriceModeSpot && mode != types.PriceModeQuery {
		// historic price is required for modes other than spot and query
		var numStamps uint32
		historicPrice, numStamps, err = k.oracleKeeper.MedianOfHistoricMedians(
			ctx, strings.ToUpper(t.SymbolDenom), uint64(t.HistoricMedians))
		if err != nil {
			return sdk.ZeroDec(), t.Exponent, errors.Wrap(err, "oracle")
		}
		if numStamps < t.HistoricMedians {
			return sdk.ZeroDec(), t.Exponent, types.ErrNoHistoricMedians.Wrapf(
				"requested %d, got %d",
				t.HistoricMedians,
				numStamps,
			)
		}
	}

	switch mode {
	case types.PriceModeSpot, types.PriceModeQuery:
		price = spotPrice.Rate
	case types.PriceModeHistoric:
		price = historicPrice
	case types.PriceModeHigh, types.PriceModeQueryHigh:
		price = sdk.MaxDec(spotPrice.Rate, historicPrice)
	case types.PriceModeLow, types.PriceModeQueryLow:
		price = sdk.MinDec(spotPrice.Rate, historicPrice)
	default:
		return sdk.ZeroDec(), t.Exponent, types.ErrInvalidPriceMode.Wrapf("%d", mode)
	}

	if price.IsNil() || !price.IsPositive() {
		return sdk.ZeroDec(), t.Exponent, types.ErrInvalidOraclePrice.Wrap(baseDenom)
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
func (k Keeper) TokenValue(ctx sdk.Context, coin sdk.Coin, mode types.PriceMode) (sdk.Dec, error) {
	p, exp, err := k.TokenPrice(ctx, coin.Denom, mode)
	if err != nil {
		return sdk.ZeroDec(), err
	}
	return exponent(p.Mul(toDec(coin.Amount)), int32(exp)*-1), nil
}

// TotalTokenValue returns the total value of all supplied tokens. It is
// equivalent to the sum of TokenValue on each coin individually, except it
// ignores unregistered and blacklisted tokens instead of returning an error.
func (k Keeper) TotalTokenValue(ctx sdk.Context, coins sdk.Coins, mode types.PriceMode) (sdk.Dec, error) {
	total := sdk.ZeroDec()

	accepted := k.filterAcceptedCoins(ctx, coins)

	for _, c := range accepted {
		v, err := k.TokenValue(ctx, c, mode)
		if err != nil {
			return sdk.ZeroDec(), err
		}

		total = total.Add(v)
	}

	return total, nil
}

// ValueWithBorrowFactor returns the total value of all input tokens, each multiplied
// by borrow factor (which is the minimum of 2.0 and 1/collateral weight). It
// ignores unregistered and blacklisted tokens instead of returning an error, but
// will error on unavailable prices.
func (k Keeper) ValueWithBorrowFactor(ctx sdk.Context, coins sdk.Coins, mode types.PriceMode) (sdk.Dec, error) {
	total := sdk.ZeroDec()

	for _, c := range coins {
		token, err := k.GetTokenSettings(ctx, c.Denom)
		if err != nil {
			continue
		}
		v, err := k.TokenValue(ctx, c, mode)
		if err != nil {
			return sdk.ZeroDec(), err
		}

		total = total.Add(v.Mul(token.BorrowFactor()))
	}

	return total, nil
}

// VisibleTokenValue functions like TotalTokenValue, but interprets missing oracle prices
// as zero value instead of returning an error.
func (k Keeper) VisibleTokenValue(ctx sdk.Context, coins sdk.Coins, mode types.PriceMode) (sdk.Dec, error) {
	total := sdk.ZeroDec()

	accepted := k.filterAcceptedCoins(ctx, coins)

	for _, c := range accepted {
		v, err := k.TokenValue(ctx, c, mode)
		if err == nil {
			total = total.Add(v)
		}
		if nonOracleError(err) {
			return sdk.ZeroDec(), err
		}
	}

	return total, nil
}

// VisibleUTokensValue converts uTokens to tokens and calls VisibleTokenValue. Errors on non-uTokens.
func (k Keeper) VisibleUTokensValue(ctx sdk.Context, uTokens sdk.Coins, mode types.PriceMode) (sdk.Dec, error) {
	tokens := sdk.NewCoins()

	for _, u := range uTokens {
		t, err := k.ToToken(ctx, u)
		if err != nil {
			return sdk.ZeroDec(), err
		}
		tokens = tokens.Add(t)
	}

	return k.VisibleTokenValue(ctx, tokens, mode)
}

// TokenWithValue creates a token of a given denom with an given USD value.
// Returns an error on invalid price or denom. Rounds down, i.e. the
// value of the token returned may be slightly less than the requested value.
func (k Keeper) TokenWithValue(ctx sdk.Context, denom string, value sdk.Dec, mode types.PriceMode) (sdk.Coin, error) {
	// get token price (guaranteed positive if nil error) and exponent
	price, exp, err := k.TokenPrice(ctx, denom, mode)
	if err != nil {
		return sdk.Coin{}, err
	}

	// amount = USD value * 10^exponent / symbol price
	amount := exponent(value, int32(exp)).Quo(price)
	return sdk.NewCoin(denom, amount.TruncateInt()), nil
}

// UTokenWithValue creates a uToken of a given denom with an given USD value.
// Returns an error on invalid price or non-uToken denom. Rounds down, i.e. the
// value of the uToken returned may be slightly less than the requested value.
func (k Keeper) UTokenWithValue(ctx sdk.Context, denom string, value sdk.Dec, mode types.PriceMode) (sdk.Coin, error) {
	base := coin.StripUTokenDenom(denom)
	if base == "" {
		return sdk.Coin{}, types.ErrNotUToken.Wrap(denom)
	}

	token, err := k.TokenWithValue(ctx, base, value, mode)
	if err != nil {
		return sdk.Coin{}, err
	}

	uTokenExchangeRate := k.DeriveExchangeRate(ctx, base)
	uTokenAmount := sdk.NewDecFromInt(token.Amount).Quo(uTokenExchangeRate).TruncateInt()

	return sdk.NewCoin(denom, uTokenAmount), nil
}

// PriceRatio computes the ratio of the USD prices of two base tokens, as sdk.Dec(fromPrice/toPrice).
// Will return an error if either token price is not positive, and guarantees a positive output.
// Computation uses price of token's symbol denom to avoid rounding errors for exponent >= 18 tokens,
// but returns in terms of base tokens. Uses the same price mode for both token denoms involved.
func (k Keeper) PriceRatio(ctx sdk.Context, fromDenom, toDenom string, mode types.PriceMode) (sdk.Dec, error) {
	p1, e1, err := k.TokenPrice(ctx, fromDenom, mode)
	if err != nil {
		return sdk.ZeroDec(), err
	}
	p2, e2, err := k.TokenPrice(ctx, toDenom, mode)
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

// fundModules transfers requested coins to other module account, as
// long as the leverage module account has sufficient unreserved assets.
func (k Keeper) fundModules(ctx sdk.Context, toOracle, toAuction sdk.Coins) error {
	toOracleCheck := sdk.Coins{}
	toAuctionCheck := sdk.Coins{}
	available := map[string]sdkmath.Int{}

	// reduce rewards if they exceed unreserved module balance
	for _, o := range toOracle {
		avl := k.AvailableLiquidity(ctx, o.Denom)
		amt := sdk.MinInt(o.Amount, avl)
		if amt.IsPositive() {
			toOracleCheck = toOracleCheck.Add(sdk.NewCoin(o.Denom, amt))
			avl.Sub(amt)
		}
		available[o.Denom] = avl
	}
	for _, o := range toAuction {
		avl, ok := available[o.Denom]
		if !ok {
			avl = k.AvailableLiquidity(ctx, o.Denom)
		}
		amt := sdk.MinInt(o.Amount, avl)
		if amt.IsPositive() {
			toAuctionCheck = toAuctionCheck.Add(sdk.NewCoin(o.Denom, amt))
			avl.Sub(amt)
		}
		available[o.Denom] = avl
	}

	// This action is caused by end blocker, not a message handler, so we need to emit an event
	k.Logger(ctx).Debug(
		"funded ",
		"to oracle", toOracleCheck,
		"to auction", toAuctionCheck,
	)
	send := k.bankKeeper.SendCoinsFromModuleToModule
	if !toOracleCheck.IsZero() {
		sdkutil.Emit(&ctx, &types.EventFundOracle{Assets: toOracleCheck})
		if err := send(ctx, types.ModuleName, oracletypes.ModuleName, toOracleCheck); err != nil {
			return err
		}
	}
	if !toAuctionCheck.IsZero() {
		auction.EmitFundRewardsAuction(&ctx, toOracleCheck)
		return k.bankKeeper.SendCoinsFromModuleToAccount(ctx, types.ModuleName, k.rewardsAuction, toAuctionCheck)
	}

	return nil
}
