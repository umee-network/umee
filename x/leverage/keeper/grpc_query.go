package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/umee-network/umee/v6/util/coin"
	"github.com/umee-network/umee/v6/x/leverage/types"
)

var _ types.QueryServer = Querier{}

// Querier implements a QueryServer for the x/leverage module.
type Querier struct {
	Keeper
}

func NewQuerier(k Keeper) Querier {
	return Querier{Keeper: k}
}

func (q Querier) Params(
	goCtx context.Context,
	req *types.QueryParams,
) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	params := q.Keeper.GetParams(ctx)

	return &types.QueryParamsResponse{Params: params}, nil
}

func (q Querier) RegisteredTokens(
	goCtx context.Context,
	req *types.QueryRegisteredTokens,
) (*types.QueryRegisteredTokensResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	var tokens []types.Token
	if len(req.BaseDenom) != 0 {
		token, err := q.Keeper.GetTokenSettings(ctx, req.BaseDenom)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	} else {
		tokens = q.Keeper.GetAllRegisteredTokens(ctx)
	}

	return &types.QueryRegisteredTokensResponse{
		Registry: tokens,
	}, nil
}

func (q Querier) RegisteredTokenMarkets(
	goCtx context.Context,
	req *types.QueryRegisteredTokenMarkets,
) (*types.QueryRegisteredTokenMarketsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	tokens := q.Keeper.GetAllRegisteredTokens(ctx)
	markets := []types.TokenMarket{}

	for _, token := range tokens {
		marketSumnmary, err := q.MarketSummary(goCtx, &types.QueryMarketSummary{Denom: token.BaseDenom})
		if err != nil {
			// absorb error overall query error into struct, which may be empty, but proceed with this query
			marketSumnmary.Errors = marketSumnmary.Errors + err.Error()
		}
		markets = append(markets, types.TokenMarket{
			Token:  &token,
			Market: marketSumnmary,
		})
	}

	return &types.QueryRegisteredTokenMarketsResponse{
		Markets: markets,
	}, nil
}

func (q Querier) SpecialAssets(
	goCtx context.Context,
	req *types.QuerySpecialAssets,
) (*types.QuerySpecialAssetsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	var pairs []types.SpecialAssetPair
	if req.Denom == "" {
		// all pairs
		pairs = q.Keeper.GetAllSpecialAssetPairs(ctx)
	} else {
		// only pairs affecting one asset
		pairs = q.Keeper.GetSpecialAssetPairs(ctx, req.Denom)
	}

	return &types.QuerySpecialAssetsResponse{
		Pairs: pairs,
	}, nil
}

func (q Querier) MarketSummary(
	goCtx context.Context,
	req *types.QueryMarketSummary,
) (*types.QueryMarketSummaryResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.Denom == "" {
		return nil, status.Error(codes.InvalidArgument, "empty denom")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	token, err := q.Keeper.GetTokenSettings(ctx, req.Denom)
	if err != nil {
		return nil, err
	}

	rate := q.Keeper.DeriveExchangeRate(ctx, req.Denom)
	supplyAPY := q.Keeper.DeriveSupplyAPY(ctx, req.Denom)
	borrowAPY := q.Keeper.DeriveBorrowAPY(ctx, req.Denom)

	supplied, _ := q.Keeper.GetTotalSupply(ctx, req.Denom)
	balance := q.Keeper.ModuleBalance(ctx, req.Denom).Amount
	reserved := q.Keeper.GetReserves(ctx, req.Denom).Amount
	borrowed := q.Keeper.GetTotalBorrowed(ctx, req.Denom)
	liquidity := q.Keeper.AvailableLiquidity(ctx, req.Denom)

	uDenom := coin.ToUTokenDenom(req.Denom)
	uSupply := q.Keeper.GetUTokenSupply(ctx, uDenom)
	uCollateral := q.Keeper.GetTotalCollateral(ctx, uDenom)

	// maxBorrow is based on MaxSupplyUtilization
	maxBorrow := token.MaxSupplyUtilization.MulInt(supplied.Amount).TruncateInt()

	// minimum liquidity respects both MaxSupplyUtilization and MinCollateralLiquidity
	minLiquidityFromSupply := supplied.Amount.Sub(maxBorrow)
	minLiquidityFromCollateral := token.MinCollateralLiquidity.Mul(rate.MulInt(uCollateral.Amount)).TruncateInt()
	minLiquidity := sdk.MinInt(minLiquidityFromCollateral, minLiquidityFromSupply)

	// availableBorrow respects both maxBorrow and minLiquidity
	availableBorrow := liquidity.Sub(minLiquidity)
	availableBorrow = sdk.MinInt(availableBorrow, maxBorrow.Sub(borrowed.Amount))
	availableBorrow = sdk.MaxInt(availableBorrow, sdk.ZeroInt())

	// availableWithdraw is based on minLiquidity
	availableWithdraw := liquidity.Sub(minLiquidity)
	availableWithdraw = sdk.MaxInt(availableWithdraw, sdk.ZeroInt())

	// availableCollateralize respects both MaxCollateralShare and MinCollateralLiquidity
	maxCollateral, _ := q.Keeper.maxCollateralFromShare(ctx, uDenom)
	if token.MinCollateralLiquidity.IsPositive() {
		maxCollateralFromLiquidity := toDec(liquidity).Quo(token.MinCollateralLiquidity).TruncateInt()
		maxCollateral = sdk.MinInt(maxCollateral, maxCollateralFromLiquidity)
	}
	availableCollateralize := maxCollateral.Sub(uCollateral.Amount)
	availableCollateralize = sdk.MaxInt(availableCollateralize, sdk.ZeroInt())

	resp := types.QueryMarketSummaryResponse{
		SymbolDenom:            token.SymbolDenom,
		Exponent:               token.Exponent,
		UTokenExchangeRate:     rate,
		Supply_APY:             supplyAPY,
		Borrow_APY:             borrowAPY,
		Supplied:               supplied.Amount,
		Reserved:               reserved,
		Collateral:             uCollateral.Amount,
		Borrowed:               borrowed.Amount,
		Liquidity:              balance.Sub(reserved),
		MaximumBorrow:          maxBorrow,
		MaximumCollateral:      maxCollateral,
		MinimumLiquidity:       minLiquidity,
		UTokenSupply:           uSupply.Amount,
		AvailableBorrow:        availableBorrow,
		AvailableWithdraw:      availableWithdraw,
		AvailableCollateralize: availableCollateralize,
	}

	// Oracle price in response will be nil if the oracle module has no price at all, but will instead
	// show the most recent price if one existed.
	oraclePrice, _, oracleErr := q.Keeper.TokenPrice(ctx, req.Denom, types.PriceModeQuery)
	if oracleErr == nil {
		resp.OraclePrice = &oraclePrice
	} else {
		resp.Errors += oracleErr.Error()
	}
	historicPrice, _, historicErr := q.Keeper.TokenPrice(ctx, req.Denom, types.PriceModeHistoric)
	if historicErr == nil {
		resp.OracleHistoricPrice = &historicPrice
	} else {
		resp.Errors += historicErr.Error()
	}

	return &resp, nil
}

func (q Querier) AccountBalances(
	goCtx context.Context,
	req *types.QueryAccountBalances,
) (*types.QueryAccountBalancesResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "empty address")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}

	supplied, err := q.Keeper.GetAllSupplied(ctx, addr)
	if err != nil {
		return nil, err
	}
	collateral := q.Keeper.GetBorrowerCollateral(ctx, addr)
	borrowed := q.Keeper.GetBorrowerBorrows(ctx, addr)

	return &types.QueryAccountBalancesResponse{
		Supplied:   supplied,
		Collateral: collateral,
		Borrowed:   borrowed,
	}, nil
}

func (q Querier) AccountSummary(
	goCtx context.Context,
	req *types.QueryAccountSummary,
) (*types.QueryAccountSummaryResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "empty address")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}

	supplied, err := q.Keeper.GetAllSupplied(ctx, addr)
	if err != nil {
		return nil, err
	}
	collateral := q.Keeper.GetBorrowerCollateral(ctx, addr)
	borrowed := q.Keeper.GetBorrowerBorrows(ctx, addr)

	// the following price calculations use the most recent prices if spot prices are missing
	lastSuppliedValue, err := q.Keeper.VisibleTokenValue(ctx, supplied, types.PriceModeQuery)
	if err != nil {
		return nil, err
	}
	lastBorrowedValue, err := q.Keeper.VisibleTokenValue(ctx, borrowed, types.PriceModeQuery)
	if err != nil {
		return nil, err
	}
	lastCollateralValue, err := q.Keeper.VisibleCollateralValue(ctx, collateral, types.PriceModeQuery)
	if err != nil {
		return nil, err
	}

	// these use leverage-like prices: the lower of spot or historic price for supplied tokens and higher for borrowed.
	// unlike transactions, this query will use expired prices instead of skipping them.
	suppliedValue, err := q.Keeper.VisibleTokenValue(ctx, supplied, types.PriceModeQueryLow)
	if err != nil {
		return nil, err
	}
	collateralValue, err := q.Keeper.VisibleCollateralValue(ctx, collateral, types.PriceModeQueryLow)
	if err != nil {
		return nil, err
	}
	borrowedValue, err := q.Keeper.VisibleTokenValue(ctx, borrowed, types.PriceModeQueryHigh)
	if err != nil {
		return nil, err
	}

	resp := &types.QueryAccountSummaryResponse{
		SuppliedValue:       suppliedValue,
		CollateralValue:     collateralValue,
		BorrowedValue:       borrowedValue,
		SpotSuppliedValue:   lastSuppliedValue,
		SpotCollateralValue: lastCollateralValue,
		SpotBorrowedValue:   lastBorrowedValue,
	}

	// values computed from position use the same prices found in leverage logic:
	// using the lower of spot or historic prices for each collateral token
	// and the higher of spot or historic prices for each borrowed token
	// skips collateral tokens with missing prices, but errors on borrow tokens missing prices
	// (for oracle errors only the relevant response fields will be left nil)
	ap, err := q.Keeper.GetAccountPosition(ctx, addr, false)
	if nonOracleError(err) {
		return nil, err
	}
	if err == nil {
		// on missing borrow price, borrow limit is nil
		borrowLimit := ap.Limit()
		resp.BorrowLimit = &borrowLimit
	}

	// liquidation threshold shown here as it is used in leverage logic: using spot prices.
	// skips borrowed tokens with missing prices, but errors on collateral missing prices
	// (for oracle errors only the relevant response fields will be left nil)
	ap, err = q.Keeper.GetAccountPosition(ctx, addr, true)
	if nonOracleError(err) {
		return nil, err
	}
	if err == nil {
		// on missing collateral price, liquidation threshold is nil
		liquidationThreshold := ap.Limit()
		resp.LiquidationThreshold = &liquidationThreshold
	}

	return resp, nil
}

func (q Querier) LiquidationTargets(
	goCtx context.Context,
	req *types.QueryLiquidationTargets,
) (*types.QueryLiquidationTargetsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if !q.Keeper.liquidatorQueryEnabled {
		return nil, types.ErrNotLiquidatorNode
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	targets, err := q.Keeper.GetEligibleLiquidationTargets(ctx)
	if err != nil {
		return nil, err
	}

	stringTargets := []string{}
	for _, addr := range targets {
		stringTargets = append(stringTargets, addr.String())
	}

	return &types.QueryLiquidationTargetsResponse{Targets: stringTargets}, nil
}

func (q Querier) BadDebts(
	goCtx context.Context,
	req *types.QueryBadDebts,
) (*types.QueryBadDebtsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	targets := q.Keeper.getAllBadDebts(ctx)

	return &types.QueryBadDebtsResponse{Targets: targets}, nil
}

func (q Querier) MaxWithdraw(
	goCtx context.Context,
	req *types.QueryMaxWithdraw,
) (*types.QueryMaxWithdrawResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if err := req.ValidateBasic(); err != nil {
		return nil, err
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}

	denoms := []string{}
	maxUTokens := sdk.NewCoins()
	maxTokens := sdk.NewCoins()

	if req.Denom != "" {
		// Denom specified
		denoms = []string{req.Denom}
	} else {
		// Denom not specified
		for _, t := range q.Keeper.GetAllRegisteredTokens(ctx) {
			if !t.Blacklist {
				denoms = append(denoms, t.BaseDenom)
			}
		}
	}

	for _, denom := range denoms {
		// If a price is missing for the borrower's collateral,
		// but not this uToken or any of their borrows, error
		// will be nil and the resulting value will be what
		// can safely be withdrawn even with missing prices.
		// On non-nil error here, max withdraw is zero.
		uToken, _, err := q.Keeper.userMaxWithdraw(ctx, addr, denom)
		if err == nil && uToken.IsPositive() {
			token, err := q.Keeper.ToToken(ctx, uToken)
			if err != nil {
				return nil, err
			}
			maxUTokens = maxUTokens.Add(uToken)
			maxTokens = maxTokens.Add(token)
		}
		// Non-price errors will cause the query itself to fail
		if nonOracleError(err) {
			return nil, err
		}
	}

	return &types.QueryMaxWithdrawResponse{
		Tokens:  maxTokens,
		UTokens: maxUTokens,
	}, nil
}

func (q Querier) MaxBorrow(
	goCtx context.Context,
	req *types.QueryMaxBorrow,
) (*types.QueryMaxBorrowResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "empty address")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}

	denoms := []string{}
	maxTokens := sdk.NewCoins()

	if req.Denom != "" {
		// Denom specified
		denoms = []string{req.Denom}
	} else {
		// Denom not specified
		for _, t := range q.Keeper.GetAllRegisteredTokens(ctx) {
			if !t.Blacklist {
				denoms = append(denoms, t.BaseDenom)
			}
		}
	}

	for _, denom := range denoms {
		// If a price is missing for the borrower's collateral,
		// but not this token or any of their borrows, error
		// will be nil and the resulting value will be what
		// can safely be borrowed even with missing prices.
		// On non-nil error here, max borrow is zero.
		maxBorrow, err := q.Keeper.userMaxBorrow(ctx, addr, denom)
		if err == nil && maxBorrow.IsPositive() {
			maxTokens = maxTokens.Add(maxBorrow)
		}
		// Non-price errors will cause the query itself to fail
		if nonOracleError(err) {
			return nil, err
		}
	}

	return &types.QueryMaxBorrowResponse{
		Tokens: maxTokens,
	}, nil
}
