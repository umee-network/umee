package keeper

import (
	"context"

	"cosmossdk.io/store/prefix"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/query"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/umee-network/umee/v6/x/leverage/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"
)

func (q Querier) accountSummary(ctx sdk.Context, addr sdk.AccAddress) (*types.QueryAccountSummaryResponse, error) {
	supplied, err := q.GetAllSupplied(ctx, addr)
	if err != nil {
		return nil, err
	}
	collateral := q.GetBorrowerCollateral(ctx, addr)
	borrowed := q.GetBorrowerBorrows(ctx, addr)

	// the following price calculations use the most recent prices if spot prices are missing
	lastSuppliedValue, err := q.VisibleTokenValue(ctx, supplied, types.PriceModeQuery)
	if err != nil {
		return nil, err
	}
	lastBorrowedValue, err := q.VisibleTokenValue(ctx, borrowed, types.PriceModeQuery)
	if err != nil {
		return nil, err
	}
	lastCollateralValue, err := q.VisibleCollateralValue(ctx, collateral, types.PriceModeQuery)
	if err != nil {
		return nil, err
	}

	// these use leverage-like prices: the lower of spot or historic price for supplied tokens and higher for borrowed.
	// unlike transactions, this query will use expired prices instead of skipping them.
	suppliedValue, err := q.VisibleTokenValue(ctx, supplied, types.PriceModeQueryLow)
	if err != nil {
		return nil, err
	}
	collateralValue, err := q.VisibleCollateralValue(ctx, collateral, types.PriceModeQueryLow)
	if err != nil {
		return nil, err
	}
	borrowedValue, err := q.VisibleTokenValue(ctx, borrowed, types.PriceModeQueryHigh)
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
	ap, err := q.GetAccountPosition(ctx, addr, false)
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
	ap, err = q.GetAccountPosition(ctx, addr, true)
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

// AccountSummaries implements types.QueryServer.
func (q Querier) AccountSummaries(goCtx context.Context, req *types.QueryAccountSummaries) (
	*types.QueryAccountSummariesResponse, error,
) {
	if req == nil {
		return nil, status.Error(codes.InvalidArgument, "empty request")
	}
	ctx := sdk.UnwrapSDKContext(goCtx)
	// get the all accounts
	store := ctx.KVStore(q.akStoreKey)
	accountsStore := prefix.NewStore(store, authtypes.AddressStoreKeyPrefix)

	var accounts []*types.AccountSummary
	pageRes, err := query.Paginate(accountsStore, req.Pagination, func(key, value []byte) error {
		var acc sdk.AccountI
		if err := q.cdc.UnmarshalInterface(value, &acc); err != nil {
			return err
		}
		accSummary, err := q.accountSummary(ctx, acc.GetAddress())
		if err != nil {
			return err
		}
		accounts = append(accounts, &types.AccountSummary{
			Address:        acc.GetAddress().String(),
			AccountSummary: accSummary,
		})
		return nil
	})

	return &types.QueryAccountSummariesResponse{AccountSummaries: accounts, Pagination: pageRes}, err
}

// AccountSummary implements types.QueryServer.
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
	return q.accountSummary(ctx, addr)
}
