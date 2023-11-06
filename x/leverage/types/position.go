package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AccountPosition must be created by NewAccountPosition for proper initialization.
// Contains an account's borrowed and collateral values, arranged into special asset
// pairs and normal assets. Special pairs will always be sorted by weight.
// Also caches some relevant values that will be reused in computation, like token
// settings and the borrower's total borrowed value and collateral value.
type AccountPosition struct {
	// all special asset pairs which apply to the account. Specifically, this should
	// cache any special pairs which match the account's collateral, even if they do
	// not match one of its borrows. This is the widest set of information that could
	// be needed when calculating maxWithdraw of an already collateralized asset, or
	// maxBorrow of any asset (even if not initially present in the position).
	// A pair that does not currently apply to the account (but could do so if a borrow
	// were added or existing pairs were rearranged) will have zero USD value
	// but will still be initialized with the proper weight, and sorted in order.
	// Special pairs which would reduce an asset's weight are included, but have no effect.
	specialPairs WeightedSpecialPairs
	// account's collateral USD value of each token, excluding that which is in special pairs above
	normalCollateral sdk.DecCoins
	// account's borrowed USD value of each token, excluding that which is in special pairs above
	normalBorrowed sdk.DecCoins
	// tracks whether the position was built using collateral weight or liquidation threshold
	isForLiquidation bool
	// caches retrieved token settings
	tokens map[string]Token
	// caches the minimum collateral weight or minimum liquidation threshold
	minimumBorrowFactor sdk.Dec
}

// NewAccountPosition creates an account position based on a user's borrowed and collateral
// values of each token, and additional information including token settings and special asset pairs.
func NewAccountPosition(
	tokens []Token,
	pairs []SpecialAssetPair,
	collateralValue, borrowedValue sdk.DecCoins,
	forLiquidation bool,
	minimumBorrowFactor sdk.Dec,
) (AccountPosition, error) {
	return AccountPosition{
		isForLiquidation: forLiquidation,
	}, nil
}

// MaxBorrow finds the maximum additional amount of an asset a position can
// borrow without exceeding its borrow limit. Does not interact with
// special asset pairs. Returns the amount of borrows added.
// If the requested token denom did not exist or the borrower was already
// at or over their borrow limit, this is a no-op which returns zero.
// Returns zero if a position was computed with liquidation in mind.
func (ap *AccountPosition) MaxBorrow(denom string) sdk.Dec {
	if ap.isForLiquidation {
		return sdk.ZeroDec()
	}
	return sdk.ZeroDec()
}

// MaxWithdraw finds the maximum additional amount of an asset a position can
// withdraw without exceeding its borrow limit.
// Returns zero if a position was computed with liquidation in mind.
func (ap *AccountPosition) MaxWithdraw(denom string) sdk.Dec {
	if ap.isForLiquidation {
		return sdk.ZeroDec()
	}
	return sdk.ZeroDec()
}

// HasCollateral returns true if a position contains any collateral of a given
// type.
func (ap *AccountPosition) HasCollateral(denom string) bool {
	return false
}

// Limit calculates the borrow limit of an account position
// (or liquidation threshold if ap.isForLiquidation is true).
func (ap *AccountPosition) Limit() sdk.Dec {
	return sdk.ZeroDec()
}

// BorrowedValue() sums the total borrowed value in a position.
func (ap *AccountPosition) BorrowedValue() sdk.Dec {
	return sdk.ZeroDec()
}

// CollateralValue() sums the total collateral value in a position.
func (ap *AccountPosition) CollateralValue() sdk.Dec {
	return sdk.ZeroDec()
}

// IsHealthy() returns true if a position's borrowed value is below
// its borrow limit (or liquidation threshold if ap.isForLiquidation is true).
func (ap *AccountPosition) IsHealthy() bool {
	return false
}
