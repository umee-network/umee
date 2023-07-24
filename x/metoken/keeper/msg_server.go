package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v5/util/sdkutil"

	"github.com/umee-network/umee/v5/x/metoken"
)

var _ metoken.MsgServer = msgServer{}

type msgServer struct {
	kb Builder
}

// NewMsgServerImpl returns an implementation of MsgServer for the x/metoken
// module.
func NewMsgServerImpl(kb Builder) metoken.MsgServer {
	return &msgServer{kb: kb}
}

// Swap handles the request for the swap, delegates the execution and returns the response.
func (m msgServer) Swap(goCtx context.Context, msg *metoken.MsgSwap) (*metoken.MsgSwapResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k := m.kb.Keeper(&ctx)

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	userAddr, err := sdk.AccAddressFromBech32(msg.User)
	if err != nil {
		return nil, err
	}

	resp, err := k.swap(userAddr, msg.MetokenDenom, msg.Asset)
	if err != nil {
		return nil, err
	}

	k.Logger().Debug(
		"swap executed",
		"user", userAddr,
		"meTokens", resp.meTokens.String(),
		"calculateFee", resp.fee.String(),
		"reserved", resp.reserved.String(),
		"leveraged", resp.leveraged.String(),
	)

	sdkutil.Emit(
		&ctx, &metoken.EventSwap{
			Recipient: msg.User,
			Asset:     sdk.NewCoin(msg.Asset.Denom, resp.reserved.Amount.Add(resp.leveraged.Amount)),
			Metoken:   resp.meTokens,
			Fee:       resp.fee,
		},
	)

	return &metoken.MsgSwapResponse{
		Fee:      resp.fee,
		Returned: resp.meTokens,
	}, nil
}

// Redeem handles the request for the redemption, delegates the execution and returns the response.
func (m msgServer) Redeem(goCtx context.Context, msg *metoken.MsgRedeem) (*metoken.MsgRedeemResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	k := m.kb.Keeper(&ctx)

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	userAddr, err := sdk.AccAddressFromBech32(msg.User)
	if err != nil {
		return nil, err
	}

	resp, err := k.redeem(userAddr, msg.Metoken, msg.AssetDenom)
	if err != nil {
		return nil, err
	}

	k.Logger().Debug(
		"redeem executed",
		"user", userAddr,
		"calculateFee", resp.fee.String(),
		"from_reserves", resp.fromReserves.String(),
		"from_leverage", resp.fromLeverage.String(),
		"burned", msg.Metoken.String(),
	)

	totalRedeemed := sdk.NewCoin(
		msg.AssetDenom,
		resp.fromReserves.Amount.Add(resp.fromLeverage.Amount).Sub(resp.fee.Amount),
	)
	sdkutil.Emit(
		&ctx, &metoken.EventRedeem{
			Recipient: msg.User,
			Metoken:   msg.Metoken,
			Asset:     totalRedeemed,
			Fee:       resp.fee,
		},
	)

	return &metoken.MsgRedeemResponse{
		Returned: totalRedeemed,
		Fee:      resp.fee,
	}, nil
}

// GovSetParams handles the request for updating Params.
func (m msgServer) GovSetParams(goCtx context.Context, msg *metoken.MsgGovSetParams) (
	*metoken.MsgGovSetParamsResponse,
	error,
) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	if err := m.kb.Keeper(&ctx).SetParams(msg.Params); err != nil {
		return nil, err
	}

	return &metoken.MsgGovSetParamsResponse{}, nil
}

// GovUpdateRegistry handles the request for updating the indexes' registry.
func (m msgServer) GovUpdateRegistry(
	goCtx context.Context,
	msg *metoken.MsgGovUpdateRegistry,
) (*metoken.MsgGovUpdateRegistryResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := msg.ValidateBasic(); err != nil {
		return nil, err
	}

	if err := m.kb.Keeper(&ctx).UpdateIndexes(msg.AddIndex, msg.UpdateIndex); err != nil {
		return nil, err
	}

	return &metoken.MsgGovUpdateRegistryResponse{}, nil
}
