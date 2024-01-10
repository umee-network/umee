package keeper

import (
	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v6/x/leverage/types"
)

// assertBorrowerHealth returns an error if a borrower is currently above their borrow limit,
// under either recent (historic median) or current prices. Error if borrowed asset prices
// cannot be calculated, but will try to treat collateral whose prices are unavailable as
// having zero value. This can still result in a borrow limit being too low, unless the
// remaining collateral is enough to cover all borrows.
// This should be checked in msg_server.go at the end of any transaction which is restricted
// by borrow limits, i.e. Borrow, Decollateralize, Withdraw, MaxWithdraw, LeveragedLiquidate.
// MaxUsage sets the maximum percent of a user's borrow limit that can be in use: set to 1
// to allow up to 100% borrow limit, or a lower value (e.g. 0.9) if a transaction should fail
// if a safety margin is desired (e.g. <90% borrow limit).
func (k Keeper) assertBorrowerHealth(ctx sdk.Context, borrowerAddr sdk.AccAddress, maxUsage sdk.Dec) error {
	position, err := k.GetAccountPosition(ctx, borrowerAddr, false)
	if err != nil {
		return err
	}
	borrowedValue := position.BorrowedValue()
	borrowLimit := position.Limit()
	if borrowedValue.GT(borrowLimit.Mul(maxUsage)) {
		return types.ErrUndercollateralized.Wrapf(
			"borrowed: %s, limit: %s, max usage %s", borrowedValue, borrowLimit, maxUsage,
		)
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

// AvailableLiquiditySubMetokenSupply gets the unreserved module balance of a given token, considering meToken supply.
func (k Keeper) AvailableLiquiditySubMetokenSupply(ctx sdk.Context, denom string) (sdkmath.Int, error) {
	meTokenSupply, err := k.GetSupplied(ctx, k.meTokenAddr, denom)
	if err != nil {
		return sdk.ZeroInt(), err
	}

	return sdk.MaxInt(k.AvailableLiquidity(ctx, denom).Sub(meTokenSupply.Amount), sdk.ZeroInt()), nil
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
// In order to fully protect meToken supply, it's subtracted from the obtained amount.
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
	moduleMaxBorrow := sdk.MaxInt(
		maxSupplyUtilization.MulInt(liquidity).Add(maxSupplyUtilization.MulInt(totalBorrowed)).Sub(sdk.NewDecFromInt(totalBorrowed)).TruncateInt(),
		sdk.ZeroInt(),
	)

	// MeToken module supply is fully protected in order to guarantee its availability for redemption.
	meTokenSupply, err := k.GetSupplied(ctx, k.meTokenAddr, denom)
	if err != nil {
		return sdk.ZeroInt(), err
	}

	// Use the minimum between module_max_borrow and module_available_liquidity
	return sdk.MaxInt(
		sdk.MinInt(
			moduleAvailableLiquidity,
			moduleMaxBorrow,
		).Sub(meTokenSupply.Amount),
		sdk.ZeroInt(),
	), nil
}
