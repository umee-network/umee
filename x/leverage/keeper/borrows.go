package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v3/x/leverage/types"
)

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

// setBorrow sets the amount borrowed by an address in a given denom.
// If the amount is zero, any stored value is cleared.
func (k Keeper) setBorrow(ctx sdk.Context, borrowerAddr sdk.AccAddress, borrow sdk.Coin) error {
	// Apply interest scalar to determine adjusted amount
	newAdjustedAmount := toDec(borrow.Amount).Quo(k.getInterestScalar(ctx, borrow.Denom))

	// Set new borrow value
	if err := k.setAdjustedBorrow(ctx, borrowerAddr, sdk.NewDecCoinFromDec(borrow.Denom, newAdjustedAmount)); err != nil {
		return err
	}
	return nil
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
			// get USD value of base assets
			v, err := k.TokenValue(ctx, baseAsset)
			if err != nil {
				return sdk.ZeroDec(), err
			}
			// add each collateral coin's weighted value to borrow limit
			limit = limit.Add(v.Mul(ts.CollateralWeight))
		}
	}

	return limit, nil
}

// CalculateLiquidationThreshold determines the maximum borrowed value (in USD) that a
// borrower with given collateral could reach before being eligible for liquidation, using
// each token's oracle price, uToken exchange rate, and liquidation threshold.
// An error is returned if any input coins are not uTokens or if value
// calculation fails.
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
			v, err := k.TokenValue(ctx, baseAsset)
			if err != nil {
				return sdk.ZeroDec(), err
			}

			// add each collateral coin's weighted value to liquidation threshold
			totalThreshold = totalThreshold.Add(v.Mul(ts.LiquidationThreshold))
		}
	}

	return totalThreshold, nil
}
