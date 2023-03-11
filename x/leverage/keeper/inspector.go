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

	switch strings.ToLower(req.Flavor) {
	case "borrowed":
		filter = withMinBorrowedValue(req.Value)
		sorting = moreBorrowed()
	case "health":
		filter = withMinBorrowedValue(req.Value)
		sorting = lessHealthy()
	default:
		status.Error(codes.InvalidArgument, "unknown inspector flavor: "+req.Flavor)
	}
	borrowers, err := k.filteredSortedBorrowers(ctx,
		filter,
		sorting,
	)
	if err != nil {
		return nil, err
	}

	return &types.QueryInspectResponse{Borrowers: borrowers}, nil
}

// filteredSortedBorrowers returns a list of borrower addresses and their account summaries sorted and filtered
// by selected methods. Sorting is ascending in X if the sort function provided is "less X", and descending in
// X if the function provided is "more X"
func (k Keeper) filteredSortedBorrowers(ctx sdk.Context, filter inspectorFilter, sorting inspectorSort,
) ([]types.BorrowerSummary, error) {
	borrowers := k.unsortedBorrowers(ctx)

	// filters the borrowers
	filteredBorrowers := []*types.BorrowerSummary{}
	for _, bs := range borrowers {
		if filter(bs) {
			filteredBorrowers = append(filteredBorrowers, bs)
		}
	}

	// sorts the borrowers
	sort.Sort(byCustom{
		bs:   borrowers,
		less: sorting,
	})
	sortedBorrowers := []types.BorrowerSummary{}

	// convert from pointers
	for _, b := range borrowers {
		sortedBorrowers = append(sortedBorrowers, *b)
	}

	return sortedBorrowers, nil
}

// unsortedBorrowers returns a list of borrower addresses and their account summaries, without filters or sorting
func (k Keeper) unsortedBorrowers(ctx sdk.Context) []*types.BorrowerSummary {
	prefix := types.KeyPrefixAdjustedBorrow

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
		borrowedValue, err := k.TotalTokenValue(ctx, borrowed, types.PriceModeSpot)
		if err != nil {
			return err
		}

		supplied, err := k.GetAllSupplied(ctx, addr)
		if err != nil {
			return err
		}
		collateral := k.GetBorrowerCollateral(ctx, addr)

		suppliedValue, err := k.TotalTokenValue(ctx, supplied, types.PriceModeSpot)
		if err != nil {
			return err
		}
		collateralValue, err := k.CalculateCollateralValue(ctx, collateral)
		if err != nil {
			return err
		}
		borrowLimit, err := k.CalculateBorrowLimit(ctx, collateral)
		if err != nil {
			return err
		}
		liquidationThreshold, err := k.CalculateLiquidationThreshold(ctx, collateral)
		if err != nil {
			return err
		}

		summary := types.BorrowerSummary{
			Address:              addr.String(),
			SuppliedValue:        suppliedValue,
			CollateralValue:      collateralValue,
			BorrowedValue:        borrowedValue,
			BorrowLimit:          borrowLimit,
			LiquidationThreshold: liquidationThreshold,
			TopBorrowed:          "not implemented",
			TopCollateral:        "not implemented",
		}
		borrowers = append(borrowers, &summary)
		return nil
	}

	// collect all borrower summaries (unsorted)
	_ = k.iterate(ctx, prefix, iterator)
	return borrowers
}
