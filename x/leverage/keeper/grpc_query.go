package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/umee-network/umee/v3/x/leverage/types"
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

	tokens := q.Keeper.GetAllRegisteredTokens(ctx)

	resp := &types.QueryRegisteredTokensResponse{
		Registry: make([]types.Token, len(tokens)),
	}

	copy(resp.Registry, tokens)

	return resp, nil
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
	balance := q.Keeper.ModuleBalance(ctx, req.Denom)
	reserved := q.Keeper.GetReserveAmount(ctx, req.Denom)
	borrowed := q.Keeper.GetTotalBorrowed(ctx, req.Denom)
	liquidity := q.Keeper.Liquidity(ctx, req.Denom)

	uDenom := types.ToUTokenDenom(req.Denom)
	uSupply := q.Keeper.GetUTokenSupply(ctx, uDenom)
	uCollateral := q.Keeper.GetTotalCollateral(ctx, uDenom)

	// maxBorrow is based on MaxSupplyUtilization
	maxBorrow := token.MaxSupplyUtilization.MulInt(supplied.Amount).TruncateInt()

	// availableBorrow respects maxBorrow but also liquidity
	// TODO #1163 - additional restriction
	availableBorrow := sdk.MinInt(liquidity, maxBorrow.Sub(borrowed.Amount))
	availableBorrow = sdk.MaxInt(availableBorrow, sdk.ZeroInt())

	// minimum liquidity respects both MaxSupplyUtilization and MinCollateralLiquidity
	// TODO #1163 - additional restriction
	minLiquidity := supplied.Amount.Sub(maxBorrow)

	maxCollateral, _ := q.Keeper.maxCollateral(ctx, uDenom)

	maxCollateral, _ := q.Keeper.maxCollateral(ctx, uDenom)

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
		AvailableWithdraw:      uSupply.Amount, // TODO #1163 - implement limits
		AvailableCollateralize: sdk.MaxInt(maxCollateral.Sub(uCollateral.Amount), sdk.ZeroInt()),
	}

	// Oracle price in response will be nil if it is unavailable
	if oraclePrice, oracleErr := q.Keeper.TokenPrice(ctx, req.Denom); oracleErr == nil {
		resp.OraclePrice = &oraclePrice
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

	suppliedValue, err := q.Keeper.TotalTokenValue(ctx, supplied)
	if err != nil {
		return nil, err
	}
	borrowedValue, err := q.Keeper.TotalTokenValue(ctx, borrowed)
	if err != nil {
		return nil, err
	}
	collateralValue, err := q.Keeper.CalculateCollateralValue(ctx, collateral)
	if err != nil {
		return nil, err
	}
	borrowLimit, err := q.Keeper.CalculateBorrowLimit(ctx, collateral)
	if err != nil {
		return nil, err
	}
	liquidationThreshold, err := q.Keeper.CalculateLiquidationThreshold(ctx, collateral)
	if err != nil {
		return nil, err
	}

	return &types.QueryAccountSummaryResponse{
		SuppliedValue:        suppliedValue,
		CollateralValue:      collateralValue,
		BorrowedValue:        borrowedValue,
		BorrowLimit:          borrowLimit,
		LiquidationThreshold: liquidationThreshold,
	}, nil
}

func (q Querier) LiquidationTargets(
	goCtx context.Context,
	req *types.QueryLiquidationTargets,
) (*types.QueryLiquidationTargetsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
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
