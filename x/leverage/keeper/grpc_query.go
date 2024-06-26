package keeper

import (
	"context"

	sdkmath "cosmossdk.io/math"
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
	params := q.GetParams(ctx)

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
		token, err := q.GetTokenSettings(ctx, req.BaseDenom)
		if err != nil {
			return nil, err
		}
		tokens = append(tokens, token)
	} else {
		tokens = q.GetAllRegisteredTokens(ctx)
	}

	return &types.QueryRegisteredTokensResponse{
		Registry: tokens,
	}, nil
}

func (q Querier) RegisteredTokensWithMarkets(
	goCtx context.Context,
	req *types.QueryRegisteredTokensWithMarkets,
) (*types.QueryRegisteredTokensWithMarketsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	tokens := q.Keeper.GetAllRegisteredTokens(ctx)
	markets := []types.TokenMarket{}

	for _, token := range tokens {
		marketSumnmary, err := q.MarketSummary(goCtx, &types.QueryMarketSummary{Denom: token.BaseDenom})
		if err != nil {
			// absorb overall query error into struct, which may be empty, but proceed with this query
			marketSumnmary.Errors += err.Error()
		}
		markets = append(markets, types.TokenMarket{
			Token:  token,
			Market: *marketSumnmary,
		})
	}

	return &types.QueryRegisteredTokensWithMarketsResponse{
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
		pairs = q.GetAllSpecialAssetPairs(ctx)
	} else {
		// only pairs affecting one asset
		pairs = q.GetSpecialAssetPairs(ctx, req.Denom)
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
	token, err := q.GetTokenSettings(ctx, req.Denom)
	if err != nil {
		return nil, err
	}

	rate := q.DeriveExchangeRate(ctx, req.Denom)
	supplyAPY := q.DeriveSupplyAPY(ctx, req.Denom)
	borrowAPY := q.DeriveBorrowAPY(ctx, req.Denom)

	supplied, _ := q.GetTotalSupply(ctx, req.Denom)
	balance := q.ModuleBalance(ctx, req.Denom).Amount
	reserved := q.GetReserves(ctx, req.Denom).Amount
	borrowed := q.GetTotalBorrowed(ctx, req.Denom)
	liquidity := q.AvailableLiquidity(ctx, req.Denom)

	uDenom := coin.ToUTokenDenom(req.Denom)
	uSupply := q.GetUTokenSupply(ctx, uDenom)
	uCollateral := q.GetTotalCollateral(ctx, uDenom)

	// maxBorrow is based on MaxSupplyUtilization
	maxBorrow := token.MaxSupplyUtilization.MulInt(supplied.Amount).TruncateInt()

	// minimum liquidity respects both MaxSupplyUtilization and MinCollateralLiquidity
	minLiquidityFromSupply := supplied.Amount.Sub(maxBorrow)
	minLiquidityFromCollateral := token.MinCollateralLiquidity.Mul(rate.MulInt(uCollateral.Amount)).TruncateInt()
	minLiquidity := sdkmath.MinInt(minLiquidityFromCollateral, minLiquidityFromSupply)

	// availableBorrow respects both maxBorrow and minLiquidity
	availableBorrow := liquidity.Sub(minLiquidity)
	availableBorrow = sdkmath.MinInt(availableBorrow, maxBorrow.Sub(borrowed.Amount))
	availableBorrow = sdkmath.MaxInt(availableBorrow, sdkmath.ZeroInt())

	// availableWithdraw is based on minLiquidity
	availableWithdraw := liquidity.Sub(minLiquidity)
	availableWithdraw = sdkmath.MaxInt(availableWithdraw, sdkmath.ZeroInt())

	// availableCollateralize respects both MaxCollateralShare and MinCollateralLiquidity
	maxCollateral, _ := q.maxCollateralFromShare(ctx, uDenom)
	if token.MinCollateralLiquidity.IsPositive() {
		maxCollateralFromLiquidity := toDec(liquidity).Quo(token.MinCollateralLiquidity).TruncateInt()
		maxCollateral = sdkmath.MinInt(maxCollateral, maxCollateralFromLiquidity)
	}
	availableCollateralize := maxCollateral.Sub(uCollateral.Amount)
	availableCollateralize = sdkmath.MaxInt(availableCollateralize, sdkmath.ZeroInt())

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
	oraclePrice, _, oracleErr := q.TokenPrice(ctx, req.Denom, types.PriceModeQuery)
	if oracleErr == nil {
		resp.OraclePrice = &oraclePrice
	} else {
		resp.Errors += oracleErr.Error()
	}
	historicPrice, _, historicErr := q.TokenPrice(ctx, req.Denom, types.PriceModeHistoric)
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

	supplied, err := q.GetAllSupplied(ctx, addr)
	if err != nil {
		return nil, err
	}
	collateral := q.GetBorrowerCollateral(ctx, addr)
	borrowed := q.GetBorrowerBorrows(ctx, addr)

	return &types.QueryAccountBalancesResponse{
		Supplied:   supplied,
		Collateral: collateral,
		Borrowed:   borrowed,
	}, nil
}

func (q Querier) LiquidationTargets(
	goCtx context.Context,
	req *types.QueryLiquidationTargets,
) (*types.QueryLiquidationTargetsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	if !q.liquidatorQueryEnabled {
		return nil, types.ErrNotLiquidatorNode
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	targets, err := q.GetEligibleLiquidationTargets(ctx)
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
	targets := q.getAllBadDebts(ctx)

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
		for _, t := range q.GetAllRegisteredTokens(ctx) {
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

		userMaxWithdrawUToken, _, err := q.Keeper.userMaxWithdraw(ctx, addr, denom)
		if err != nil {
			if nonOracleError(err) {
				return nil, err
			}
			continue
		}

		moduleMaxWithdrawUToken, err := q.Keeper.ModuleMaxWithdraw(ctx, userMaxWithdrawUToken)
		if err != nil {
			if nonOracleError(err) {
				return nil, err
			}
			continue
		}

		uToken := sdk.NewCoin(userMaxWithdrawUToken.Denom, sdk.MinInt(userMaxWithdrawUToken.Amount, moduleMaxWithdrawUToken))
		if uToken.IsPositive() {
			token, err := q.ToToken(ctx, uToken)
			if err != nil {
				return nil, err
			}
			maxUTokens = maxUTokens.Add(uToken)
			maxTokens = maxTokens.Add(token)
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
		for _, t := range q.GetAllRegisteredTokens(ctx) {
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
		maxBorrow, err := q.userMaxBorrow(ctx, addr, denom)
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
