package types

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// AccountPosition must be created by NewAccountPosition for proper initialization.
// Contains an account's borrowed and collateral values, arranged into special asset
// pairs and regular assets. Special pairs will always be sorted by weight.
// Also caches some relevant values that will be reused in computation, like token
// settings and the borrower's total borrowed value and collateral value.
type NuAccountPosition struct {
	// all special asset pairs which apply to the account. Specifically, this should
	// cache any special pairs which match the account's collateral, even if they do
	// not match one of its borrows. This is the widest set of information that could
	// be needed when calculating maxWithdraw of an already collateralized asset, or
	// maxBorrow of any asset (even if not initially present in the position).
	// A pair that does not currently apply to the account (but could do so if a borrow
	// were added or existing pairs were rearranged) will have zero USD value
	// but will still be initialized with the proper weights, and sorted in order.
	specialPairs WeightedSpecialPairs
	// collateral and borrowed that have matched with each other after special pairs
	// are sorted in order of descending collateral weight (of the collateral primarily).
	// these will all be nonzero after a position is initialized.
	normalPairs WeightedNormalPairs
	// unpaired collateral after all borrows have been matched
	unpairedCollateral WeightedDecCoins
	// unpaired borrows if collateral was inadequate
	unpairedBorrows WeightedDecCoins
	// caches retrieved token settings
	tokens map[string]Token
	// total collateral value (using PriceModeLow and interpreting unknown prices as zero)
	collateralValue sdk.Dec
	// total borrowed value (using PriceModeHigh and requiring all prices known)
	borrowedValue sdk.Dec
	// isForLiquidation tracks whether the position was built using collateral weight
	// or liquidation threshold
	isForLiquidation bool
	// minimum borrow factor is the minimum collateral weight and minimum liquidation threshold
	// allowed when a borrowed token is limiting the efficiency of a pair of assets.
	// TODO: parameterize this in the leverage module
	minimumBorrowFactor sdk.Dec
}
