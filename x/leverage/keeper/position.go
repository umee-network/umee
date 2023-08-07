package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v5/util/coin"
	"github.com/umee-network/umee/v5/x/leverage/types"
)

// minimum borrow factor is the minimum collateral weight and minimum liquidation threshold
// allowed when a borrowed token is limiting the efficiency of a pair of assets.
// TODO: parameterize this in the leverage module
var minimumBorrowFactor = sdk.MustNewDecFromStr("0.5")

// GetAccountPosition creates and sorts an accountPosition for an address, using information
// from the keeper's special asset pairs and token collateral weights as well as oracle prices.
// Will treat collateral with missing prices as zero-valued, but will error on missing borrow prices.
// On computing liquidation threshold, will treat borrows with missing prices as zero and error on
// missing collateral prices instead, as well as using spot prices instead of both spot and historic.
// Also stores all token settings and any special asset pairs that could apply to the account's collateral.
func (k Keeper) GetAccountPosition(ctx sdk.Context, addr sdk.AccAddress, isForLiquidation bool,
) (types.AccountPosition, error) {
	tokenSettings := k.GetAllRegisteredTokens(ctx)
	specialPairs := []types.SpecialAssetPair{}
	collateral := k.GetBorrowerCollateral(ctx, addr)
	collateralValue := sdk.NewDecCoins()
	borrowed := k.GetBorrowerBorrows(ctx, addr)
	borrowedValue := sdk.NewDecCoins()

	var (
		v   sdk.Dec
		err error
	)

	// get the borrower's collateral value by token
	for _, c := range collateral {
		if isForLiquidation {
			// for liquidation threshold, error on collateral without prices
			// and use spot price
			v, err = k.CalculateCollateralValue(ctx, sdk.NewCoins(c), types.PriceModeSpot)
		} else {
			// for borrow limit, max borrow, and max withdraw, ignore collateral without prices
			// and use the lower of historic or spot prices
			v, err = k.VisibleCollateralValue(ctx, sdk.NewCoins(c), types.PriceModeLow)
		}
		if err != nil {
			return types.AccountPosition{}, err
		}
		denom := coin.StripUTokenDenom(c.Denom)
		collateralValue = collateralValue.Add(sdk.NewDecCoinFromDec(denom, v))
		// get special asset pairs which could apply to this collateral token
		specialPairs = append(specialPairs, k.GetSpecialAssetPairs(ctx, c.Denom)...)
	}

	// get the borrower's borrowed value by token
	for _, b := range borrowed {
		if isForLiquidation {
			// for liquidation threshold, ignore borrow without prices
			// and use spot price
			v, err = k.VisibleTokenValue(ctx, sdk.NewCoins(b), types.PriceModeSpot)
		} else {
			// for borrow limit, max borrow, and max withdraw, error on borrow without prices
			// and use the higher of historic or spot prices
			v, err = k.TokenValue(ctx, b, types.PriceModeHigh)
		}
		if err != nil {
			return types.AccountPosition{}, err
		}
		if v.IsPositive() {
			borrowedValue = borrowedValue.Add(sdk.NewDecCoinFromDec(b.Denom, v))
		}
	}

	return types.NewAccountPosition(
		tokenSettings, specialPairs, collateralValue, borrowedValue, isForLiquidation, minimumBorrowFactor,
	)
}
