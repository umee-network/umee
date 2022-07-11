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

	if err := s.keeper.Supply(ctx, supplierAddr, msg.Asset); err != nil {
		return nil, err
	}

	s.keeper.Logger(ctx).Debug(
		"assets supplied",
		"supplier", supplierAddr.String(),
		"amount", msg.Asset.String(),
	)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeLoanAsset,
			sdk.NewAttribute(types.EventAttrSupplier, supplierAddr.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Asset.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.EventAttrModule),
			sdk.NewAttribute(sdk.AttributeKeySender, supplierAddr.String()),
		),
	})

	return &types.MsgSupplyResponse{}, nil
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

	if err := s.keeper.Withdraw(ctx, supplierAddr, msg.Asset); err != nil {
		return nil, err
	}

	s.keeper.Logger(ctx).Debug(
		"supplied assets withdrawn",
		"supplier", supplierAddr.String(),
		"amount", msg.Asset.String(),
	)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeWithdraw,
			sdk.NewAttribute(types.EventAttrSupplier, supplierAddr.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Asset.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.EventAttrModule),
			sdk.NewAttribute(sdk.AttributeKeySender, supplierAddr.String()),
		),
	})

	return &types.MsgWithdrawResponse{}, nil
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

	if err := s.keeper.Collateralize(ctx, borrowerAddr, msg.Coin); err != nil {
		return nil, err
	}

	s.keeper.Logger(ctx).Debug(
		"collateral added",
		"borrower", borrowerAddr.String(),
		"amount", msg.Coin.String(),
	)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeCollateralize,
			sdk.NewAttribute(types.EventAttrBorrower, borrowerAddr.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Coin.String()),
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

	if err := s.keeper.Decollateralize(ctx, borrowerAddr, msg.Coin); err != nil {
		return nil, err
	}

	s.keeper.Logger(ctx).Debug(
		"collateral removed",
		"borrower", borrowerAddr.String(),
		"amount", msg.Coin.String(),
	)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeDecollateralize,
			sdk.NewAttribute(types.EventAttrBorrower, borrowerAddr.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Coin.String()),
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
		"assets borrowed",
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

	repaidCoin := sdk.NewCoin(msg.Asset.Denom, repaid)

	s.keeper.Logger(ctx).Debug(
		"borrowed assets repaid",
		"borrower", borrowerAddr.String(),
		"amount", repaidCoin.String(),
		"attempted", msg.Asset.String(),
	)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeRepayBorrowedAsset,
			sdk.NewAttribute(types.EventAttrBorrower, borrowerAddr.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, repaidCoin.String()),
			sdk.NewAttribute(types.EventAttrAttempted, msg.Asset.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.EventAttrModule),
			sdk.NewAttribute(sdk.AttributeKeySender, borrowerAddr.String()),
		),
	})

	return &types.MsgRepayResponse{
		Repaid: repaidCoin,
	}, nil
}

func (s msgServer) Liquidate(
	goCtx context.Context,
	msg *types.MsgLiquidate,
) (*types.MsgLiquidateResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	liquidatorAddr, err := sdk.AccAddressFromBech32(msg.Liquidator)
	if err != nil {
		return nil, err
	}

	borrowerAddr, err := sdk.AccAddressFromBech32(msg.Borrower)
	if err != nil {
		return nil, err
	}

	repaid, reward, err := s.keeper.LiquidateBorrow(ctx, liquidatorAddr, borrowerAddr, msg.Repayment, msg.Reward)
	if err != nil {
		return nil, err
	}

	repaidCoin := sdk.NewCoin(msg.Repayment.Denom, repaid)
	rewardCoin := sdk.NewCoin(msg.Reward.Denom, reward)

	s.keeper.Logger(ctx).Debug(
		"borrowed assets repaid by liquidator",
		"liquidator", liquidatorAddr.String(),
		"borrower", borrowerAddr.String(),
		"amount", repaidCoin.String(),
		"reward", rewardCoin.String(),
		"attempted", msg.Repayment.String(),
	)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeLiquidate,
			sdk.NewAttribute(types.EventAttrLiquidator, liquidatorAddr.String()),
			sdk.NewAttribute(types.EventAttrBorrower, borrowerAddr.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, repaidCoin.String()),
			sdk.NewAttribute(types.EventAttrReward, rewardCoin.String()),
			sdk.NewAttribute(types.EventAttrAttempted, msg.Repayment.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.EventAttrModule),
			sdk.NewAttribute(sdk.AttributeKeySender, liquidatorAddr.String()),
		),
	})

	return &types.MsgLiquidateResponse{
		Repaid: repaidCoin,
		Reward: rewardCoin,
	}, nil
}
