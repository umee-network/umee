package keeper

import (
	"context"
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

	tokens := k.GetAllRegisteredTokens(ctx)
	symbols := map[string]string{}
	exponents := map[string]uint32{}
	exchangeRates := map[string]sdk.Dec{}
	for _, t := range tokens {
		symbols[t.BaseDenom] = t.SymbolDenom
		exponents[t.BaseDenom] = t.Exponent
		exchangeRates[t.BaseDenom] = k.DeriveExchangeRate(ctx, t.BaseDenom)
		if strings.EqualFold(t.SymbolDenom, req.Symbol) {
			// convert request to match the case of token registry symbol denom
			req.Symbol = t.SymbolDenom
		}
	}

	// inspect every borrower only once
	borrowers := []*types.InspectAccount{}
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
		borrowedValue, _ := k.TotalTokenValue(ctx, borrowed, types.PriceModeSpot)
		collateral := k.GetBorrowerCollateral(ctx, addr)
		collateralValue, _ := k.CalculateCollateralValue(ctx, collateral)
		liquidationThreshold, _ := k.CalculateLiquidationThreshold(ctx, collateral)

		summary := types.InspectAccount{
			Address: addr.String(),
			Analysis: &types.RiskInfo{
				Borrowed:    neat(borrowedValue),
				Liquidation: neat(liquidationThreshold),
				Value:       neat(collateralValue),
			},
			Position: &types.DecBalances{
				Collateral: symbolDecCoins(collateral, symbols, exponents, exchangeRates),
				Borrowed:   symbolDecCoins(borrowed, symbols, exponents, exchangeRates),
			},
		}
		borrowers = append(borrowers, &summary)
		return nil
	}

	// collect all accounts (unsorted)
	_ = k.iterate(ctx, prefix, iterator)

	// filters the borrowers
	filteredBorrowers := []*types.InspectAccount{}
	for _, account := range borrowers {
		ok := account.Analysis.Borrowed > req.Borrowed
		ok = ok && account.Analysis.Value > req.Collateral
		ok = ok && account.Analysis.Liquidation > req.Danger
		ok = ok && account.Analysis.Borrowed/account.Analysis.Value > req.Ltv
		if ok {
			filteredBorrowers = append(filteredBorrowers, account)
		}
	}

	// sorts the borrowers
	sort.Sort(byCustom{
		bs:   filteredBorrowers,
		less: moreBorrowed(req.Symbol),
	})

	// convert from pointers
	sortedBorrowers := []types.InspectAccount{}
	for _, b := range filteredBorrowers {
		sortedBorrowers = append(sortedBorrowers, *b)
	}
	return &types.QueryInspectResponse{Borrowers: sortedBorrowers}, nil
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

// byCustom implements sort.Interface by providing Less and using the Len and
// Swap methods of the embedded bsums value.
type byCustom struct {
	bs   []*types.InspectAccount
	less inspectorSort
}

func (s byCustom) Len() int           { return len(s.bs) }
func (s byCustom) Swap(i, j int)      { s.bs[i], s.bs[j] = s.bs[j], s.bs[i] }
func (s byCustom) Less(i, j int) bool { return s.less(s.bs[i], s.bs[j]) }

// inspectorSort defines a Less function for sorting inspected borrower summaries,
// which must return true if a should come before b using custom logic for sorts.
type inspectorSort func(a, b *types.InspectAccount) bool

// moreBorrowed sorts accounts by borrowed amount of a given symbol denom, or
// borrowed USD value if no symbol denom is provided
func moreBorrowed(symbol string) inspectorSort {
	if symbol != "" {
		return func(a, b *types.InspectAccount) bool {
			return a.Position.Borrowed.AmountOf(symbol).GTE(b.Position.Borrowed.AmountOf(symbol))
		}
	}
	return func(a, b *types.InspectAccount) bool {
		return a.Analysis.Borrowed >= b.Analysis.Borrowed
	}
}