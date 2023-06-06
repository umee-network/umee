package keeper

import (
	"github.com/umee-network/umee/v5/x/leverage/types"
)

// byCustom implements sort.Interface by providing Less and using the Len and
// Swap methods of the embedded bsums value.
type byCustom struct {
	bs   []*types.BorrowerSummary
	less inspectorSort
}

func (s byCustom) Len() int           { return len(s.bs) }
func (s byCustom) Swap(i, j int)      { s.bs[i], s.bs[j] = s.bs[j], s.bs[i] }
func (s byCustom) Less(i, j int) bool { return s.less(s.bs[i], s.bs[j]) }

// inspectorSort defines a Less function for sorting inspected borrower summaries,
// which must return true if a should come before b using custom logic for sorts.
type inspectorSort func(a, b *types.BorrowerSummary) bool

func moreLTV() inspectorSort {
	return func(a, b *types.BorrowerSummary) bool {
		ha := a.BorrowedValue.Quo(a.CollateralValue)
		hb := b.BorrowedValue.Quo(b.CollateralValue)
		return ha.GT(hb)
	}
}

func moreDanger() inspectorSort {
	return func(a, b *types.BorrowerSummary) bool {
		ha := a.LiquidationThreshold.Quo(a.BorrowedValue)
		hb := b.LiquidationThreshold.Quo(b.BorrowedValue)
		return ha.LT(hb)
	}
}

func moreBorrowed(specific bool) inspectorSort {
	if specific {
		return func(a, b *types.BorrowerSummary) bool {
			return a.SpecificBorrowValue.GTE(b.SpecificBorrowValue)
		}
	}
	return func(a, b *types.BorrowerSummary) bool {
		return a.BorrowedValue.GTE(b.BorrowedValue)
	}
}

func moreCollateral(specific bool) inspectorSort {
	if specific {
		return func(a, b *types.BorrowerSummary) bool {
			return a.SpecificCollateralValue.GTE(b.SpecificCollateralValue)
		}
	}
	return func(a, b *types.BorrowerSummary) bool {
		return a.CollateralValue.GTE(b.CollateralValue)
	}
}
