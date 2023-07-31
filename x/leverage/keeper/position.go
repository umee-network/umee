package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v5/x/leverage/types"
)

// accountPosition contains an account's borrowed and collateral values, arranged
// into special asset pairs and regular assets. If created by getAccountPosition,
// each list will always be sorted by collateral weight. Also caches some relevant
// values that will be reused in computation, like token settings.
type accountPosition struct {
	// all special asset pairs which apply to the account. Specifically, this should
	// cache any special pairs which match the account's collateral, even if they do
	// not match one of its borrows. This is the widest set of information that could
	// be needed when calculating maxWithdraw of an already collateralized asset, or
	// maxBorrow of any asset (even if not initially present in the position)
	specialPairs []types.WeightedSpecialPair
	// all collateral value not being used for special asset pairs
	collateral types.WeightedDecCoins
	// all borrowed value not being used for special asset pairs
	borrowed types.WeightedDecCoins
	// caches retrieved token settings
	tokens map[string]types.Token
}

func (ap *accountPosition) tokenCollateralWeight(ctx sdk.Context, k Keeper, denom string) sdk.Dec {
	if ap.tokens == nil {
		ap.tokens = map[string]types.Token{}
	}
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
	if ap.tokens == nil {
		ap.tokens = map[string]types.Token{}
	}
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
	position := accountPosition{}

	// get the borrower's collateral value by token, sorted by collateral weight (descending)
	collateral := k.GetBorrowerCollateral(ctx, addr)
	for _, c := range collateral {
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
		}
	}

	// get the borrower's borrowed value by token, sorted by collateral weight (descending)
	borrowed := k.GetBorrowerBorrows(ctx, addr)
	for _, b := range borrowed {
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
		}
	}

	// get all special pairs, sorted by collateral weight (descending)
	// pairs := k.GetAllSpecialAssetPairs(ctx)

	// TODO: take special pairs into account, sort the results, then return
	// Q: do we need to sort pairs by denom in GetAll, or just here, or not at all?

	return position, nil
}

// TODO: bump to the bottom, or top, when computing max borrow
// TODO: similar when computing max withdraw
// TODO: isolate special pairs and bump
