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

	borrowers, err := []types.BorrowerSummary{}, error(nil)
	ctx := sdk.UnwrapSDKContext(goCtx)

	// The "all" symbol denom is converted to empty
	if strings.ToLower(req.Symbol) == "all" {
		req.Symbol = ""
	}

	switch strings.ToLower(req.Flavor) {
	case "borrowed":
		borrowers, err = q.Keeper.GetSortedBorrowers(ctx, req.Value)
	default:
		status.Error(codes.InvalidArgument, "unknown inspector flavor: "+req.Flavor)
	}
	if err != nil {
		return nil, err
	}

	return &types.QueryInspectResponse{Borrowers: borrowers}, nil
}

// Bsums will implement sort.Sort
type Bsums []*types.BorrowerSummary

func (s Bsums) Len() int      { return len(s) }
func (s Bsums) Swap(i, j int) { s[i], s[j] = s[j], s[i] }

// ByHealth implements sort.Interface by providing Less and using the Len and
// Swap methods of the embedded bsums value.
type ByHealth struct{ Bsums }

func (s ByHealth) Less(i, j int) bool {
	// health is > 100% when a borrower cannot be liquidated, and < 100% otherwise
	hi := s.Bsums[i].LiquidationThreshold.Quo(s.Bsums[i].BorrowedValue)
	hj := s.Bsums[j].LiquidationThreshold.Quo(s.Bsums[j].BorrowedValue)
	return hi.LT(hj)
}

// ByValue implements sort.Interface by providing Less and using the Len and
// Swap methods of the embedded bsums value.
type ByValue struct{ Bsums }

func (s ByValue) Less(i, j int) bool { return s.Bsums[i].BorrowedValue.LTE(s.Bsums[j].BorrowedValue) }

// GetSortedBorrowers returns a list of borrower addresses and their account summaries sorted by
// borrowed value (descending) and filtered below a minimum borrowed value.
func (k Keeper) GetSortedBorrowers(ctx sdk.Context, minValue sdk.Dec) ([]types.BorrowerSummary, error) {
	prefix := types.KeyPrefixAdjustedBorrow

	// which addresses have already been checked
	borrowers := Bsums{}
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

		// ignore borrowers smaller than the cutoff
		if borrowedValue.LT(minValue) {
			return nil
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
		}
		borrowers = append(borrowers, &summary)

		return nil
	}

	// collect all borrower summaries (unsorted)
	if err := k.iterate(ctx, prefix, iterator); err != nil {
		return nil, err
	}

	// sorts the borrowers
	sort.Sort(ByValue{borrowers})
	// sort.Sort(ByHealth{borrowers})
	sortedBorrowers := []types.BorrowerSummary{}
	for _, b := range borrowers {
		sortedBorrowers = append(sortedBorrowers, *b)
	}

	return sortedBorrowers, nil
}
