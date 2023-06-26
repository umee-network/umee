package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v5/x/leverage/types"
)

// assertBorrowerHealth returns an error if a borrower is currently above their borrow limit,
// under either recent (historic median) or current prices. Checks using borrow limit based
// on collateral weight, then check separately for borrow limit using borrow factor. Error if
// borrowed asset prices cannot be calculated, but will try to treat collateral whose prices are
// unavailable as having zero value. This can still result in a borrow limit being too low,
// unless the remaining collateral is enough to cover all borrows.
// This should be checked in msg_server.go at the end of any transaction which is restricted
// by borrow limits, i.e. Borrow, Decollateralize, Withdraw, MaxWithdraw.
// MaxUsage sets the maximum percent of a user's borrow limit that can be in use: set to 1
// to allow up to 100% borrow limit, or a lower value (e.g. 0.9) if a transaction should fail
// if a safety margin is desired (e.g. <90% borrow limit).
func (k Keeper) assertBorrowerHealth(ctx sdk.Context, borrowerAddr sdk.AccAddress, maxUsage sdk.Dec) error {
	borrowed := k.GetBorrowerBorrows(ctx, borrowerAddr)
	collateral := k.GetBorrowerCollateral(ctx, borrowerAddr)

	// check health using collateral weight
	borrowValue, err := k.TotalTokenValue(ctx, borrowed, types.PriceModeHigh)
	if err != nil {
		return err
	}
	borrowLimit, err := k.VisibleBorrowLimit(ctx, collateral)
	if err != nil {
		return err
	}
	if borrowValue.GT(borrowLimit.Mul(maxUsage)) {
		return types.ErrUndercollaterized.Wrapf(
			"borrowed: %s, limit: %s, max usage %s", borrowValue, borrowLimit, maxUsage)
	}

	// check health using borrow factor
	weightedBorrowValue, err := k.WeightedBorrowValue(ctx, borrowed, types.PriceModeHigh)
	if err != nil {
		return err
	}
	collateralValue, err := k.VisibleUTokenValue(ctx, collateral, types.PriceModeLow)
	if err != nil {
		return err
	}
	if weightedBorrowValue.GT(collateralValue.Mul(maxUsage)) {
		return types.ErrUndercollaterized.Wrapf(
			"weighted borrow: %s, collateral value: %s, max usage %s", weightedBorrowValue, collateralValue, maxUsage)
	}

	return nil
}

// GetBorrow returns an sdk.Coin representing how much of a given denom a
// borrower currently owes.
func (k Keeper) GetBorrow(ctx sdk.Context, borrowerAddr sdk.AccAddress, denom string) sdk.Coin {
	adjustedAmount := k.getAdjustedBorrow(ctx, borrowerAddr, denom)
	owedAmount := adjustedAmount.Mul(k.getInterestScalar(ctx, denom)).Ceil().TruncateInt()
	return sdk.NewCoin(denom, owedAmount)
}

// repayBorrow repays tokens borrowed by borrowAddr by sending coins in fromAddr to the module. This
// occurs during normal repayment (in which case fromAddr and borrowAddr are the same) and during
// liquidations, where fromAddr is the liquidator instead.
func (k Keeper) repayBorrow(ctx sdk.Context, fromAddr, borrowAddr sdk.AccAddress, repay sdk.Coin) error {
	err := k.bankKeeper.SendCoinsFromAccountToModule(ctx, fromAddr, types.ModuleName, sdk.NewCoins(repay))
	if err != nil {
		return err
	}
	return k.setBorrow(ctx, borrowAddr, k.GetBorrow(ctx, borrowAddr, repay.Denom).Sub(repay))
}

// moveBorrow transfers a debt from fromAddr to toAddr without moving any tokens. This occurs during
// fast liquidations, where a liquidator takes on a borrower's debt.
func (k Keeper) moveBorrow(ctx sdk.Context, fromAddr, toAddr sdk.AccAddress, repay sdk.Coin) error {
	err := k.setBorrow(ctx, fromAddr, k.GetBorrow(ctx, fromAddr, repay.Denom).Sub(repay))
	if err != nil {
		return err
	}
	return k.setBorrow(ctx, toAddr, k.GetBorrow(ctx, toAddr, repay.Denom).Add(repay))
}

// setBorrow sets the amount borrowed by an address in a given denom.
// If the amount is zero, any stored value is cleared.
func (k Keeper) setBorrow(ctx sdk.Context, borrowerAddr sdk.AccAddress, borrow sdk.Coin) error {
	// Apply interest scalar to determine adjusted amount
	newAdjustedAmount := toDec(borrow.Amount).Quo(k.getInterestScalar(ctx, borrow.Denom))

	// Set new borrow value
	return k.setAdjustedBorrow(ctx, borrowerAddr, sdk.NewDecCoinFromDec(borrow.Denom, newAdjustedAmount))
}

// GetTotalBorrowed returns the total borrowed in a given denom.
func (k Keeper) GetTotalBorrowed(ctx sdk.Context, denom string) sdk.Coin {
	adjustedTotal := k.getAdjustedTotalBorrowed(ctx, denom)

	// Apply interest scalar
	total := adjustedTotal.Mul(k.getInterestScalar(ctx, denom)).Ceil().TruncateInt()
	return sdk.NewCoin(denom, total)
}

// AvailableLiquidity gets the unreserved module balance of a given token.
func (k Keeper) AvailableLiquidity(ctx sdk.Context, denom string) sdkmath.Int {
	moduleBalance := k.ModuleBalance(ctx, denom).Amount
	reserveAmount := k.GetReserves(ctx, denom).Amount

	return sdk.MaxInt(moduleBalance.Sub(reserveAmount), sdk.ZeroInt())
}

// SupplyUtilization calculates the current supply utilization of a token denom.
func (k Keeper) SupplyUtilization(ctx sdk.Context, denom string) sdk.Dec {
	// Supply utilization is equal to total borrows divided by the token supply
	// (including borrowed tokens yet to be repaid and excluding tokens reserved).
	availableLiquidity := toDec(k.AvailableLiquidity(ctx, denom))
	totalBorrowed := toDec(k.GetTotalBorrowed(ctx, denom).Amount)
	tokenSupply := totalBorrowed.Add(availableLiquidity)

	// This edge case can be safely interpreted as 100% utilization.
	if totalBorrowed.GTE(tokenSupply) {
		return sdk.OneDec()
	}

	return totalBorrowed.Quo(tokenSupply)
}

// CalculateBorrowLimit uses the price oracle to determine the borrow limit (in USD) provided by
// collateral sdk.Coins, using each token's uToken exchange rate and collateral weight.
// The lower of spot price or historic price is used for each collateral token.
// An error is returned if any input coins are not uTokens or if value calculation fails.
func (k Keeper) CalculateBorrowLimit(ctx sdk.Context, collateral sdk.Coins) (sdk.Dec, error) {
	limit := sdk.ZeroDec()

	for _, coin := range collateral {
		// convert uToken collateral to base assets
		baseAsset, err := k.ExchangeUToken(ctx, coin)
		if err != nil {
			return sdk.ZeroDec(), err
		}

		ts, err := k.GetTokenSettings(ctx, baseAsset.Denom)
		if err != nil {
			return sdk.ZeroDec(), err
		}

		// ignore blacklisted tokens
		if !ts.Blacklist {
			// get USD value of base assets using the chosen price mode
			v, err := k.TokenValue(ctx, baseAsset, types.PriceModeLow)
			if err != nil {
				return sdk.ZeroDec(), err
			}
			// add each collateral coin's weighted value to borrow limit
			limit = limit.Add(v.Mul(ts.CollateralWeight))
		}
	}

	return limit, nil
}

// VisibleBorrowLimit uses the price oracle to determine the borrow limit (in USD) provided by
// collateral sdk.Coins, using each token's uToken exchange rate and collateral weight.
// The lower of spot price or historic price is used for each collateral token.
// An error is returned if any input coins are not uTokens.
// This function skips assets that are missing prices, which will lead to a lower borrow
// limit when prices are down instead of a complete loss of borrowing ability.
func (k Keeper) VisibleBorrowLimit(ctx sdk.Context, collateral sdk.Coins) (sdk.Dec, error) {
	limit := sdk.ZeroDec()

	for _, coin := range collateral {
		// convert uToken collateral to base assets
		baseAsset, err := k.ExchangeUToken(ctx, coin)
		if err != nil {
			return sdk.ZeroDec(), err
		}

		ts, err := k.GetTokenSettings(ctx, baseAsset.Denom)
		if err != nil {
			return sdk.ZeroDec(), err
		}

		// ignore blacklisted tokens
		if !ts.Blacklist {
			// get USD value of base assets using the chosen price mode
			v, err := k.TokenValue(ctx, baseAsset, types.PriceModeLow)
			if err == nil {
				// if both spot and historic (if required) prices exist,
				// add collateral coin's weighted value to borrow limit
				limit = limit.Add(v.Mul(ts.CollateralWeight))
			}
			if nonOracleError(err) {
				return sdk.ZeroDec(), err
			}
		}
	}

	return limit, nil
}

// CalculateLiquidationThreshold determines the maximum borrowed value (in USD) that a
// borrower with given collateral could reach before being eligible for liquidation, using
// each token's oracle price, uToken exchange rate, and liquidation threshold.
// An error is returned if any input coins are not uTokens or if value
// calculation fails. Always uses spot prices.
func (k Keeper) CalculateLiquidationThreshold(ctx sdk.Context, collateral sdk.Coins) (sdk.Dec, error) {
	totalThreshold := sdk.ZeroDec()

	for _, coin := range collateral {
		// convert uToken collateral to base assets
		baseAsset, err := k.ExchangeUToken(ctx, coin)
		if err != nil {
			return sdk.ZeroDec(), err
		}

		ts, err := k.GetTokenSettings(ctx, baseAsset.Denom)
		if err != nil {
			return sdk.ZeroDec(), err
		}

		// ignore blacklisted tokens
		if !ts.Blacklist {
			// get USD value of base assets
			v, err := k.TokenValue(ctx, baseAsset, types.PriceModeSpot)
			if err != nil {
				return sdk.ZeroDec(), err
			}

			// add each collateral coin's weighted value to liquidation threshold
			totalThreshold = totalThreshold.Add(v.Mul(ts.LiquidationThreshold))
		}
	}

	return totalThreshold, nil
}

// checkSupplyUtilization returns the appropriate error if a token denom's
// supply utilization has exceeded MaxSupplyUtilization
func (k Keeper) checkSupplyUtilization(ctx sdk.Context, denom string) error {
	token, err := k.GetTokenSettings(ctx, denom)
	if err != nil {
		return err
	}

	utilization := k.SupplyUtilization(ctx, denom)
	if utilization.GT(token.MaxSupplyUtilization) {
		return types.ErrMaxSupplyUtilization.Wrap(utilization.String())
	}
	return nil
}

// moduleMaxBorrow calculates maximum amount of Token to borrow from the module.
// The calculation first finds the maximum amount of Token that can be borrowed from the module,
// respecting the min_collateral_liquidity parameter, then determines the maximum amount of Token that can be borrowed
// from the module, respecting the max_supply_utilization parameter. The minimum between these two values is
// selected, given that the min_collateral_liquidity and max_supply_utilization are both limiting factors.
func (k Keeper) moduleMaxBorrow(ctx sdk.Context, denom string) (sdkmath.Int, error) {
	// Get the module_available_liquidity
	moduleAvailableLiquidity, err := k.ModuleAvailableLiquidity(ctx, denom)
	if err != nil {
		return sdk.ZeroInt(), err
	}

	// If module_available_liquidity is zero, we cannot borrow anything
	if !moduleAvailableLiquidity.IsPositive() {
		return sdk.ZeroInt(), nil
	}

	// Get max_supply_utilization for the denom
	token, err := k.GetTokenSettings(ctx, denom)
	if err != nil {
		return sdk.ZeroInt(), err
	}
	maxSupplyUtilization := token.MaxSupplyUtilization

	// Get total_borrowed from module for the denom
	totalBorrowed := k.GetTotalBorrowed(ctx, denom).Amount

	// Get module liquidity for the denom
	liquidity := k.AvailableLiquidity(ctx, denom)

	// The formula to calculate max_borrow respecting the max_supply_utilization is as follows:
	//
	// max_supply_utilization = (total_borrowed +  module_max_borrow) / (module_liquidity + total_borrowed)
	// module_max_borrow = max_supply_utilization * module_liquidity + max_supply_utilization * total_borrowed
	//						- total_borrowed
	moduleMaxBorrow := maxSupplyUtilization.MulInt(liquidity).Add(maxSupplyUtilization.MulInt(totalBorrowed)).Sub(
		sdk.NewDecFromInt(totalBorrowed),
	)

	// If module_max_borrow is zero, we cannot borrow anything
	if !moduleMaxBorrow.IsPositive() {
		return sdk.ZeroInt(), nil
	}

	// Use the minimum between module_max_borrow and module_available_liquidity
	return sdk.MinInt(moduleAvailableLiquidity, moduleMaxBorrow.TruncateInt()), nil
}
