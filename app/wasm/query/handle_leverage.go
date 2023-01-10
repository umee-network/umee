package query

import (
	"context"

	"github.com/gogo/protobuf/proto"

	lvtypes "github.com/umee-network/umee/v4/x/leverage/types"
)

// HandleLeverageParams handles the get the x/leverage module's parameters.
func (q UmeeQuery) HandleLeverageParams(
	ctx context.Context,
	qs lvtypes.QueryServer,
) (proto.Message, error) {
	return qs.Params(ctx, &lvtypes.QueryParams{})
}

// HandleRegisteredTokens handles the get all registered tokens query and response.
func (q UmeeQuery) HandleRegisteredTokens(
	ctx context.Context,
	qs lvtypes.QueryServer,
) (proto.Message, error) {
	return qs.RegisteredTokens(ctx, &lvtypes.QueryRegisteredTokens{})
}

// HandleMarketSummary queries a base asset's current borrowing and supplying conditions.
func (q UmeeQuery) HandleMarketSummary(
	ctx context.Context,
	qs lvtypes.QueryServer,
) (proto.Message, error) {
	return qs.MarketSummary(ctx, &lvtypes.QueryMarketSummary{Denom: q.MarketSummary.Denom})
}

// HandleAccountBalances queries an account's current supply, collateral, and borrow positions.
func (q UmeeQuery) HandleAccountBalances(
	ctx context.Context,
	qs lvtypes.QueryServer,
) (proto.Message, error) {
	req := &lvtypes.QueryAccountBalances{Address: q.AccountBalances.Address}
	return qs.AccountBalances(ctx, req)
}

// HandleAccountSummary queries USD values representing an account's total
// positions and borrowing limits. It requires oracle prices to return successfully.
func (q UmeeQuery) HandleAccountSummary(
	ctx context.Context,
	qs lvtypes.QueryServer,
) (proto.Message, error) {
	req := &lvtypes.QueryAccountSummary{Address: q.AccountSummary.Address}
	return qs.AccountSummary(ctx, req)
}

// HandleLiquidationTargets queries a list of all borrower account addresses eligible for liquidation.
func (q UmeeQuery) HandleLiquidationTargets(
	ctx context.Context,
	qs lvtypes.QueryServer,
) (proto.Message, error) {
	return qs.LiquidationTargets(ctx, &lvtypes.QueryLiquidationTargets{})
}

// HandleBadDebts queries bad debts.
func (q UmeeQuery) HandleBadDebts(
	ctx context.Context,
	qs lvtypes.QueryServer,
) (proto.Message, error) {
	return qs.BadDebts(ctx, &lvtypes.QueryBadDebts{})
}

// HandleBadDebts queries bad debts.
func (q UmeeQuery) HandleMaxWithdraw(
	ctx context.Context,
	qs lvtypes.QueryServer,
) (proto.Message, error) {
	req := &lvtypes.QueryMaxWithdraw{Address: q.MaxWithdraw.Address, Denom: q.MaxWithdraw.Denom}
	return qs.MaxWithdraw(ctx, req)
}
