package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v2/x/leverage/types"
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

func (s msgServer) Supply(
	goCtx context.Context,
	msg *types.MsgSupply,
) (*types.MsgSupplyResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	supplierAddr, err := sdk.AccAddressFromBech32(msg.Supplier)
	if err != nil {
		return nil, err
	}

	received, err := s.keeper.Supply(ctx, supplierAddr, msg.Asset)
	if err != nil {
		return nil, err
	}

	s.keeper.Logger(ctx).Debug(
		"assets supplied",
		"supplier", supplierAddr.String(),
		"supplied", msg.Asset.String(),
		"received", received.String(),
	)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeSupply,
			sdk.NewAttribute(types.EventAttrSupplier, supplierAddr.String()),
			sdk.NewAttribute(types.EventAttrSupplied, msg.Asset.String()),
			sdk.NewAttribute(types.EventAttrReceived, received.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.EventAttrModule),
			sdk.NewAttribute(sdk.AttributeKeySender, supplierAddr.String()),
		),
	})

	return &types.MsgSupplyResponse{
		Received: received,
	}, nil
}

func (s msgServer) Withdraw(
	goCtx context.Context,
	msg *types.MsgWithdraw,
) (*types.MsgWithdrawResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	supplierAddr, err := sdk.AccAddressFromBech32(msg.Supplier)
	if err != nil {
		return nil, err
	}

	received, err := s.keeper.Withdraw(ctx, supplierAddr, msg.Asset)
	if err != nil {
		return nil, err
	}

	s.keeper.Logger(ctx).Debug(
		"supplied assets withdrawn",
		"supplier", supplierAddr.String(),
		"redeemed", msg.Asset.String(),
		"received", received.String(),
	)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeWithdraw,
			sdk.NewAttribute(types.EventAttrSupplier, supplierAddr.String()),
			sdk.NewAttribute(types.EventAttrRedeemed, msg.Asset.String()),
			sdk.NewAttribute(types.EventAttrReceived, received.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.EventAttrModule),
			sdk.NewAttribute(sdk.AttributeKeySender, supplierAddr.String()),
		),
	})

	return &types.MsgWithdrawResponse{
		Received: received,
	}, nil
}

func (s msgServer) Collateralize(
	goCtx context.Context,
	msg *types.MsgCollateralize,
) (*types.MsgCollateralizeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	borrowerAddr, err := sdk.AccAddressFromBech32(msg.Borrower)
	if err != nil {
		return nil, err
	}

	if err := s.keeper.Collateralize(ctx, borrowerAddr, msg.Asset); err != nil {
		return nil, err
	}

	s.keeper.Logger(ctx).Debug(
		"collateral added",
		"borrower", borrowerAddr.String(),
		"amount", msg.Asset.String(),
	)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCollateralize,
			sdk.NewAttribute(types.EventAttrBorrower, borrowerAddr.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Asset.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.EventAttrModule),
			sdk.NewAttribute(sdk.AttributeKeySender, borrowerAddr.String()),
		),
	})

	return &types.MsgCollateralizeResponse{}, nil
}

func (s msgServer) Decollateralize(
	goCtx context.Context,
	msg *types.MsgDecollateralize,
) (*types.MsgDecollateralizeResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	borrowerAddr, err := sdk.AccAddressFromBech32(msg.Borrower)
	if err != nil {
		return nil, err
	}

	if err := s.keeper.Decollateralize(ctx, borrowerAddr, msg.Asset); err != nil {
		return nil, err
	}

	s.keeper.Logger(ctx).Debug(
		"collateral removed",
		"borrower", borrowerAddr.String(),
		"amount", msg.Asset.String(),
	)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeDecollateralize,
			sdk.NewAttribute(types.EventAttrBorrower, borrowerAddr.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Asset.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.EventAttrModule),
			sdk.NewAttribute(sdk.AttributeKeySender, borrowerAddr.String()),
		),
	})

	return &types.MsgDecollateralizeResponse{}, nil
}

func (s msgServer) Borrow(
	goCtx context.Context,
	msg *types.MsgBorrow,
) (*types.MsgBorrowResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	borrowerAddr, err := sdk.AccAddressFromBech32(msg.Borrower)
	if err != nil {
		return nil, err
	}

	if err := s.keeper.Borrow(ctx, borrowerAddr, msg.Asset); err != nil {
		return nil, err
	}

	s.keeper.Logger(ctx).Debug(
		"coins borrowed",
		"borrower", borrowerAddr.String(),
		"amount", msg.Asset.String(),
	)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeBorrow,
			sdk.NewAttribute(types.EventAttrBorrower, borrowerAddr.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Asset.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.EventAttrModule),
			sdk.NewAttribute(sdk.AttributeKeySender, borrowerAddr.String()),
		),
	})

	return &types.MsgBorrowResponse{}, nil
}

func (s msgServer) Repay(
	goCtx context.Context,
	msg *types.MsgRepay,
) (*types.MsgRepayResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	borrowerAddr, err := sdk.AccAddressFromBech32(msg.Borrower)
	if err != nil {
		return nil, err
	}

	repaid, err := s.keeper.Repay(ctx, borrowerAddr, msg.Asset)
	if err != nil {
		return nil, err
	}

	s.keeper.Logger(ctx).Debug(
		"borrowed coins repaid",
		"borrower", borrowerAddr.String(),
		"attempted", msg.Asset.String(),
		"repaid", repaid.String(),
	)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeRepay,
			sdk.NewAttribute(types.EventAttrBorrower, borrowerAddr.String()),
			sdk.NewAttribute(types.EventAttrAttempted, msg.Asset.String()),
			sdk.NewAttribute(types.EventAttrRepaid, repaid.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.EventAttrModule),
			sdk.NewAttribute(sdk.AttributeKeySender, borrowerAddr.String()),
		),
	})

	return &types.MsgRepayResponse{
		Repaid: repaid,
	}, nil
}

func (s msgServer) Liquidate(
	goCtx context.Context,
	msg *types.MsgLiquidate,
) (*types.MsgLiquidateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	liquidator, err := sdk.AccAddressFromBech32(msg.Liquidator)
	if err != nil {
		return nil, err
	}

	borrower, err := sdk.AccAddressFromBech32(msg.Borrower)
	if err != nil {
		return nil, err
	}

	repaid, liquidated, reward, err := s.keeper.Liquidate(ctx, liquidator, borrower, msg.Repayment, msg.RewardDenom)
	if err != nil {
		return nil, err
	}

	s.keeper.Logger(ctx).Debug(
		"borrowed coins repaid by liquidator",
		"liquidator", liquidator.String(),
		"borrower", borrower.String(),
		"attempted", msg.Repayment.String(),
		"repaid", repaid.String(),
		"liquidated", liquidated.String(),
		"reward", reward.String(),
	)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeLiquidate,
			sdk.NewAttribute(types.EventAttrLiquidator, liquidator.String()),
			sdk.NewAttribute(types.EventAttrBorrower, borrower.String()),
			sdk.NewAttribute(types.EventAttrAttempted, msg.Repayment.String()),
			sdk.NewAttribute(types.EventAttrRepaid, reward.String()),
			sdk.NewAttribute(types.EventAttrLiquidated, liquidated.String()),
			sdk.NewAttribute(types.EventAttrReward, reward.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.EventAttrModule),
			sdk.NewAttribute(sdk.AttributeKeySender, liquidator.String()),
		),
	})

	return &types.MsgLiquidateResponse{
		Repaid:     repaid,
		Collateral: liquidated,
		Reward:     reward,
	}, nil
}
