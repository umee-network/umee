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

func (q Querier) Loaned(
	goCtx context.Context,
	req *types.QueryLoanedRequest,
) (*types.QueryLoanedResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid address")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	lender, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}

	if len(req.Denom) == 0 {
		tokens, err := q.Keeper.GetLenderLoaned(ctx, lender)
		if err != nil {
			return nil, err
		}

		return &types.QueryLoanedResponse{Loaned: tokens}, nil
	}

	if !q.Keeper.IsAcceptedToken(ctx, req.Denom) {
		return nil, status.Error(codes.InvalidArgument, "not accepted Token denom")
	}

	token, err := q.Keeper.GetLoaned(ctx, lender, req.Denom)
	if err != nil {
		return nil, err
	}

	return &types.QueryLoanedResponse{Loaned: sdk.NewCoins(token)}, nil
}

func (q Querier) LoanedValue(
	goCtx context.Context,
	req *types.QueryLoanedValueRequest,
) (*types.QueryLoanedValueResponse, error) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	if req.Address == "" {
		return nil, status.Error(codes.InvalidArgument, "invalid address")
	}

	ctx := sdk.UnwrapSDKContext(goCtx)

	lender, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}

	var tokens sdk.Coins

	if len(req.Denom) == 0 {
		tokens, err = q.Keeper.GetLenderLoaned(ctx, lender)
		if err != nil {
			return nil, err
		}
	} else {
		if !q.Keeper.IsAcceptedToken(ctx, req.Denom) {
			return nil, status.Error(codes.InvalidArgument, "not accepted Token denom")
		}

		loaned, err := q.Keeper.GetLoaned(ctx, lender, req.Denom)
		if err != nil {
			return nil, err
		}

		tokens = sdk.NewCoins(loaned)
	}

	value, err := q.Keeper.TotalTokenValue(ctx, tokens)
	if err != nil {
		return nil, err
	}

	return &types.QueryLoanedValueResponse{LoanedValue: value}, nil
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

	lendAPY := q.Keeper.DeriveLendAPY(ctx, req.Denom)

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

	marketSizeCoin, err := q.Keeper.GetTotalLoaned(ctx, req.Denom)
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

	marketSizeCoin, err := q.Keeper.GetTotalLoaned(ctx, req.Denom)
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

	lender, err := sdk.AccAddressFromBech32(req.Address)
	if err != nil {
		return nil, err
	}

	var uTokens sdk.Coins

	if len(req.Denom) == 0 {
		uTokens = q.Keeper.GetBorrowerCollateral(ctx, lender)
	} else {
		if !q.Keeper.IsAcceptedUToken(ctx, req.Denom) {
			return nil, status.Error(codes.InvalidArgument, "not accepted uToken denom")
		}

		collateral := q.Keeper.GetCollateralAmount(ctx, lender, req.Denom)

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

func (q Querier) LiquidationLimit(
	goCtx context.Context,
	req *types.QueryLiquidationLimitRequest,
) (*types.QueryLiquidationLimitResponse, error) {
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

	limit, err := q.Keeper.CalculateLiquidationLimit(ctx, collateral)
	if err != nil {
		return nil, err
	}

	return &types.QueryLiquidationLimitResponse{LiquidationLimit: limit}, nil
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

	token, err := q.Keeper.GetRegisteredToken(ctx, req.Denom)
	if err != nil {
		return nil, status.Error(codes.InvalidArgument, "not accepted Token denom")
	}
	rate := q.Keeper.DeriveExchangeRate(ctx, req.Denom)
	lendAPY := q.Keeper.DeriveLendAPY(ctx, req.Denom)
	borrowAPY := q.Keeper.DeriveBorrowAPY(ctx, req.Denom)
	marketSizeCoin, _ := q.Keeper.GetTotalLoaned(ctx, req.Denom)
	availableBorrow := q.Keeper.GetAvailableToBorrow(ctx, req.Denom)
	reserved := q.Keeper.GetReserveAmount(ctx, req.Denom)

	resp := types.QueryMarketSummaryResponse{
		SymbolDenom:        token.SymbolDenom,
		Exponent:           token.Exponent,
		UTokenExchangeRate: rate,
		Lend_APY:           lendAPY,
		Borrow_APY:         borrowAPY,
		MarketSize:         marketSizeCoin.Amount,
		AvailableBorrow:    availableBorrow,
		Reserved:           reserved,
	}

	if oraclePrice, oracleErr := q.Keeper.TokenPrice(ctx, req.Denom); oracleErr == nil {
		resp.OraclePrice = &oraclePrice
	}

	return &resp, nil
}
