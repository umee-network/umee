package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v4/x/leverage/types"
)

// inspectorFilter defines a function which decides whether to pay attention to a BorrowerSummary
type inspectorFilter func(*types.BorrowerSummary) bool

func withMinBorrowedValue(value sdk.Dec, specific bool) inspectorFilter {
	return func(bs *types.BorrowerSummary) bool {
		if specific {
			return bs.SpecificBorrowValue.GTE(value)
		}
		return bs.BorrowedValue.GTE(value)
	}
}

func withMinCollateralValue(value sdk.Dec, specific bool) inspectorFilter {
	return func(bs *types.BorrowerSummary) bool {
		if specific {
			return bs.SpecificCollateralValue.GTE(value)
		}
		return bs.CollateralValue.GTE(value)
	}
}

func withMinDanger(value sdk.Dec) inspectorFilter {
	return func(bs *types.BorrowerSummary) bool {
		return bs.LiquidationThreshold.IsPositive() && bs.BorrowedValue.Quo(bs.LiquidationThreshold).GTE(value)
	}
}

func withMinLTV(value sdk.Dec) inspectorFilter {
	return func(bs *types.BorrowerSummary) bool {
		return bs.CollateralValue.IsPositive() && bs.BorrowedValue.Quo(bs.CollateralValue).GTE(value)
	}
}

// withZeroes returns borrower summaries with unexpected zero values (knowing that borrower summaries only exist
// for accounts with borrowed tokens)
func withZeroes() inspectorFilter {
	return func(bs *types.BorrowerSummary) bool {
		return bs.CollateralValue.IsZero() || bs.LiquidationThreshold.IsZero() || bs.BorrowedValue.IsZero()
	}
}