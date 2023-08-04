package types

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
)

// minimum borrow factor is the minimum collateral weight and minimum liquidation threshold
// allowed when a borrowed token is limiting the efficiency of a pair of assets.
var minimumBorrowFactor = sdk.MustNewDecFromStr("0.5")

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
	// surplus collateral after all borrows have been matched
	surplusCollateral WeightedDecCoins
	// surplus borrows if collateral was inadequate
	surplusBorrows WeightedDecCoins
	// caches retrieved token settings
	tokens map[string]Token
	// total collateral value (using PriceModeLow and interpreting unknown prices as zero)
	collateralValue sdk.Dec
	// total borrowed value (using PriceModeHigh and requiring all prices known)
	borrowedValue sdk.Dec
	// isForLiquidation tracks whether the position was built using collateral weight
	// or liquidation threshold
	isForLiquidation bool
}

func (ap *AccountPosition) String() string {
	s := "special:\n"
	for _, wsp := range ap.specialPairs {
		s += fmt.Sprintf("  %s, %s, %s\n", wsp.Collateral, wsp.Borrow, wsp.SpecialWeight)
	}
	s += "normal:\n"
	for _, wnp := range ap.normalPairs {
		s += fmt.Sprintf("  %s, %s\n", wnp.Collateral, wnp.Borrow)
	}
	for _, sc := range ap.surplusCollateral {
		s += fmt.Sprintf("  %s, -\n", sc)
	}
	for _, sb := range ap.surplusBorrows {
		s += fmt.Sprintf("  - , %s\n", sb)
	}
	return s
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
) AccountPosition {
	position := AccountPosition{
		specialPairs:      WeightedSpecialPairs{},
		normalPairs:       WeightedNormalPairs{},
		surplusCollateral: WeightedDecCoins{},
		surplusBorrows:    WeightedDecCoins{},
		tokens:            map[string]Token{},
		collateralValue:   sdk.ZeroDec(),
		borrowedValue:     sdk.ZeroDec(),
		isForLiquidation:  isLiquidation,
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

	// !! TODO: sort.SliceStable collateralValueToSort and borrowValueToSort
	// so normal pairs aren't disrupted

	for _, cv := range unsortedCollateralValue {
		// track total collateral value
		position.collateralValue = position.collateralValue.Add(cv.Amount)
	}
	for _, bv := range unsortedBorrowValue {
		// track total borrowed value
		position.borrowedValue = position.borrowedValue.Add(bv.Amount)
	}

	// match assets into special asset pairs, removing matched value from borrowedValueToSort and collateralValueToSort
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

	sortedCollateralValue := WeightedDecCoins{}
	for _, cv := range unsortedCollateralValue {
		// sort collateral assets which are not part of special pairs
		sortedCollateralValue = sortedCollateralValue.Add(WeightedDecCoin{
			Asset:  cv,
			Weight: position.tokenWeight(cv.Denom),
		})
	}
	sortedBorrowValue := WeightedDecCoins{}
	for _, bv := range unsortedBorrowValue {
		// sort borrowed assets which are not part of special pairs
		sortedBorrowValue = sortedBorrowValue.Add(WeightedDecCoin{
			Asset:  bv,
			Weight: position.tokenWeight(bv.Denom),
		})
	}

	// match assets into normal asset pairs, removing matched value from borrowedValueToSort and collateralValueToSort
	var i, j int
	for i < len(sortedCollateralValue) && j < len(sortedBorrowValue) {
		cDenom := sortedCollateralValue[i].Asset.Denom
		bDenom := sortedBorrowValue[j].Asset.Denom
		c := sortedCollateralValue[i].Asset.Amount
		b := sortedBorrowValue[j].Asset.Amount
		w := sdk.MinDec(
			// for normal asset pairs, both tokens limit the collateral weight of the pair
			position.tokenWeight(cDenom),
			sdk.MaxDec(position.tokenWeight(bDenom), minimumBorrowFactor),
		)
		// match collateral and borrow at indexes i and j, exhausting at least one of them
		pairBorrowLimit := c.Mul(w)
		var bCoin, cCoin sdk.DecCoin
		if pairBorrowLimit.LTE(b) {
			// all of the collateral is used (note: this case includes collateral with zero weight)
			cCoin = sdk.NewDecCoinFromDec(cDenom, c)
			// only some of the borrow, equal to collateral value * collateral weight is covered
			bCoin = sdk.NewDecCoinFromDec(bDenom, c.Mul(w))
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
			sortedBorrowValue = sortedBorrowValue.Sub(bCoin)
			sortedCollateralValue = sortedCollateralValue.Sub(cCoin)
			// create a normal asset pair and add it to the account position
			position.normalPairs = append(position.normalPairs, WeightedNormalPair{
				Collateral: WeightedDecCoin{
					Asset:  cCoin,
					Weight: position.tokenWeight(cDenom),
				},
				Borrow: WeightedDecCoin{
					Asset:  bCoin,
					Weight: position.tokenWeight(bDenom),
				},
			})
		}
	}

	// any remaining collateral could not be paired (so borrower is under limit)
	for _, cv := range sortedCollateralValue {
		if cv.Asset.IsPositive() {
			// sort collateral by collateral weight (or liquidation threshold) using Add function
			position.surplusCollateral = position.surplusCollateral.Add(cv)
		}
	}

	// any remaining borrows could not be paired (so borrower is over limit)
	for _, bv := range sortedBorrowValue {
		if bv.Asset.IsPositive() {
			// sort borrows by collateral weight (or liquidation threshold) using Add function
			position.surplusBorrows = position.surplusBorrows.Add(bv)
		}
	}

	return position
}

// BorrowedValue returns an account's total USD value borrowed
func (ap *AccountPosition) BorrowedValue() sdk.Dec {
	return ap.borrowedValue
}

// CollateralValue returns an account's collateral's total USD value
func (ap *AccountPosition) CollateralValue() sdk.Dec {
	return ap.collateralValue
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

// Limit computes the borrow limit or liquidation threshold of a position, depending on position.isForLiquidation.
// The result may be less or more than its borrowed value.
func (ap *AccountPosition) Limit() sdk.Dec {
	// An initialized account position already has special asset pairs matched up, so
	// this function only needs to count surplus borrows or collateral and compute
	// the distance from current borrowed value to the limit.

	// if any borrows remain after matching, user is over borrow limit (or LT) by remaining value
	remainingBorrowValue := ap.surplusBorrows.Total()
	if remainingBorrowValue.IsPositive() {
		return ap.borrowedValue.Sub(remainingBorrowValue)
	}

	// if no borrows remain after matching, user may have additional borrow limit (or LT) available
	limit := ap.borrowedValue
	for _, c := range ap.surplusCollateral {
		// the borrow limit calculation assumes no additional limitations based on borrow factor
		// this is accurate if the asset to be borrowed has a higher collateral weight than all
		// remaining collateral. otherwise, it overestimates. MaxBorrow is more precise in those cases.
		limit = limit.Add(c.Asset.Amount.Mul(c.Weight))
	}
	return limit
}

// MaxBorrow computes the maximum USD value of a given base token denom a position can borrow
// without exceeding its borrow limit.
func (ap *AccountPosition) MaxBorrow(denom string) sdk.Dec {
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
	// - borrow then added to unpaired assets
	//		- can displace borrow assets of lower weight
	//			- borrow until borrow limit is reached
	//
	// To calculate maximum new borrow (reverse procedure)
	// - collect unpaired collateral value
	//		- if none, no new borrows are possible
	// - ...
	return sdk.ZeroDec()
}

// MaxWithdraw computes the maximum USD value of a given base token denom a position can withdraw
// from its collateral.
func (ap *AccountPosition) MaxWithdraw(denom string) sdk.Dec {
	// An initialized account position already has special asset pairs matched up, but these pairs
	// could change due to withdrawal.
	//
	// Effects of collateral withdrawal:
	// - collateral first taken from unpaired assets
	//		- can displace borrow assets which were being collateralized
	//			- if reached borrow limit, stop here
	// - then taken from paired assets, lowest weight pairs first
	//		- each displaced borrow asset which lost its paired collateral must be placed again
	//			- displaced borrow asset placed in special pairs, if available
	//				- can displace additional borrowed assets from pairs (etc, chain reaction)
	//					- if reached borrow limit, stop here
	//			- displaced borrow asset placed in unpaired assets
	//				- can displace unpaired borrowed assets
	//					- if reached borrow limit, stop here
	// - if borrow limit still not reached, user is free to withdraw maxmimum
	return sdk.ZeroDec()
}

// TODO: bump to the bottom, or top, when computing max borrow
// TODO: similar when computing max withdraw
// TODO: isolate special pairs and bump

/*

Possible approach: Asset Priority Ladder

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
	C				-		<--- there is leftover collateral initially, if MaxBorrow or MaxWithdraw is nonzero
	D				-

When computing max borrow of asset B, we need to find a borrow amount such that all remaining unused collateral
is consumed. This is complicated by the fact that the borrowed B will first occupy in any special pairs which
allow borrowed B, thus pulling the opposing collateral asset in the affected pairs from either another special
pair, or some used collateral, or unused collateral. The first two options displace whatever borrowed asset was
borrowed by that collateral, which has the same effects as described from the start of this paragraph. If not
occupying a special pair, the borrow B should be inserted into the table or ordinary assets by matching all
lower weighted borrows (C and D) with the lowest weight leftover collateral assets first, thus freeing up other
collateral in the middle of the ordinary asset table which may be occupied by B. The amount of B placed throughout
all of this is the MaxBorrow.

Some kind of helper function which manipulates the pairs and assets from the AccountPosition struct is needed,
likely one that can call itself recursively for these chain reactions. It should be able to detect when it has
finally filled all empty borrow slots and then return the total amount of borrowing achieved.

Probably these functions should progressively limit their own scope if they're going to be recursive. Visually, any
operation which displaces a borrow or collateral can only affect assets below its row (or maybe above in the case
of having the same weight?)
*/
