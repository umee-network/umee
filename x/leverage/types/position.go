package types

import (
	"fmt"

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
	// account's collateralValue USD value of each token, including that which is in special pairs above
	collateralValue sdk.DecCoins
	// account's borrowedValue USD value of each token, including that which is in special pairs above
	borrowedValue sdk.DecCoins
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
	unsortedCollateralValue, unsortedBorrowValue sdk.DecCoins,
	forLiquidation bool,
	minimumBorrowFactor sdk.Dec,
) (AccountPosition, error) {
	position := AccountPosition{
		specialPairs:        WeightedSpecialPairs{},
		collateralValue:     sdk.DecCoins{},
		borrowedValue:       sdk.DecCoins{},
		isForLiquidation:    forLiquidation,
		tokens:              map[string]Token{},
		minimumBorrowFactor: minimumBorrowFactor,
	}

	// cache all registered tokens
	for _, t := range tokens {
		position.tokens[t.BaseDenom] = t
	}

	// cache all potentially relevant special asset pairs, and sort them by collateral weight (or liquidation threshold).
	// Initialize their amounts, which will eventually store matching asset value, to zero.
	for _, sp := range pairs {
		weight := sp.CollateralWeight
		if forLiquidation {
			weight = sp.LiquidationThreshold
		}
		if weight.LTE(
			// pairs may not reduce collateral weight or liquidation threshold
			// below what the tokens would produce without the special pair.
			sdk.MinDec(
				position.tokenWeight(sp.Collateral),
				sdk.MaxDec(position.tokenWeight(sp.Borrow), minimumBorrowFactor),
			),
		) || weight.IsZero() {
			// Such pairs as well as those with zero weight are omitted from the
			// position entirely.
			continue
		}
		wp := WeightedSpecialPair{
			Collateral:    sdk.NewDecCoinFromDec(sp.Collateral, sdk.ZeroDec()),
			Borrow:        sdk.NewDecCoinFromDec(sp.Borrow, sdk.ZeroDec()),
			SpecialWeight: weight,
		}
		// sorting is performed by Add function
		position.specialPairs = position.specialPairs.Add(wp)
	}

	// Copy borrowed and collateral value before special pairs are taken into account
	for _, cv := range unsortedCollateralValue {
		position.collateralValue = position.collateralValue.Add(cv)
	}
	for _, bv := range unsortedBorrowValue {
		position.borrowedValue = position.borrowedValue.Add(bv)
	}

	// match assets into special asset pairs, removing matched value from unsortedBorrowValue and unsortedCollateralValue
	// but not changing position.borrowedValue and position.collateralValue
	for i, p := range position.specialPairs {
		b := unsortedBorrowValue.AmountOf(p.Borrow.Denom)
		c := unsortedCollateralValue.AmountOf(p.Collateral.Denom)
		if b.IsPositive() && c.IsPositive() {
			// some unpaired assets match the special pair
			bCoin := sdk.NewDecCoinFromDec(p.Borrow.Denom, sdk.ZeroDec())
			cCoin := sdk.NewDecCoinFromDec(p.Collateral.Denom, sdk.ZeroDec())
			pairBorrowLimit := c.Mul(p.SpecialWeight)
			if pairBorrowLimit.GTE(b) {
				// all of the borrow is covered by collateral in this pair
				bCoin = sdk.NewDecCoinFromDec(bCoin.Denom, b)
				// some collateral, equal to borrow value / collateral weight, is used
				cCoin = sdk.NewDecCoinFromDec(cCoin.Denom, b.Quo(p.SpecialWeight))
			} else {
				// only some of the borrow, equal to collateral value * collateal weight is covered
				bCoin = sdk.NewDecCoinFromDec(bCoin.Denom, c.Mul(p.SpecialWeight))
				// all of the collateral is used
				cCoin = sdk.NewDecCoinFromDec(cCoin.Denom, c)
			}
			// subtract newly paired assets from unsorted assets
			unsortedBorrowValue = unsortedBorrowValue.Sub(sdk.NewDecCoins(bCoin))
			unsortedCollateralValue = unsortedCollateralValue.Sub(sdk.NewDecCoins(cCoin))
			// add newly paired assets to the appropriate special pair
			position.specialPairs[i].Borrow = bCoin
			position.specialPairs[i].Collateral = cCoin
		}
	}

	// always validates the position before returning for safety
	return position, position.Validate()
}

// validates basic properties of a position that should always be true
func (ap *AccountPosition) Validate() error {
	if err := ap.borrowedValue.Validate(); err != nil {
		return err
	}
	if err := ap.collateralValue.Validate(); err != nil {
		return err
	}
	for _, t := range ap.tokens {
		if err := t.Validate(); err != nil {
			return err
		}
	}
	if ap.minimumBorrowFactor.IsNil() || !ap.minimumBorrowFactor.IsPositive() {
		return fmt.Errorf("invalid minimum borrow factor: %s", ap.minimumBorrowFactor)
	}

	pairedBorrows := sdk.DecCoins{}
	pairedCollateral := sdk.DecCoins{}
	for _, sp := range ap.specialPairs {
		if err := sp.Validate(); err != nil {
			return err
		}
		pairedBorrows = pairedBorrows.Add(sp.Borrow)
		pairedCollateral = pairedCollateral.Add(sp.Collateral)
	}

	if _, neg := ap.borrowedValue.SafeSub(pairedBorrows); neg {
		return fmt.Errorf("special borrows exceeded total")
	}
	if _, neg := ap.collateralValue.SafeSub(pairedCollateral); neg {
		return fmt.Errorf("special collateral exceeded total")
	}

	return nil
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

// tokenWeight gets a token's collateral weight or liquidation threshold if it is registered, else zero
func (ap *AccountPosition) tokenWeight(denom string) sdk.Dec {
	if t, ok := ap.tokens[denom]; ok {
		if ap.isForLiquidation {
			return t.LiquidationThreshold
		}
		return t.CollateralWeight
	}
	return sdk.ZeroDec()
}
