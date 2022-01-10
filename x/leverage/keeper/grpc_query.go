package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/umee-network/umee/x/leverage/types"
)

var _ types.QueryServer = Querier{}

// Querier implements a QueryServer for the x/leverage module.
type Querier struct {
	Keeper
}

func NewQuerier(k Keeper) Querier {
	return Querier{Keeper: k}
}

func (q Querier) RegisteredTokens(
	goCtx context.Context,
	req *types.QueryRegisteredTokens,
) (*types.QueryRegisteredTokensResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	tokens, err := q.Keeper.GetAllRegisteredTokens(ctx)
	if err != nil {
		return nil, err
	}

	resp := &types.QueryRegisteredTokensResponse{
		Registry: make([]types.Token, len(tokens)),
	}

	for i, t := range tokens {
		resp.Registry[i] = t
	}

	return resp, nil
}

func (q Querier) Params(
	goCtx context.Context,
	req *types.QueryParamsRequest,
) (*types.QueryParamsResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	params := q.Keeper.GetParams(ctx)

	return &types.QueryParamsResponse{Params: params}, nil
}

func (q Querier) Borrowed(
	goCtx context.Context,
	req *types.QueryBorrowedRequest,
) (*types.QueryBorrowedResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid address")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	borrower, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}

	if len(req.Denom) == 0 {
		tokens := q.Keeper.GetBorrowerBorrows(ctx, borrower)

		return &types.QueryBorrowedResponse{Borrowed: tokens}, nil
	}

	if !q.Keeper.IsAcceptedToken(ctx, req.Denom) {
		return nil, status.Error(codes.InvalidArgument, "not accepted Token denom")
	}

	token := q.Keeper.GetBorrow(ctx, borrower, req.Denom)

	return &types.QueryBorrowedResponse{Borrowed: sdk.NewCoins(token)}, nil
}

func (q Querier) AvailableBorrow(
	goCtx context.Context,
	req *types.QueryAvailableBorrowRequest,
) (*types.QueryAvailableBorrowResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.Denom == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid denom")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if !q.Keeper.IsAcceptedToken(ctx, req.Denom) {
		return nil, status.Error(codes.InvalidArgument, "not accepted Token denom")
	}

	// Available for borrow = Module Balance - Reserve Amount
	moduleBalance := q.Keeper.ModuleBalance(ctx, req.Denom)
	reserveAmount := q.Keeper.GetReserveAmount(ctx, req.Denom)

	return &types.QueryAvailableBorrowResponse{Amount: moduleBalance.Sub(reserveAmount).ToDec()}, nil
}

func (q Querier) BorrowAPY(
	goCtx context.Context,
	req *types.QueryBorrowAPYRequest,
) (*types.QueryBorrowAPYResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.Denom == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid denom")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if !q.Keeper.IsAcceptedToken(ctx, req.Denom) {
		return nil, status.Error(codes.InvalidArgument, "not accepted Token denom")
	}

	borrowAPY := q.Keeper.GetBorrowAPY(ctx, req.Denom)

	return &types.QueryBorrowAPYResponse{APY: borrowAPY}, nil
}

func (q Querier) LendAPY(
	goCtx context.Context,
	req *types.QueryLendAPYRequest,
) (*types.QueryLendAPYResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.Denom == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid denom")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if !q.Keeper.IsAcceptedToken(ctx, req.Denom) {
		return nil, status.Error(codes.InvalidArgument, "not accepted Token denom")
	}

	lendAPY := q.Keeper.GetLendAPY(ctx, req.Denom)

	return &types.QueryLendAPYResponse{APY: lendAPY}, nil
}

func (q Querier) MarketSize(
	goCtx context.Context,
	req *types.QueryMarketSizeRequest,
) (*types.QueryMarketSizeResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.Denom == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid denom")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	if !q.Keeper.IsAcceptedToken(ctx, req.Denom) {
		return nil, status.Error(codes.InvalidArgument, "not accepted Token denom")
	}

	uTokenDenom := q.Keeper.FromTokenToUTokenDenom(ctx, req.Denom)
	marketSizeCoin, err := q.Keeper.ExchangeUToken(ctx, q.Keeper.TotalUTokenSupply(ctx, uTokenDenom))
	if err != nil {
		return nil, err
	}

	marketSizeUSD, err := q.Keeper.TokenValue(ctx, marketSizeCoin)
	if err != nil {
		return nil, err
	}

	return &types.QueryMarketSizeResponse{MarketSizeUsd: marketSizeUSD}, nil
}

func (q Querier) ReserveAmount(
	goCtx context.Context,
	req *types.QueryReserveAmountRequest,
) (*types.QueryReserveAmountResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.Denom == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid denom")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !q.Keeper.IsAcceptedToken(ctx, req.Denom) {
		return nil, status.Error(codes.InvalidArgument, "not accepted Token denom")
	}

	amount := q.Keeper.GetReserveAmount(ctx, req.Denom)

	return &types.QueryReserveAmountResponse{Amount: amount}, nil
}

func (q Querier) CollateralSetting(
	goCtx context.Context,
	req *types.QueryCollateralSettingRequest,
) (*types.QueryCollateralSettingResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid address")
	}
	if req.Denom == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid denom")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	borrower, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}

	if !q.Keeper.IsAcceptedUToken(ctx, req.Denom) {
		return nil, status.Error(codes.InvalidArgument, "not accepted uToken denom")
	}

	setting := q.Keeper.GetCollateralSetting(ctx, borrower, req.Denom)

	return &types.QueryCollateralSettingResponse{Enabled: setting}, nil
}

func (q Querier) Collateral(
	goCtx context.Context,
	req *types.QueryCollateralRequest,
) (*types.QueryCollateralResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid address")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	borrower, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}

	if len(req.Denom) == 0 {
		tokens := q.Keeper.GetBorrowerCollateral(ctx, borrower)

		return &types.QueryCollateralResponse{Collateral: tokens}, nil
	}

	if !q.Keeper.IsAcceptedUToken(ctx, req.Denom) {
		return nil, status.Error(codes.InvalidArgument, "not accepted uToken denom")
	}

	token := q.Keeper.GetCollateralAmount(ctx, borrower, req.Denom)

	return &types.QueryCollateralResponse{Collateral: sdk.NewCoins(token)}, nil
}

func (q Querier) ExchangeRate(
	goCtx context.Context,
	req *types.QueryExchangeRateRequest,
) (*types.QueryExchangeRateResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.Denom == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid denom")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	if !q.Keeper.IsAcceptedToken(ctx, req.Denom) {
		return nil, status.Error(codes.InvalidArgument, "not accepted Token denom")
	}

	rate, err := q.Keeper.GetExchangeRate(ctx, req.Denom)
	if err != nil {
		return nil, err
	}

	return &types.QueryExchangeRateResponse{ExchangeRate: rate}, nil
}

func (q Querier) BorrowLimit(
	goCtx context.Context,
	req *types.QueryBorrowLimitRequest,
) (*types.QueryBorrowLimitResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid address")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	borrower, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}

	collateral := q.Keeper.GetBorrowerCollateral(ctx, borrower)

	limit, err := q.Keeper.CalculateBorrowLimit(ctx, collateral)
	if err != nil {
		return nil, err
	}

	return &types.QueryBorrowLimitResponse{BorrowLimit: limit}, nil
}

func (q Querier) LiquidationTargets(
	goCtx context.Context,
	req *types.QueryLiquidationTargetsRequest,
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
