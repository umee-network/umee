package keeper

import (
	"context"
	"math"
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/umee-network/umee/v4/x/leverage/types"
)

// Otherwise from grpc_query.go
func (q Querier) Inspect(
	goCtx context.Context,
	req *types.QueryInspect,
) (*types.QueryInspectResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	// This query is also disabled by default as a safety measure. Enable with liquidator queries.
	if !q.Keeper.liquidatorQueryEnabled {
		return nil, types.ErrNotLiquidatorNode
	}

	k, ctx := q.Keeper, sdk.UnwrapSDKContext(goCtx)
	var filter inspectorFilter
	var sorting inspectorSort

	// The "all" symbol denom is converted to empty symbol denom
	if strings.EqualFold(req.Symbol, "all") {
		req.Symbol = ""
	}
	if req.Value.IsNil() {
		req.Value = sdk.ZeroDec()
	}
	specific := req.Symbol != ""

	switch strings.ToLower(req.Flavor) {
	case "borrowed":
		filter = withMinBorrowedValue(req.Value, specific)
		sorting = moreBorrowed(specific)
	case "collateral":
		filter = withMinCollateralValue(req.Value, specific)
		sorting = moreBorrowed(specific)
	case "borrowed-by-danger":
		filter = withMinBorrowedValue(req.Value, specific)
		sorting = moreDanger()
	case "danger-by-borrowed":
		filter = withMinDanger(req.Value)
		sorting = moreBorrowed(specific)
	default:
		status.Error(codes.InvalidArgument, "unknown inspector flavor: "+req.Flavor)
	}
	borrowers, err := k.filteredSortedBorrowers(ctx,
		filter,
		sorting,
		req.Symbol,
	)
	if err != nil {
		return nil, err
	}

	return &types.QueryInspectResponse{Borrowers: borrowers}, nil
}

// Otherwise from grpc_query.go
func (q Querier) InspectNeat(
	goCtx context.Context,
	req *types.QueryInspectNeat,
) (*types.QueryInspectNeatResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	// The "all" symbol denom is converted to empty symbol denom
	if strings.EqualFold(req.Symbol, "all") {
		req.Symbol = ""
	}

	req2 := types.QueryInspect{
		Symbol: req.Symbol,
		Value:  req.Value,
		Flavor: req.Flavor,
	}

	resp2, err := q.Inspect(goCtx, &req2)
	if err != nil {
		return nil, err
	}

	borrowers := []types.BorrowerSummaryNeat{}
	for _, b := range resp2.Borrowers {
		if b.SuppliedValue.IsPositive() {
			borrowed := b.BorrowedValue
			if req.Symbol != "" {
				borrowed = b.SpecificBorrowValue
			}
			neat := types.BorrowerSummaryNeat{
				Account:  b.Address,
				Borrowed: neat(borrowed),
				L:        neat(b.BorrowedValue.Quo(b.BorrowLimit)),
				Q:        neat(b.BorrowedValue.Quo(b.LiquidationThreshold)),
				V:        neat(b.BorrowedValue.Quo(b.CollateralValue)),
			}
			borrowers = append(borrowers, neat)
		}
	}

	return &types.QueryInspectNeatResponse{Borrowers: borrowers}, nil
}

// Neat truncates an sdk.Dec to a common-sense precision based on its size and converts it to float.
// This makes a big difference in readability when using the CLI.
func neat(num sdk.Dec) float64 {
	n := num.MustFloat64()
	precision := 3 // Round to thousandths for ratios and small-dollar amounts
	if n > 10 {
		precision = 1 // Round to dime
	}
	if n > 1000 {
		precision = 0 // Round to dollar
	}
	if n > 1_000_000 {
		precision = -3 // Round to thousand
	}
	if n < 0.001 {
		precision = 6 // round to millionths
	}
	if n < 0.000001 {
		return n // maximum precision
	}
	// Truncate the float at a certain precision (can be negative)
	x := math.Pow(10, float64(precision))
	return float64(int(n*x)) / x
}

// filteredSortedBorrowers returns a list of borrower addresses and their account summaries sorted and filtered
// by selected methods. Sorting is ascending in X if the sort function provided is "less X", and descending in
// X if the function provided is "more X"
func (k Keeper) filteredSortedBorrowers(ctx sdk.Context, filter inspectorFilter, sorting inspectorSort, symbol string,
) ([]types.BorrowerSummary, error) {
	borrowers := k.unsortedBorrowers(ctx, symbol)

	// filters the borrowers
	filteredBorrowers := []*types.BorrowerSummary{}
	for _, bs := range borrowers {
		if filter(bs) {
			filteredBorrowers = append(filteredBorrowers, bs)
		}
	}

	// sorts the borrowers
	sort.Sort(byCustom{
		bs:   filteredBorrowers,
		less: sorting,
	})
	sortedBorrowers := []types.BorrowerSummary{}

	// convert from pointers
	for _, b := range filteredBorrowers {
		sortedBorrowers = append(sortedBorrowers, *b)
	}

	return sortedBorrowers, nil
}

// unsortedBorrowers returns a list of borrower addresses and their account summaries, without filters or sorting.
// Also accepts a symbol denom to use for the SpecificBorrowValue and SpecificCollateralValue fields.
func (k Keeper) unsortedBorrowers(ctx sdk.Context, symbol string) []*types.BorrowerSummary {
	prefix := types.KeyPrefixAdjustedBorrow

	denom, uDenom := "", ""
	tokens := k.GetAllRegisteredTokens(ctx)
	for _, t := range tokens {
		if strings.EqualFold(t.SymbolDenom, symbol) {
			denom = t.BaseDenom
			uDenom = types.ToUTokenDenom(denom)
		}
	}

	// which addresses have already been checked
	borrowers := []*types.BorrowerSummary{}
	checkedAddrs := map[string]interface{}{}

	iterator := func(key, _ []byte) error {
		// get borrower address from key
		addr := types.AddressFromKey(key, prefix)

		// if the address is already checked, do not check again
		if _, ok := checkedAddrs[addr.String()]; ok {
			return nil
		}
		checkedAddrs[addr.String()] = struct{}{}

		borrowed := k.GetBorrowerBorrows(ctx, addr)
		borrowedValue, _ := k.TotalTokenValue(ctx, borrowed, types.PriceModeSpot)
		supplied, _ := k.GetAllSupplied(ctx, addr)
		collateral := k.GetBorrowerCollateral(ctx, addr)
		suppliedValue, _ := k.TotalTokenValue(ctx, supplied, types.PriceModeSpot)
		collateralValue, _ := k.CalculateCollateralValue(ctx, collateral)
		borrowLimit, _ := k.CalculateBorrowLimit(ctx, collateral)
		liquidationThreshold, _ := k.CalculateLiquidationThreshold(ctx, collateral)

		specificBorrowedValue := sdk.ZeroDec()
		specificCollateralValue := sdk.ZeroDec()
		if denom != "" {
			specificBorrowed := sdk.NewCoin(denom, borrowed.AmountOf(denom))
			specificCollateral := sdk.NewCoin(uDenom, collateral.AmountOf(uDenom))
			specificBorrowedValue, _ = k.TokenValue(ctx, specificBorrowed, types.PriceModeSpot)
			specificCollateralValue, _ = k.TokenValue(ctx, specificCollateral, types.PriceModeSpot)
		}

		summary := types.BorrowerSummary{
			Address:                 addr.String(),
			SuppliedValue:           suppliedValue,
			CollateralValue:         collateralValue,
			BorrowedValue:           borrowedValue,
			BorrowLimit:             borrowLimit,
			LiquidationThreshold:    liquidationThreshold,
			SpecificCollateralValue: specificBorrowedValue,
			SpecificBorrowValue:     specificCollateralValue,
		}
		borrowers = append(borrowers, &summary)
		return nil
	}

	// collect all borrower summaries (unsorted)
	_ = k.iterate(ctx, prefix, iterator)
	return borrowers
}
