package keeper

import (
	"sort"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v5/x/leverage/types"
)

// accountPosition contains an account's borrowed and collateral values, arranged
// into special asset pairs and regular assets. If created by getAccountPosition,
// each list will always be sorted by collateral weight.
type accountPosition struct {
	// all special asset pairs which apply to the account
	specialPairs []weightedSpecialPair
	// all collateral value not being used for special asset pairs
	collateral weightedDecCoins
	// all borrowed value not being used for special asset pairs
	borrowed weightedDecCoins
}

// getAccountPosition creates and sorts an accountPosition for an address, using information
// from the keeper's special asset pairs and token collateral weights as well as oracle prices.
// Will treat collateral with missing prices as zero-valued, but will error on missing borrow prices.
func (k Keeper) getAccountPosition(ctx sdk.Context, addr sdk.AccAddress) (accountPosition, error) {
	position := accountPosition{}
	tokens := map[string]types.Token{}
	collateralWeight := func(denom string) sdk.Dec {
		if t, ok := tokens[denom]; ok {
			return t.CollateralWeight
		}
		t, err := k.GetTokenSettings(ctx, denom)
		if err != nil {
			return sdk.ZeroDec()
		}
		tokens[denom] = t
		return t.CollateralWeight
	}

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
				weightedDecCoin{
					asset:  sdk.NewDecCoinFromDec(denom, v),
					weight: collateralWeight(denom),
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
				weightedDecCoin{
					asset:  sdk.NewDecCoinFromDec(denom, v),
					weight: collateralWeight(denom),
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

type weightedDecCoin struct {
	// the USD value of an asset in a position
	asset sdk.DecCoin
	// the collateral weight
	weight sdk.Dec
}

// higher returns true if a weightedDecCoin wdc should be sorted after
// another weightedDecCoin b
func (wdc weightedDecCoin) higher(b weightedDecCoin) bool {
	if wdc.weight.GT(b.weight) {
		return true // sort first by collateral weight
	}
	return wdc.asset.Denom < b.asset.Denom // break ties by denom
}

type weightedDecCoins []weightedDecCoin

// Add returns the sum of a weightedDecCoins and a since additional weightedDecCoin.
// The result is sorted by collateral weight (descending) then denom (alphabetical).
func (wdc weightedDecCoins) Add(add weightedDecCoin) (sum weightedDecCoins) {
	if len(wdc) == 0 {
		return weightedDecCoins{add}
	}
	for _, c := range wdc {
		if c.asset.Denom == add.asset.Denom {
			sum = append(sum, weightedDecCoin{
				asset:  c.asset.Add(add.asset),
				weight: add.weight,
			})
		} else {
			sum = append(sum, c)
		}
	}
	// sorts the sum. Fixes unsorted input as well.
	sort.SliceStable(sum, func(i, j int) bool {
		return sum[i].higher(sum[j])
	})
	return sum
}

type weightedSpecialPair struct {
	// the collateral asset and its value and weight
	collateral weightedDecCoin
	// the borrowed asset and its value and weight
	borrow weightedDecCoin
	// the collateral weight of the special pair
	specialWeight sdk.Dec
}
