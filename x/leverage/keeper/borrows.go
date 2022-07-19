package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v2/x/leverage/types"
)

// GetBorrow returns an sdk.Coin representing how much of a given denom a
// borrower currently owes.
func (k Keeper) GetBorrow(ctx sdk.Context, borrowerAddr sdk.AccAddress, denom string) sdk.Coin {
	store := ctx.KVStore(k.storeKey)
	owed := sdk.NewCoin(denom, sdk.ZeroInt())
	key := types.CreateAdjustedBorrowKey(borrowerAddr, denom)

	adjustedAmount := sdk.ZeroDec()
	if bz := store.Get(key); bz != nil {
		err := adjustedAmount.Unmarshal(bz)
		if err != nil {
			panic(err)
		}
	}

	// Apply interest scalar
	owed.Amount = adjustedAmount.Mul(k.getInterestScalar(ctx, denom)).Ceil().TruncateInt()
	return owed
}

// setBorrow sets the amount borrowed by an address in a given denom.
// If the amount is zero, any stored value is cleared.
func (k Keeper) setBorrow(ctx sdk.Context, borrowerAddr sdk.AccAddress, borrow sdk.Coin) error {
	// Apply interest scalar to determine adjusted amount
	newAdjustedAmount := borrow.Amount.ToDec().Quo(k.getInterestScalar(ctx, borrow.Denom))

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

// GetAvailableToBorrow gets the amount available to borrow of a given token.
func (k Keeper) GetAvailableToBorrow(ctx sdk.Context, denom string) sdk.Int {
	// Available for borrow = Module Balance - Reserve Amount
	moduleBalance := k.ModuleBalance(ctx, denom)
	reserveAmount := k.GetReserveAmount(ctx, denom)

	return sdk.MaxInt(moduleBalance.Sub(reserveAmount), sdk.ZeroInt())
}

// SupplyUtilization calculates the current supply utilization of a token denom.
func (k Keeper) SupplyUtilization(ctx sdk.Context, denom string) sdk.Dec {
	// Current utilization is calculated using a hypothetical zero-amount borrow
	return k.supplyUtilizationAfterBorrow(ctx, sdk.NewCoin(denom, sdk.ZeroInt()))
}

// supplyUtilizationAfterBorrow calculates the resulting supply utilization of a token
// denom if a proposed borrow were to execute.
func (k Keeper) supplyUtilizationAfterBorrow(ctx sdk.Context, borrow sdk.Coin) sdk.Dec {
	// Supply utilization is equal to total borrows divided by the token supply
	// (including borrowed tokens yet to be repaid and excluding tokens reserved).
	moduleBalance := k.ModuleBalance(ctx, borrow.Denom)
	reserveAmount := k.GetReserveAmount(ctx, borrow.Denom)
	totalBorrowed := k.GetTotalBorrowed(ctx, borrow.Denom).Amount
	tokenSupply := totalBorrowed.Add(moduleBalance).Sub(reserveAmount)

	// Adjust based on the proposed borrow. Note that token supply is unchanged because
	// module balance would decrease by the same amount total borrows increase.
	totalBorrowed = totalBorrowed.Add(borrow.Amount)

	// This case is impossible to reach in practice
	if tokenSupply.IsNegative() {
		return sdk.MaxSortableDec
	}

	if tokenSupply.IsZero() {
		return sdk.ZeroDec()
	}

	// Utilization ranges 0 to 1 generally, or > 1 when reserves exceed module balance.
	// In this calculation, utilization will exceed 1 only if the proposed borrow is greater
	// than unreserved module balance.
	return totalBorrowed.ToDec().Quo(tokenSupply.ToDec())
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

		// get USD value of base assets
		v, err := k.TokenValue(ctx, baseAsset)
		if err != nil {
			return sdk.ZeroDec(), err
		}

		ts, err := k.GetTokenSettings(ctx, baseAsset.Denom)
		if err != nil {
			return sdk.ZeroDec(), err
		}

		// add each collateral coin's weighted value to borrow limit
		limit = limit.Add(v.Mul(ts.CollateralWeight))
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

		// get USD value of base assets
		v, err := k.TokenValue(ctx, baseAsset)
		if err != nil {
			return sdk.ZeroDec(), err
		}

		ts, err := k.GetTokenSettings(ctx, baseAsset.Denom)
		if err != nil {
			return sdk.ZeroDec(), err
		}

		totalThreshold = totalThreshold.Add(v.Mul(ts.LiquidationThreshold))
	}

	return totalThreshold, nil
}

// setBadDebtAddress sets or deletes an address in a denom's list of addresses with unpaid bad debt.
func (k Keeper) setBadDebtAddress(ctx sdk.Context, addr sdk.AccAddress, denom string, hasDebt bool) error {
	if err := sdk.ValidateDenom(denom); err != nil {
		return err
	}
	if addr.Empty() {
		return types.ErrEmptyAddress
	}

	store := ctx.KVStore(k.storeKey)
	key := types.CreateBadDebtKey(denom, addr)

	if hasDebt {
		store.Set(key, []byte{0x01})
	} else {
		store.Delete(key)
	}
	return nil
}
