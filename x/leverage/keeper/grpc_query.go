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

	token := q.Keeper.GetBorrow(ctx, borrower, req.Denom)

	return &types.QueryBorrowedResponse{Borrowed: sdk.NewCoins(token)}, nil
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
	rate, err := q.Keeper.GetExchangeRate(ctx, req.Denom)
	if err != nil {
		return nil, err
	}

	return &types.QueryExchangeRateResponse{ExchangeRate: rate}, nil
}
