package keeper

import (
	"context"
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
	if strings.ToLower(req.Symbol) == "all" {
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
	for _, b := range borrowers {
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
		if strings.ToUpper(t.SymbolDenom) == strings.ToUpper(symbol) {
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
