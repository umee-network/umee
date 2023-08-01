package query

import (
	"context"

	"github.com/gogo/protobuf/proto"
	"github.com/umee-network/umee/v5/x/metoken"
)

// HandleMeTokenParams handles the get the x/metoken module's parameters.
func (q UmeeQuery) HandleMeTokenParams(
	ctx context.Context,
	qs metoken.QueryServer,
) (proto.Message, error) {
	return qs.Params(ctx, &metoken.QueryParams{})
}

// HandleMeTokenIndexes handles the get the x/metoken indexes.
func (q UmeeQuery) HandleMeTokenIndexes(
	ctx context.Context,
	qs metoken.QueryServer,
) (proto.Message, error) {
	req := metoken.QueryIndexes{MetokenDenom: q.Indexes.MetokenDenom}
	return qs.Indexes(ctx, &req)
}

// HandleMeTokenSwapFee handles the get the x/metoken swap fee.
func (q UmeeQuery) HandleMeTokenSwapFee(
	ctx context.Context,
	qs metoken.QueryServer,
) (proto.Message, error) {
	req := metoken.QuerySwapFee{Asset: q.SwapFee.Asset, MetokenDenom: q.SwapFee.MetokenDenom}
	return qs.SwapFee(ctx, &req)
}

// HandleMeTokenRedeemFee handles the get the x/metoken redeem fee.
func (q UmeeQuery) HandleMeTokenRedeemFee(
	ctx context.Context,
	qs metoken.QueryServer,
) (proto.Message, error) {
	req := metoken.QueryRedeemFee{AssetDenom: q.RedeemFee.AssetDenom, Metoken: q.RedeemFee.Metoken}
	return qs.RedeemFee(ctx, &req)
}

// HandleMeTokenIndexBalances handles the get the x/metoken indexes balances.
func (q UmeeQuery) HandleMeTokenIndexBalances(
	ctx context.Context,
	qs metoken.QueryServer,
) (proto.Message, error) {
	req := metoken.QueryIndexBalances{MetokenDenom: q.IndexBalances.MetokenDenom}
	return qs.IndexBalances(ctx, &req)
}

// HandleMeTokenIndexPrice handles the get the x/metoken indexe price.
func (q UmeeQuery) HandleMeTokenIndexPrice(
	ctx context.Context,
	qs metoken.QueryServer,
) (proto.Message, error) {
	req := metoken.QueryIndexPrice{MetokenDenom: q.IndexPrice.MetokenDenom}
	return qs.IndexPrice(ctx, &req)
}
