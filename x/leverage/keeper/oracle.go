package keeper

import (
	"strings"

	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v4/util/sdkutil"
	"github.com/umee-network/umee/v4/x/leverage/types"
	oracletypes "github.com/umee-network/umee/v4/x/oracle/types"
)

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

	// if a token is exempt from historic pricing, all price modes return Spot price
	if t.HistoricMedians == 0 {
		mode = types.PriceModeSpot
	}

	var price, spotPrice, historicPrice sdk.Dec
	if mode != types.PriceModeHistoric {
		// spot price is required for modes other than historic
		spotPrice, err = k.oracleKeeper.GetExchangeRate(ctx, t.SymbolDenom)
		if err != nil {
			return sdk.ZeroDec(), t.Exponent, errors.Wrap(err, "oracle")
		}
	}
	if mode != types.PriceModeSpot {
		// historic price is required for modes other than spot
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
	case types.PriceModeSpot:
		price = spotPrice
	case types.PriceModeHistoric:
		price = historicPrice
	case types.PriceModeHigh:
		price = sdk.MaxDec(spotPrice, historicPrice)
	case types.PriceModeLow:
		price = sdk.MinDec(spotPrice, historicPrice)
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
	base := types.ToTokenDenom(denom)
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
	sdkutil.Emit(&ctx, &types.EventFundOracle{Assets: rewards})

	// Send rewards
	if !rewards.IsZero() {
		return k.bankKeeper.SendCoinsFromModuleToModule(ctx, types.ModuleName, oracletypes.ModuleName, rewards)
	}

	return nil
}
