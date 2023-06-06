package keeper

import (
	"context"
	"fmt"
	"math"
	"sort"
	"strings"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/umee-network/umee/v5/x/leverage/types"
)

// Separated from grpc_query.go
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
	var filters []inspectorFilter
	var sorting inspectorSort

	// The "all" symbol denom is converted to empty symbol denom
	if strings.EqualFold(req.Symbol, "all") {
		req.Symbol = ""
	}
	modeMin := sdk.MustNewDecFromStr(fmt.Sprintf("%f", req.ModeMin))
	sortMin := sdk.MustNewDecFromStr(fmt.Sprintf("%f", req.SortMin))
	specific := req.Symbol != ""

	switch strings.ToLower(req.Mode) {
	case "borrowed":
		filters = append(filters, withMinBorrowedValue(modeMin, specific))
	case "collateral":
		filters = append(filters, withMinCollateralValue(modeMin, specific))
	case "danger":
		filters = append(filters, withMinDanger(modeMin))
	case "ltv":
		filters = append(filters, withMinLTV(modeMin))
	case "zeroes":
		filters = append(filters, withZeroes())
	default:
		return &types.QueryInspectResponse{}, status.Error(codes.InvalidArgument, "unknown inspector mode: "+req.Mode)
	}
	switch strings.ToLower(req.Sort) {
	case "borrowed":
		sorting = moreBorrowed(specific)
		if sortMin.IsPositive() {
			filters = append(filters, withMinBorrowedValue(sortMin, specific))
		}
	case "collateral":
		sorting = moreCollateral(specific)
		if sortMin.IsPositive() {
			filters = append(filters, withMinCollateralValue(sortMin, specific))
		}
	case "danger":
		sorting = moreDanger()
		if sortMin.IsPositive() {
			filters = append(filters, withMinDanger(sortMin))
		}
	case "ltv":
		sorting = moreLTV()
		if sortMin.IsPositive() {
			filters = append(filters, withMinLTV(sortMin))
		}
	default:
		// if no sort mode is specified, return all borrowers sorted by total collateral value
		sorting = moreCollateral(false)
	}

	borrowers := k.filteredSortedBorrowers(ctx,
		filters,
		sorting,
		req.Symbol,
	)
	return &types.QueryInspectResponse{Borrowers: borrowers}, nil
}

// Separated from grpc_query.go
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
		Symbol:  req.Symbol,
		Mode:    req.Mode,
		Sort:    req.Sort,
		ModeMin: req.ModeMin,
		SortMin: req.SortMin,
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
				LTB:      neat(b.BorrowedValue.Quo(b.BorrowLimit)),
				LTL:      neat(b.BorrowedValue.Quo(b.LiquidationThreshold)),
				LTV:      neat(b.BorrowedValue.Quo(b.CollateralValue)),
			}
			borrowers = append(borrowers, neat)
		}
	}

	return &types.QueryInspectNeatResponse{Borrowers: borrowers}, nil
}

// Separated from grpc_query.go
func (q Querier) RiskData(
	goCtx context.Context,
	req *types.QueryRiskData,
) (*types.QueryRiskDataResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	// This query is also disabled by default as a safety measure. Enable with liquidator queries.
	if !q.Keeper.liquidatorQueryEnabled {
		return nil, types.ErrNotLiquidatorNode
	}
	k, ctx := q.Keeper, sdk.UnwrapSDKContext(goCtx)
	borrowers := k.riskData(ctx)

	return &types.QueryRiskDataResponse{Borrowers: borrowers}, nil
}

// neat truncates an sdk.Dec to a common-sense precision based on its size and converts it to float.
// This makes a big difference in readability when using the CLI.
func neat(num sdk.Dec) float64 {
	n := num.MustFloat64()
	precision := 3 // Round to thousandths if 0.001 <= n <= 10
	if n > 10 {
		precision = 1 // above $10: Round to dime
	}
	if n > 1000 {
		precision = 0 // above $1000: Round to dollar
	}
	if n > 1_000_000 {
		precision = -3 // above $1000000: Round to thousand
	}
	if n < 0.001 {
		precision = 6 // round to millionths
	}
	if n < 0.000001 {
		return n // tiny: maximum precision
	}
	// Truncate the float at a certain precision (can be negative)
	x := math.Pow(10, float64(precision))
	return float64(int(n*x)) / x
}

// filteredSortedBorrowers returns a list of borrower addresses and their account summaries sorted and filtered
// by selected methods. Sorting is ascending in X if the sort function provided is "less X", and descending in
// X if the function provided is "more X"
func (k Keeper) filteredSortedBorrowers(
	ctx sdk.Context, filters []inspectorFilter, sorting inspectorSort, symbol string,
) []types.BorrowerSummary {
	borrowers := k.unsortedBorrowers(ctx, symbol)

	// filters the borrowers
	filteredBorrowers := []*types.BorrowerSummary{}
	for _, bs := range borrowers {
		ok := true
		for _, f := range filters {
			if !f(bs) {
				ok = false
			}
		}
		if ok {
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

	return sortedBorrowers
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
			specificBorrowedValue, _ = k.VisibleTokenValue(ctx, sdk.NewCoins(specificBorrowed), types.PriceModeSpot)
			specificCollateralValue, _ = k.VisibleTokenValue(ctx, sdk.NewCoins(specificCollateral), types.PriceModeSpot)
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

// riskData
func (k Keeper) riskData(ctx sdk.Context) []types.AccountSymbolBalances {
	tokens := k.GetAllRegisteredTokens(ctx)
	symbols := map[string]string{}
	exponents := map[string]uint32{}
	exchangeRates := map[string]sdk.Dec{}
	for _, t := range tokens {
		symbols[t.BaseDenom] = t.SymbolDenom
		exponents[t.BaseDenom] = t.Exponent
		exchangeRates[t.BaseDenom] = k.DeriveExchangeRate(ctx, t.BaseDenom)
	}

	// which addresses have already been checked
	accounts := []types.AccountSymbolBalances{}
	checkedAddrs := map[string]interface{}{}

	prefix := types.KeyPrefixAdjustedBorrow
	iterator := func(key, _ []byte) error {
		// get borrower address from key
		addr := types.AddressFromKey(key, prefix)

		// if the address is already checked, do not check again
		if _, ok := checkedAddrs[addr.String()]; ok {
			return nil
		}
		checkedAddrs[addr.String()] = struct{}{}

		borrowed := k.GetBorrowerBorrows(ctx, addr)
		supplied, _ := k.GetAllSupplied(ctx, addr)
		collateral := k.GetBorrowerCollateral(ctx, addr)

		symbolBorrowed := symbolDecCoins(borrowed, symbols, exponents, exchangeRates)
		symbolSupplied := symbolDecCoins(supplied, symbols, exponents, exchangeRates)
		symbolCollateral := symbolDecCoins(collateral, symbols, exponents, exchangeRates)

		accounts = append(accounts, types.AccountSymbolBalances{
			Supplied:   symbolSupplied,
			Collateral: symbolCollateral,
			Borrowed:   symbolBorrowed,
		})

		return nil
	}

	// collect all account symbol balances (unsorted)
	_ = k.iterate(ctx, prefix, iterator)
	return accounts
}

// symbolDecCoins converts an sdk.Coins containing base tokens or uTokens into an sdk.DecCoins containing symbol denom
// base tokens. for example, 1000u/uumee becomes 0.0015UMEE at an exponent of 6 and uToken exchange rate of 1.5
func symbolDecCoins(
	coins sdk.Coins,
	symbols map[string]string,
	exponents map[string]uint32,
	exchangeRates map[string]sdk.Dec,
) sdk.DecCoins {
	symbolCoins := sdk.NewDecCoins()

	for _, c := range coins {
		if _, ok := symbols[c.Denom]; !ok {
			// unregistered tokens cannot be converted, but can be returned as base denom
			symbolCoins = symbolCoins.Add(sdk.NewDecCoinFromDec(c.Denom, sdk.NewDecFromInt(c.Amount)))
			continue
		}

		exchangeRate := sdk.OneDec()
		if types.HasUTokenPrefix(c.Denom) {
			c.Denom = types.ToTokenDenom(c.Denom)
			exchangeRate = exchangeRates[c.Denom]
		}
		exponentMultiplier := ten.Power(uint64(exponents[c.Denom]))
		denom := symbols[c.Denom]
		amount := exchangeRate.MulInt(c.Amount).Mul(exponentMultiplier)
		symbolCoins = symbolCoins.Add(sdk.NewDecCoinFromDec(denom, amount))
	}

	return symbolCoins
}
