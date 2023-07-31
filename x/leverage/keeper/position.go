package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v5/x/leverage/types"
)

// getAccountPosition creates and sorts an accountPosition for an address, using information
// from the keeper's special asset pairs and token collateral weights as well as oracle prices.
// Will treat collateral with missing prices as zero-valued, but will error on missing borrow prices.
// Also stores all token settings and any special asset pairs that could apply to the account's collateral.
func (k Keeper) getAccountPosition(ctx sdk.Context, addr sdk.AccAddress) (types.AccountPosition, error) {
	tokenSettings := k.GetAllRegisteredTokens(ctx)
	specialPairs := []types.SpecialAssetPair{}
	collateral := k.GetBorrowerCollateral(ctx, addr)
	collateralValue := sdk.NewDecCoins()
	borrowed := k.GetBorrowerBorrows(ctx, addr)
	borrowedValue := sdk.NewDecCoins()

	// get the borrower's collateral value by token
	for _, c := range collateral {
		v, err := k.VisibleCollateralValue(ctx, sdk.NewCoins(c), types.PriceModeLow)
		if err != nil {
			return types.AccountPosition{}, err
		}
		denom := types.ToTokenDenom(c.Denom)
		collateralValue = collateralValue.Add(sdk.NewDecCoinFromDec(denom, v))
		// get special asset pairs which could apply to this collateral token
		specialPairs = append(specialPairs, k.GetSpecialAssetPairs(ctx, c.Denom)...)
	}

	// get the borrower's borrowed value by token, sorted by collateral weight (descending)
	for _, b := range borrowed {
		v, err := k.TokenValue(ctx, b, types.PriceModeHigh)
		if err != nil {
			return types.AccountPosition{}, err
		}
		if v.IsPositive() {
			denom := types.ToTokenDenom(b.Denom)
			borrowedValue = borrowedValue.Add(sdk.NewDecCoinFromDec(denom, v))
		}
	}

	return types.NewAccountPosition(tokenSettings, specialPairs, collateralValue, borrowedValue), nil
}
