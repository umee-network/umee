package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v5/util/coin"
	"github.com/umee-network/umee/v5/x/leverage/types"
)

// accountPosition contains an account's borrowed and collateral values, arranged
// into special asset pairs and regular assets. Must be created by getAccountPosition.
// Each list will always be sorted by collateral weight. Also caches some relevant
// values that will be reused in computation, like token settings and the borrower's
// total borrowed and collateral tokens and total value.
type accountPosition struct {
	// all special asset pairs which apply to the account. Specifically, this should
	// cache any special pairs which match the account's collateral, even if they do
	// not match one of its borrows. This is the widest set of information that could
	// be needed when calculating maxWithdraw of an already collateralized asset, or
	// maxBorrow of any asset (even if not initially present in the position).
	// A pair that does not currently apply to the account (but could do so if a borrow
	// were added or existing pairs were rearranged) will have zero USD value
	// but will still be initialized with the proper weights, and sorted in order.
	specialPairs types.WeightedSpecialPairs
	// all collateral value not being used for special asset pairs
	collateral types.WeightedDecCoins
	// all borrowed value not being used for special asset pairs
	borrowed types.WeightedDecCoins
	// caches retrieved token settings
	tokens map[string]types.Token
	// collateral tokens
	collateralTokens sdk.Coins
	// borrowed tokens
	borrowedTokens sdk.Coins
	// collateral value (using PriceModeLow and interpreting unknown prices as zero)
	collateralValue sdk.Dec
	// borrowed value (using PriceModeHigh and requiring all prices known)
	borrowedValue sdk.Dec
}

func (ap *accountPosition) tokenCollateralWeight(ctx sdk.Context, k Keeper, denom string) sdk.Dec {
	if t, ok := ap.tokens[denom]; ok {
		return t.CollateralWeight
	}
	t, err := k.GetTokenSettings(ctx, denom)
	if err != nil {
		return sdk.ZeroDec()
	}
	ap.tokens[denom] = t
	return t.CollateralWeight
}

func (ap *accountPosition) tokenLiquidationThreshold(ctx sdk.Context, k Keeper, denom string) sdk.Dec {
	if t, ok := ap.tokens[denom]; ok {
		return t.LiquidationThreshold
	}
	t, err := k.GetTokenSettings(ctx, denom)
	if err != nil {
		return sdk.ZeroDec()
	}
	ap.tokens[denom] = t
	return t.LiquidationThreshold
}

// getAccountPosition creates and sorts an accountPosition for an address, using information
// from the keeper's special asset pairs and token collateral weights as well as oracle prices.
// Will treat collateral with missing prices as zero-valued, but will error on missing borrow prices.
func (k Keeper) getAccountPosition(ctx sdk.Context, addr sdk.AccAddress) (accountPosition, error) {
	position := accountPosition{
		// initialize
		specialPairs:    types.WeightedSpecialPairs{},
		borrowed:        types.WeightedDecCoins{},
		collateral:      types.WeightedDecCoins{},
		tokens:          map[string]types.Token{},
		borrowedValue:   sdk.ZeroDec(),
		collateralValue: sdk.ZeroDec(),
	}

	// get the borrower's collateral value by token, sorted by collateral weight (descending).
	// also gets any special pairs which could apply to existing collateral
	position.collateralTokens = k.GetBorrowerCollateral(ctx, addr)
	for _, c := range position.collateralTokens {
		v, err := k.VisibleCollateralValue(ctx, sdk.NewCoins(c), types.PriceModeLow)
		if err != nil {
			return accountPosition{}, err
		}
		if v.IsPositive() {
			denom := types.ToTokenDenom(c.Denom)
			position.collateral = position.collateral.Add(
				types.WeightedDecCoin{
					Asset:  sdk.NewDecCoinFromDec(denom, v),
					Weight: position.tokenCollateralWeight(ctx, k, denom),
				},
			)
			position.collateralValue = position.collateralValue.Add(v)
			// get relevant special asset pairs, initialized at zero USD value
			pairs := k.GetSpecialAssetPairs(ctx, c.Denom)
			for _, p := range pairs {
				wp := types.WeightedSpecialPair{
					Collateral:           coin.ZeroDec(denom),
					SpecialWeight:        p.CollateralWeight,
					LiquidationThreshold: p.LiquidationThreshold,
				}
				position.specialPairs = position.specialPairs.Add(wp)
			}
		}
	}

	// get the borrower's borrowed value by token, sorted by collateral weight (descending)
	position.borrowedTokens = k.GetBorrowerBorrows(ctx, addr)
	for _, b := range position.borrowedTokens {
		v, err := k.TokenValue(ctx, b, types.PriceModeHigh)
		if err != nil {
			return accountPosition{}, err
		}
		if v.IsPositive() {
			denom := types.ToTokenDenom(b.Denom)
			position.borrowed = position.borrowed.Add(
				types.WeightedDecCoin{
					Asset:  sdk.NewDecCoinFromDec(denom, v),
					Weight: position.tokenCollateralWeight(ctx, k, denom),
				},
			)
			position.borrowedValue = position.borrowedValue.Add(v)
		}
	}

	// TODO: match regular assets into special asset pairs

	return position, nil
}

// TODO: bump to the bottom, or top, when computing max borrow
// TODO: similar when computing max withdraw
// TODO: isolate special pairs and bump
