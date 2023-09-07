package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// TODO: unit-test this file

// AccountPosition must be created by NewAccountPosition for proper initialization.
// Contains an account's borrowed and collateral values, arranged into special asset
// pairs and regular assets. Each list will always be sorted by collateral weight.
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

// NewAccountPosition creates and sorts an account position based on token settings,
// special asset pairs, and the collateral and borrowed value of each token in an account.
// Once this structure is created, borrow limit calculations can be performed without
// keeper or context.
func NewAccountPosition(
	tokens []Token,
	pairs []SpecialAssetPair,
	unsortedCollateralValue,
	unsortedBorrowValue sdk.DecCoins,
	isLiquidation bool,
	minimumBorrowFactor sdk.Dec,
) (AccountPosition, error) {
	position := AccountPosition{
		specialPairs:        WeightedSpecialPairs{},
		normalPairs:         WeightedNormalPairs{},
		unpairedCollateral:  WeightedDecCoins{},
		unpairedBorrows:     WeightedDecCoins{},
		tokens:              map[string]Token{},
		collateralValue:     sdk.ZeroDec(),
		borrowedValue:       sdk.ZeroDec(),
		isForLiquidation:    isLiquidation,
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
		if isLiquidation {
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

	for _, cv := range unsortedCollateralValue {
		// track total collateral value
		position.collateralValue = position.collateralValue.Add(cv.Amount)
	}
	for _, bv := range unsortedBorrowValue {
		// track total borrowed value
		position.borrowedValue = position.borrowedValue.Add(bv.Amount)
	}

	// match assets into special asset pairs, removing matched value from unsortedBorrowValue and unsortedCollateralValue
	for i, p := range position.specialPairs {
		b := unsortedBorrowValue.AmountOf(p.Borrow.Denom)
		c := unsortedCollateralValue.AmountOf(p.Collateral.Denom)
		if b.IsPositive() && c.IsPositive() {
			// some unmatched assets match the special pair
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

	for _, cv := range unsortedCollateralValue {
		// collect collateral assets which are not part of special pairs
		position.unpairedCollateral = position.unpairedCollateral.Add(WeightedDecCoin{
			Asset:  cv,
			Weight: position.tokenWeight(cv.Denom),
		})
	}
	for _, bv := range unsortedBorrowValue {
		// collect borrowed assets which are not part of special pairs
		position.unpairedBorrows = position.unpairedBorrows.Add(WeightedDecCoin{
			Asset:  bv,
			Weight: position.tokenWeight(bv.Denom),
		})
	}

	// match position's unpaired borrows and collateral into normal asset pairs
	position.sortNormalAssets()

	// always validates the position before returning for safety
	return position, position.Validate()
}

// validates basic properties of a position that should always be true
func (ap *AccountPosition) Validate() error {
	if len(ap.unpairedCollateral) > 0 && len(ap.unpairedBorrows) > 0 {
		return fmt.Errorf("position has both unpaired borrows and unpaired collateral")
	}
	totalCollateral := sdk.ZeroDec()
	totalBorrowed := sdk.ZeroDec()

	for _, sp := range ap.specialPairs {
		totalCollateral = totalCollateral.Add(sp.Collateral.Amount)
		totalBorrowed = totalBorrowed.Add(sp.Borrow.Amount)
	}
	for _, np := range ap.normalPairs {
		totalCollateral = totalCollateral.Add(np.Collateral.Asset.Amount)
		totalBorrowed = totalBorrowed.Add(np.Borrow.Asset.Amount)
	}
	for _, c := range ap.unpairedCollateral {
		totalCollateral = totalCollateral.Add(c.Asset.Amount)
	}
	for _, b := range ap.unpairedBorrows {
		totalBorrowed = totalBorrowed.Add(b.Asset.Amount)
	}

	if !totalCollateral.Equal(ap.collateralValue) {
		return fmt.Errorf("total collateral value mismatch")
	}
	if !totalBorrowed.Equal(ap.borrowedValue) {
		return fmt.Errorf("total borrow value mismatch")
	}
	return nil
}

// sortNormalAssets arranges all of a position's assets, except special pairs, according to
// the usual set of rules (paired from highest to lowest collateral weight)
func (ap *AccountPosition) sortNormalAssets() {
	normalBorrows := WeightedDecCoins{}
	normalCollateral := WeightedDecCoins{}
	// collect position's normal pairs and unpaired assets, ignoring special pairs
	for _, cv := range ap.unpairedCollateral {
		normalCollateral = normalCollateral.Add(cv)
	}
	for _, bv := range ap.unpairedBorrows {
		normalBorrows = normalBorrows.Add(bv)
	}
	for _, np := range ap.normalPairs {
		normalBorrows = normalBorrows.Add(np.Borrow)
		normalCollateral = normalCollateral.Add(np.Collateral)
	}
	// clear the position's normal pairs and unpaired assets
	ap.normalPairs = WeightedNormalPairs{}
	ap.unpairedBorrows = WeightedDecCoins{}
	ap.unpairedCollateral = WeightedDecCoins{}
	// match assets into normal asset pairs, removing matched value from normalBorrows and normalCollateral
	var i, j int
	for i < len(normalCollateral) && j < len(normalBorrows) {
		cDenom := normalCollateral[i].Asset.Denom
		bDenom := normalBorrows[j].Asset.Denom
		c := normalCollateral[i].Asset.Amount
		b := normalBorrows[j].Asset.Amount
		w := sdk.MinDec(
			// for normal asset pairs, both tokens limit the collateral weight of the pair
			ap.tokenWeight(cDenom),
			sdk.MaxDec(ap.tokenWeight(bDenom), ap.minimumBorrowFactor),
		)
		// match collateral and borrow at indexes i and j, exhausting at least one of them
		pairBorrowLimit := c.Mul(w)
		var bCoin, cCoin sdk.DecCoin
		if pairBorrowLimit.LTE(b) {
			// all of the collateral is used (note: this case includes collateral with zero weight)
			cCoin = sdk.NewDecCoinFromDec(cDenom, c)
			// only some of the borrow, equal to collateral value * collateral weight is covered
			bCoin = sdk.NewDecCoinFromDec(bDenom, pairBorrowLimit)
			// next collateral
			i++
		} else {
			// some collateral, equal to borrow value / collateral weight, is used
			cCoin = sdk.NewDecCoinFromDec(cDenom, b.Quo(w))
			// all of the borrow is covered by collateral in this pair
			bCoin = sdk.NewDecCoinFromDec(bDenom, b)
			// next borrow
			j++
		}

		// skip zero positions.
		if cCoin.IsPositive() || bCoin.IsPositive() {
			// subtract newly paired assets from unsorted assets
			normalBorrows = normalBorrows.Sub(bCoin)
			normalCollateral = normalCollateral.Sub(cCoin)
			// create a normal asset pair and add it to the account position
			ap.normalPairs = ap.normalPairs.Add(WeightedNormalPair{
				Collateral: WeightedDecCoin{
					Asset:  cCoin,
					Weight: ap.tokenWeight(cDenom),
				},
				Borrow: WeightedDecCoin{
					Asset:  bCoin,
					Weight: ap.tokenWeight(bDenom),
				},
			})
		}
	}

	// any remaining collateral could not be paired (so borrower is under limit)
	for _, cv := range normalCollateral {
		if cv.Asset.IsPositive() {
			// sort collateral by collateral weight (or liquidation threshold) using Add function
			ap.unpairedCollateral = ap.unpairedCollateral.Add(cv)
		}
	}

	// any remaining borrows could not be paired (so borrower is over limit)
	for _, bv := range normalBorrows {
		if bv.Asset.IsPositive() {
			// sort borrows by collateral weight (or liquidation threshold) using Add function
			ap.unpairedBorrows = ap.unpairedBorrows.Add(bv)
		}
	}
}

// BorrowedValue returns an account's total USD value borrowed
func (ap *AccountPosition) BorrowedValue() sdk.Dec {
	return ap.borrowedValue
}

// CollateralValue returns an account's collateral's total USD value
func (ap *AccountPosition) CollateralValue() sdk.Dec {
	return ap.collateralValue
}

// IsHealthy returns true if an account's borrowed value is less than its borrow limit.
func (ap *AccountPosition) IsHealthy() bool {
	return ap.Limit().GTE(ap.borrowedValue)
}

// HasCollateral returns true if a position has nonzero collateral value
// of a given token in special pairs, normal pairs, or unpaired collateral.
func (ap *AccountPosition) HasCollateral(denom string) bool {
	for _, sp := range ap.specialPairs {
		if sp.Collateral.Denom == denom && sp.Collateral.IsPositive() {
			return true
		}
	}
	for _, np := range ap.normalPairs {
		if np.Collateral.Asset.Denom == denom && np.Collateral.Asset.IsPositive() {
			return true
		}
	}
	for _, c := range ap.unpairedCollateral {
		if c.Asset.Denom == denom && c.Asset.IsPositive() {
			return true
		}
	}
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

// hasToken returns true if a token is registered
func (ap *AccountPosition) hasToken(denom string) bool {
	_, ok := ap.tokens[denom]
	return ok
}

// Limit computes the borrow limit or liquidation threshold of a position, depending on position.isForLiquidation.
// The result may be less or more than its borrowed value.
func (ap *AccountPosition) Limit() sdk.Dec {
	// An initialized account position already has special asset pairs matched up, so
	// this function only needs to count unpaired borrows or collateral and compute
	// the distance from current borrowed value to the limit.

	// if any borrows remain after matching, user is over borrow limit (or LT) by remaining value
	remainingBorrowValue := ap.unpairedBorrows.Total("")
	if remainingBorrowValue.IsPositive() {
		return ap.borrowedValue.Sub(remainingBorrowValue)
	}

	// if no borrows remain after matching, user may have additional borrow limit (or LT) available
	limit := ap.borrowedValue
	for _, c := range ap.unpairedCollateral {
		// the borrow limit calculation assumes no additional limitations based on borrow factor
		// this is accurate if the asset to be borrowed has a higher collateral weight than all
		// remaining collateral. otherwise, it overestimates. MaxBorrow is more precise in those cases.
		limit = limit.Add(c.Asset.Amount.Mul(c.Weight))
	}
	return limit
}

// MaxBorrow computes the maximum USD value of a given base token denom a position can borrow
// without exceeding its borrow limit. Mutates the AccountPosition to show the new borrow amount,
// meaning subsequent calls to MaxBorrow will return zero.
func (ap *AccountPosition) MaxBorrow(denom string) (sdk.Dec, error) {
	// An initialized account position already has special asset pairs matched up, but these pairs
	// could change due to new borrow.
	//
	// Effects of new borrow:
	// - borrow first added to applicable special pairs
	//		- can absorb collateral from lower weight special pairs
	//			- each displaced borrow asset which lost its paired collateral must be placed again
	//				- displaced borrow asset placed in special pairs, if available
	//					- can displace additional borrowed assets from pairs (etc, chain reaction)
	//						- if reached borrow limit, stop here
	//				- displaced borrow asset placed in unpaired assets
	//					- can displace unpaired borrowed assets
	//						- if reached borrow limit, stop here
	//		- can absorb collateral from normal pairs
	//				- only up to the amount allowed by withdrawNormalCollateral
	//						- if reached borrow limit, stop here
	// - borrow then added to normal and unpaired assets
	//		- can displace borrow assets of lower weight
	//			- borrow until borrow limit is reached
	//
	// To calculate max borrow exactly, this procedure would need to be executed in order until the
	// position runs out of unpaired collateral.
	borrowed := sdk.ZeroDec()
	// for each special asset pair which allows this borrowed token, starting with highest weight
	for i, sp := range ap.specialPairs {
		// TODO: naive MaxBorrow currently does not attempt to displace collateral from
		// lower weight special pairs with the same borrow asset, so it will
		// underestimate MaxBorrow when such pairs exist.
		if sp.Borrow.Denom == denom {
			// free up collateral by withdrawing it from the position
			c, err := ap.withdrawNormalCollateral(sp.Collateral.Denom)
			if err != nil {
				return sdk.ZeroDec(), err
			}
			// calculate additional borrow amount
			b := c.Mul(sp.SpecialWeight)
			// add the freed collateral back to the position in a special pair
			ap.specialPairs[i].Collateral.Amount = sp.Collateral.Amount.Add(c)
			ap.collateralValue = ap.collateralValue.Add(c)
			// also add the borrowed assets
			ap.specialPairs[i].Borrow.Amount = sp.Borrow.Amount.Add(b)
			ap.borrowedValue = ap.borrowedValue.Add(b)
			borrowed = borrowed.Add(b)
		}
	}
	// rearrange normal assets such that borrows which are lower weight than the
	// requested denom are pushed below unpaired collateral, and any collateral
	// which can be used to borrow the input denom becomes the new unpaired
	err := ap.displaceBorrowsAfterBorrowDenom(denom)
	if err != nil {
		return sdk.ZeroDec(), err
	}
	// borrow the maximum possible amount of input denom against all remaining unpaired collateral
	borrowed = borrowed.Add(ap.fillOrdinaryCollateral(denom, ap.collateralValue))
	return borrowed, nil
}

// MaxWithdraw computes the maximum USD value of a given base token denom a position can withdraw
// from its collateral.
func (ap *AccountPosition) MaxWithdraw(denom string) (sdk.Dec, error) {
	// An initialized account position already has special asset pairs matched up, but these pairs
	// could change due to withdrawal.
	//
	// Effects of collateral withdrawal:
	// - collateral first taken from unpaired assets
	// - then taken from paired assets, lowest weight pairs first
	//		- each displaced borrow asset which lost its paired collateral must be placed again
	//			- displaced borrow asset placed in special pairs, if available
	//				- can displace additional borrowed assets from pairs (etc, chain reaction)
	//					- if reached borrow limit, stop here
	//			- displaced borrow asset placed in normal assets
	//				- can displace normal borrowed assets
	//					- if reached borrow limit, stop here
	// - if borrow limit still not reached, user is free to withdraw maxmimum
	//
	// To calculate max withdraw exactly, this procedure would need to be executed in order until the
	// position runs out of unpaired collateral or all collateral of the input denom is withdrawn.
	withdrawn, err := ap.withdrawNormalCollateral(denom)
	if err != nil {
		return sdk.ZeroDec(), err
	}
	for i := len(ap.specialPairs) - 1; i >= 0; i-- {
		// for all special pairs, starting with the lowest weight
		if ap.specialPairs[i].Collateral.Denom == denom {
			// attempt to withdraw collateral from special pairs
			c, err := ap.withdrawFromSpecialPair(i)
			if err != nil {
				return sdk.ZeroDec(), err
			}
			withdrawn = withdrawn.Add(c)
		}
	}
	return withdrawn, nil
}

/*

Concept: Asset Priority Ladder. High priority assets at the top. Chain reactions propagate downward.

 	Collateral		Borrow
	-------------------------------
	A(sp)			B(sp)
	A(sp)			C(sp)
	B(sp)			A(sp)	<--- special asset pairs will always be present, having zero amount if unused
	A(sp)			A(sp)
	C(sp)			B(sp)
	C(sp)			D(sp)
	-------------------------------
	A				A
	B				A
	B				B		<--- these matchings of ordinary assets only exist when nonzero
	C				B
	C				C
	C				D
	C				-		<--- there is unpaired collateral initially, if MaxBorrow or MaxWithdraw is nonzero
	D				-

When computing max borrow of asset B, we need to find a borrow amount such that all remaining unused collateral
is consumed. This is complicated by the fact that the borrowed B will first occupy in any special pairs which
allow borrowed B, thus pulling the opposing collateral asset in the affected pairs from either another special
pair, or some normal pair, or unpaired collateral. The first two options displace whatever borrowed asset was
borrowed by that collateral, which has the same effects as described from the start of this paragraph. If not
occupying a special pair, the borrow B should be inserted into the table or normal pairs by matching all
lower weighted borrows (C and D) with the lowest weight normal collateral assets first, thus freeing up other
collateral in the middle of the normal asset table which may be occupied by B. The amount of B placed throughout
all of this is the MaxBorrow.

Some kind of helper function which manipulates the pairs and assets from the AccountPosition struct is needed,
likely one that can call itself recursively for the chain reactions caused by displacing special pairs. It should
be able to detect when it has finally filled all collateral with borrows and then return the total amount of borrowing
achieved.

Probably these functions should progressively limit their own scope if they're going to be recursive. Visually, any
operation which displaces a borrow or collateral can only affect assets below its row. Therefore the scope narrows
by gradually omitting rows of special pairs until only normal pairs and unpaired collateral are left, which can
be handled without recursion.

*/
