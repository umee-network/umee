package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/umee-network/umee/v2/x/leverage/types"
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

	tokens := q.Keeper.GetAllRegisteredTokens(ctx)

	resp := &types.QueryRegisteredTokensResponse{
		Registry: make([]types.Token, len(tokens)),
	}

	copy(resp.Registry, tokens)

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

func (q Querier) BorrowedValue(
	goCtx context.Context,
	req *types.QueryBorrowedValueRequest,
) (*types.QueryBorrowedValueResponse, error) {
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

	var tokens sdk.Coins

	if len(req.Denom) == 0 {
		tokens = q.Keeper.GetBorrowerBorrows(ctx, borrower)
	} else {
		if !q.Keeper.IsAcceptedToken(ctx, req.Denom) {
			return nil, status.Error(codes.InvalidArgument, "not accepted Token denom")
		}

		tokens = sdk.NewCoins(q.Keeper.GetBorrow(ctx, borrower, req.Denom))
	}

	value, err := q.Keeper.TotalTokenValue(ctx, tokens)
	if err != nil {
		return nil, err
	}

	return &types.QueryBorrowedValueResponse{BorrowedValue: value}, nil
}

func (q Querier) Supplied(
	goCtx context.Context,
	req *types.QuerySuppliedRequest,
) (*types.QuerySuppliedResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid address")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}

	if len(req.Denom) == 0 {
		tokens, err := q.Keeper.GetAllSupplied(ctx, addr)
		if err != nil {
			return nil, err
		}

		return &types.QuerySuppliedResponse{Supplied: tokens}, nil
	}

	if !q.Keeper.IsAcceptedToken(ctx, req.Denom) {
		return nil, status.Error(codes.InvalidArgument, "not accepted Token denom")
	}

	token, err := q.Keeper.GetSupplied(ctx, addr, req.Denom)
	if err != nil {
		return nil, err
	}

	return &types.QuerySuppliedResponse{Supplied: sdk.NewCoins(token)}, nil
}

func (q Querier) SuppliedValue(
	goCtx context.Context,
	req *types.QuerySuppliedValueRequest,
) (*types.QuerySuppliedValueResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid address")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}

	var tokens sdk.Coins

	if len(req.Denom) == 0 {
		tokens, err = q.Keeper.GetAllSupplied(ctx, addr)
		if err != nil {
			return nil, err
		}
	} else {
		if !q.Keeper.IsAcceptedToken(ctx, req.Denom) {
			return nil, status.Error(codes.InvalidArgument, "not accepted Token denom")
		}

		supplied, err := q.Keeper.GetSupplied(ctx, addr, req.Denom)
		if err != nil {
			return nil, err
		}

		tokens = sdk.NewCoins(supplied)
	}

	value, err := q.Keeper.TotalTokenValue(ctx, tokens)
	if err != nil {
		return nil, err
	}

	return &types.QuerySuppliedValueResponse{SuppliedValue: value}, nil
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

	amountAvailable := q.Keeper.GetAvailableToBorrow(ctx, req.Denom)

	return &types.QueryAvailableBorrowResponse{Amount: amountAvailable}, nil
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

	borrowAPY := q.Keeper.DeriveBorrowAPY(ctx, req.Denom)

	return &types.QueryBorrowAPYResponse{APY: borrowAPY}, nil
}

func (q Querier) SupplyAPY(
	goCtx context.Context,
	req *types.QuerySupplyAPYRequest,
) (*types.QuerySupplyAPYResponse, error) {
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

	supplyAPY := q.Keeper.DeriveSupplyAPY(ctx, req.Denom)

	return &types.QuerySupplyAPYResponse{APY: supplyAPY}, nil
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

	marketSizeCoin, err := q.Keeper.GetTotalSupply(ctx, req.Denom)
	if err != nil {
		return nil, err
	}

	marketSizeUSD, err := q.Keeper.TokenValue(ctx, marketSizeCoin)
	if err != nil {
		return nil, err
	}

	return &types.QueryMarketSizeResponse{MarketSizeUsd: marketSizeUSD}, nil
}

func (q Querier) TokenMarketSize(
	goCtx context.Context,
	req *types.QueryTokenMarketSizeRequest,
) (*types.QueryTokenMarketSizeResponse, error) {
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

	marketSizeCoin, err := q.Keeper.GetTotalSupply(ctx, req.Denom)
	if err != nil {
		return nil, err
	}

	return &types.QueryTokenMarketSizeResponse{MarketSize: marketSizeCoin.Amount}, nil
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

func (q Querier) CollateralValue(
	goCtx context.Context,
	req *types.QueryCollateralValueRequest,
) (*types.QueryCollateralValueResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid address")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	addr, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}

	var uTokens sdk.Coins

	if len(req.Denom) == 0 {
		uTokens = q.Keeper.GetBorrowerCollateral(ctx, addr)
	} else {
		if !q.Keeper.IsAcceptedUToken(ctx, req.Denom) {
			return nil, status.Error(codes.InvalidArgument, "not accepted uToken denom")
		}

		collateral := q.Keeper.GetCollateralAmount(ctx, addr, req.Denom)
		uTokens = sdk.NewCoins(collateral)
	}

	tokens, err := q.Keeper.ExchangeUTokens(ctx, uTokens)
	if err != nil {
		return nil, err
	}

	value, err := q.Keeper.TotalTokenValue(ctx, tokens)
	if err != nil {
		return nil, err
	}

	return &types.QueryCollateralValueResponse{CollateralValue: value}, nil
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

	rate := q.Keeper.DeriveExchangeRate(ctx, req.Denom)

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

func (q Querier) LiquidationThreshold(
	goCtx context.Context,
	req *types.QueryLiquidationThresholdRequest,
) (*types.QueryLiquidationThresholdResponse, error) {
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

	t, err := q.Keeper.CalculateLiquidationThreshold(ctx, collateral)
	if err != nil {
		return nil, err
	}

	return &types.QueryLiquidationThresholdResponse{LiquidationThreshold: t}, nil
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

func (q Querier) MarketSummary(
	goCtx context.Context,
	req *types.QueryMarketSummaryRequest,
) (*types.QueryMarketSummaryResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.Denom == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid denom")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	token, err := q.Keeper.GetTokenSettings(ctx, req.Denom)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "not accepted Token denom")
	}
	rate := q.Keeper.DeriveExchangeRate(ctx, req.Denom)
	supplyAPY := q.Keeper.DeriveSupplyAPY(ctx, req.Denom)
	borrowAPY := q.Keeper.DeriveBorrowAPY(ctx, req.Denom)
	marketSizeCoin, _ := q.Keeper.GetTotalSupply(ctx, req.Denom)
	availableBorrow := q.Keeper.GetAvailableToBorrow(ctx, req.Denom)
	reserved := q.Keeper.GetReserveAmount(ctx, req.Denom)
	collateral := q.Keeper.GetTotalCollateral(ctx, req.Denom)

	resp := types.QueryMarketSummaryResponse{
		SymbolDenom:        token.SymbolDenom,
		Exponent:           token.Exponent,
		UTokenExchangeRate: rate,
		Supply_APY:         supplyAPY,
		Borrow_APY:         borrowAPY,
		MarketSize:         marketSizeCoin.Amount,
		AvailableBorrow:    availableBorrow,
		Reserved:           reserved,
		Collateral:         collateral,
	}

	if oraclePrice, oracleErr := q.Keeper.TokenPrice(ctx, req.Denom); oracleErr == nil {
		resp.OraclePrice = &oraclePrice
	}

	return &resp, nil
}

func (q Querier) TotalCollateral(
	goCtx context.Context,
	req *types.QueryTotalCollateralRequest,
) (*types.QueryTotalCollateralResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)
	collateral := q.Keeper.GetTotalCollateral(ctx, req.Denom)
	return &types.QueryTotalCollateralResponse{Amount: collateral}, nil
}
