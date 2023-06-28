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

type tokenExchangeRate struct {
	symbol       string
	exponent     uint32
	exchangeRate sdk.Dec
}

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

	exchangeRates := map[string]tokenExchangeRate{}
	for _, t := range tokens {
		exchangeRates[t.BaseDenom] = tokenExchangeRate{
			symbol:       t.SymbolDenom,
			exponent:     t.Exponent,
			exchangeRate: k.DeriveExchangeRate(ctx, t.BaseDenom),
		}
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
		collateralValue, _ := k.CalculateCollateralValue(ctx, collateral, types.PriceModeSpot)
		liquidationThreshold, _ := k.CalculateLiquidationThreshold(ctx, collateral)

		account := types.InspectAccount{
			Address: addr.String(),
			Analysis: &types.RiskInfo{
				Borrowed:    neat(borrowedValue),
				Liquidation: neat(liquidationThreshold),
				Value:       neat(collateralValue),
			},
			Position: &types.DecBalances{
				Collateral: symbolDecCoins(collateral, exchangeRates),
				Borrowed:   symbolDecCoins(borrowed, exchangeRates),
			},
		}
		ok := account.Analysis.Borrowed > req.Borrowed
		ok = ok && account.Analysis.Value > req.Collateral
		ok = ok && account.Analysis.Borrowed/account.Analysis.Liquidation > req.Danger
		ok = ok && account.Analysis.Borrowed/account.Analysis.Value > req.Ltv
		if ok {
			borrowers = append(borrowers, &account)
		}
		return nil
	}

	// collect all accounts (filtered but unsorted)
	_ = k.iterate(ctx, prefix, iterator)

	// sorts the borrowers
	sort.SliceStable(borrowers, func(i, j int) bool {
		if req.Symbol != "" {
			// for non-empty symbol denom, sorts by borrowed amount (descending) of that token
			return borrowers[i].Position.Borrowed.AmountOf(req.Symbol).GTE(borrowers[j].Position.Borrowed.AmountOf(req.Symbol))
		}
		// otherwise, sorts by borrowed value (descending)
		return borrowers[i].Analysis.Borrowed > borrowers[j].Analysis.Borrowed
	},
	)

	// convert from pointers
	sortedBorrowers := []types.InspectAccount{}
	for _, b := range borrowers {
		sortedBorrowers = append(sortedBorrowers, *b)
	}
	return &types.QueryInspectResponse{Borrowers: sortedBorrowers}, nil
}

// symbolDecCoins converts an sdk.Coins containing base tokens or uTokens into an sdk.DecCoins containing symbol denom
// base tokens. for example, 1000u/uumee becomes 0.0015UMEE at an exponent of 6 and uToken exchange rate of 1.5
func symbolDecCoins(
	coins sdk.Coins,
	tokens map[string]tokenExchangeRate,
) sdk.DecCoins {
	symbolCoins := sdk.NewDecCoins()

	for _, c := range coins {
		exchangeRate := sdk.OneDec()
		if types.HasUTokenPrefix(c.Denom) {
			// uTokens will be converted to base tokens
			c.Denom = types.ToTokenDenom(c.Denom)
			exchangeRate = tokens[c.Denom].exchangeRate
		}
		t, ok := tokens[c.Denom]
		if !ok {
			// unregistered tokens cannot be converted, but can be returned as base denom
			symbolCoins = symbolCoins.Add(sdk.NewDecCoinFromDec(c.Denom, sdk.NewDecFromInt(c.Amount)))
			continue
		}
		exponentMultiplier := ten.Power(uint64(t.exponent))
		amount := exchangeRate.MulInt(c.Amount).Quo(exponentMultiplier)
		symbolCoins = symbolCoins.Add(sdk.NewDecCoinFromDec(t.symbol, amount))
	}

	return symbolCoins
}

// neat truncates an sdk.Dec to a common-sense precision based on its size and converts it to float.
// This greatly improves readability when viewing balances.
func neat(num sdk.Dec) float64 {
	n := num.MustFloat64()
	a := math.Abs(n)
	precision := 2 // Round to cents if 0.01 <= n <= 100
	if a > 100 {
		precision = 0 // above $100: Round to dollar
	}
	if a > 1_000_000 {
		precision = -3 // above $1000000: Round to thousand
	}
	if a < 0.01 {
		precision = 6 // round to millionths
	}
	if a < 0.000001 {
		return n // tiny: maximum precision
	}
	// Truncate the float at a certain precision (can be negative)
	x := math.Pow(10, float64(precision))
	return float64(int(n*x)) / x
}
