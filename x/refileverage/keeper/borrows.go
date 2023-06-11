package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v5/x/refileverage/types"
)

// assertBorrowerHealth returns an error if a borrower is currently above their borrow limit,
// under either recent (historic median) or current prices. It returns an error if
// borrowed asset prices cannot be calculated, but will try to treat collateral whose prices are
// unavailable as having zero value. This can still result in a borrow limit being too low,
// unless the remaining collateral is enough to cover all borrows.
// This should be checked in msg_server.go at the end of any transaction which is restricted
// by borrow limits, i.e. Borrow, Decollateralize, Withdraw, MaxWithdraw.
func (k Keeper) assertBorrowerHealth(ctx sdk.Context, borrowerAddr sdk.AccAddress) error {
	borrowed := k.GetBorrowerBorrows(ctx, borrowerAddr)
	collateral := k.GetBorrowerCollateral(ctx, borrowerAddr)

	limit, err := k.VisibleBorrowLimit(ctx, collateral)
	if err != nil {
		return err
	}
	if borrowed.GT(limit) {
		return types.ErrUndercollaterized.Wrapf(
			"borrowed: %s, limit: %s", borrowed, limit.String())
	}
	return nil
}

// GetBorrow returns an sdk.Int representing how much $$$ given borrower currently owes.
func (k Keeper) GetBorrow(ctx sdk.Context, borrowerAddr sdk.AccAddress) sdk.Int {
	adjustedAmount := k.getAdjustedBorrow(ctx, borrowerAddr)
	return adjustedAmount.Mul(k.getInterestScalar(ctx, types.Gho)).Ceil().TruncateInt()
}

// repayBorrow repays tokens borrowed by borrowAddr by sending coins in fromAddr to the module. This
// occurs during normal repayment (in which case fromAddr and borrowAddr are the same) and during
// liquidations, where fromAddr is the liquidator instead.
func (k Keeper) repayBorrow(ctx sdk.Context, fromAddr, borrowAddr sdk.AccAddress, repay sdk.Int) error {
	return k.setBorrow(ctx, borrowAddr, k.GetBorrow(ctx, borrowAddr).Sub(repay))
}

// setBorrow sets the amount borrowed by an address in a given denom.
// If the amount is zero, any stored value is cleared.
func (k Keeper) setBorrow(ctx sdk.Context, borrowerAddr sdk.AccAddress, borrow sdk.Int) error {
	// Apply interest scalar to determine adjusted amount
	newAdjustedAmount := toDec(borrow).Quo(k.getInterestScalar(ctx, types.Gho))

	return k.setAdjustedBorrow(ctx, borrowerAddr, newAdjustedAmount)
}

// GetTotalBorrowed returns the total borrowed in a given denom.
func (k Keeper) GetTotalBorrowed(ctx sdk.Context) sdk.Int {
	adjustedTotal := k.getAdjustedTotalBorrowed(ctx)

	return adjustedTotal.Mul(k.getInterestScalar(ctx, types.Gho)).Ceil().TruncateInt()
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
