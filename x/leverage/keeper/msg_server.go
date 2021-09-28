package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/x/leverage/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	keeper Keeper
}

// NewMsgServerImpl returns an implementation of MsgServer for the x/leverage
// module.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{keeper: keeper}
}

func (s msgServer) LendAsset(
	goCtx context.Context,
	msg *types.MsgLendAsset,
) (*types.MsgLendAssetResponse, error) {

	ctx := sdk.UnwrapSDKContext(goCtx)

	lenderAddr, err := sdk.AccAddressFromBech32(msg.Lender)
	if err != nil {
		return nil, err
	}

	if err := s.keeper.LendAsset(ctx, lenderAddr, msg.Amount); err != nil {
		return nil, err
	}

	s.keeper.Logger(ctx).Debug(
		"assets loaned",
		"lender", lenderAddr.String(),
		"amount", msg.Amount.String(),
	)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeLoanAsset,
			sdk.NewAttribute(types.EventAttrLender, lenderAddr.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.EventAttrModule),
			sdk.NewAttribute(sdk.AttributeKeySender, lenderAddr.String()),
		),
	})

	return &types.MsgLendAssetResponse{}, nil
}

func (s msgServer) WithdrawAsset(
	goCtx context.Context,
	msg *types.MsgWithdrawAsset,
) (*types.MsgWithdrawAssetResponse, error) {

	ctx := sdk.UnwrapSDKContext(goCtx)

	lenderAddr, err := sdk.AccAddressFromBech32(msg.Lender)
	if err != nil {
		return nil, err
	}

	if err := s.keeper.WithdrawAsset(ctx, lenderAddr, msg.Amount); err != nil {
		return nil, err
	}

	s.keeper.Logger(ctx).Debug(
		"loaned assets withdrawn",
		"lender", lenderAddr.String(),
		"amount", msg.Amount.String(),
	)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeWithdrawLoanedAsset,
			sdk.NewAttribute(types.EventAttrLender, lenderAddr.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.EventAttrModule),
			sdk.NewAttribute(sdk.AttributeKeySender, lenderAddr.String()),
		),
	})

	return &types.MsgWithdrawAssetResponse{}, nil
}
