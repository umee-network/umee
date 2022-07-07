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

func (s msgServer) WithdrawAsset(
	goCtx context.Context,
	msg *types.MsgWithdrawAsset,
) (*types.MsgWithdrawAssetResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	supplierAddr, err := sdk.AccAddressFromBech32(msg.Supplier)
	if err != nil {
		return nil, err
	}

	if err := s.keeper.WithdrawAsset(ctx, supplierAddr, msg.Asset); err != nil {
		return nil, err
	}

	s.keeper.Logger(ctx).Debug(
		"supplied assets withdrawn",
		"supplier", supplierAddr.String(),
		"amount", msg.Asset.String(),
	)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeWithdrawAsset,
			sdk.NewAttribute(types.EventAttrSupplier, supplierAddr.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Asset.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.EventAttrModule),
			sdk.NewAttribute(sdk.AttributeKeySender, supplierAddr.String()),
		),
	})

	return &types.MsgWithdrawAssetResponse{}, nil
}

func (s msgServer) AddCollateral(
	goCtx context.Context,
	msg *types.MsgAddCollateral,
) (*types.MsgAddCollateralResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	borrowerAddr, err := sdk.AccAddressFromBech32(msg.Borrower)
	if err != nil {
		return nil, err
	}

	if err := s.keeper.AddCollateral(ctx, borrowerAddr, msg.Coin); err != nil {
		return nil, err
	}

	s.keeper.Logger(ctx).Debug(
		"collateral added",
		"borrower", borrowerAddr.String(),
		"amount", msg.Coin.String(),
	)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeAddCollateral,
			sdk.NewAttribute(types.EventAttrBorrower, borrowerAddr.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Coin.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.EventAttrModule),
			sdk.NewAttribute(sdk.AttributeKeySender, borrowerAddr.String()),
		),
	})

	return &types.MsgAddCollateralResponse{}, nil
}

func (s msgServer) RemoveCollateral(
	goCtx context.Context,
	msg *types.MsgRemoveCollateral,
) (*types.MsgRemoveCollateralResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	borrowerAddr, err := sdk.AccAddressFromBech32(msg.Borrower)
	if err != nil {
		return nil, err
	}

	if err := s.keeper.RemoveCollateral(ctx, borrowerAddr, msg.Coin); err != nil {
		return nil, err
	}

	s.keeper.Logger(ctx).Debug(
		"collateral removed",
		"borrower", borrowerAddr.String(),
		"amount", msg.Coin.String(),
	)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeRemoveCollateral,
			sdk.NewAttribute(types.EventAttrBorrower, borrowerAddr.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Coin.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.EventAttrModule),
			sdk.NewAttribute(sdk.AttributeKeySender, borrowerAddr.String()),
		),
	})

	return &types.MsgRemoveCollateralResponse{}, nil
}

func (s msgServer) BorrowAsset(
	goCtx context.Context,
	msg *types.MsgBorrowAsset,
) (*types.MsgBorrowAssetResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	borrowerAddr, err := sdk.AccAddressFromBech32(msg.Borrower)
	if err != nil {
		return nil, err
	}

	if err := s.keeper.BorrowAsset(ctx, borrowerAddr, msg.Asset); err != nil {
		return nil, err
	}

	s.keeper.Logger(ctx).Debug(
		"assets borrowed",
		"borrower", borrowerAddr.String(),
		"amount", msg.Asset.String(),
	)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeBorrowAsset,
			sdk.NewAttribute(types.EventAttrBorrower, borrowerAddr.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Asset.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.EventAttrModule),
			sdk.NewAttribute(sdk.AttributeKeySender, borrowerAddr.String()),
		),
	})

	return &types.MsgBorrowAssetResponse{}, nil
}

func (s msgServer) RepayAsset(
	goCtx context.Context,
	msg *types.MsgRepayAsset,
) (*types.MsgRepayAssetResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	borrowerAddr, err := sdk.AccAddressFromBech32(msg.Borrower)
	if err != nil {
		return nil, err
	}

	repaid, err := s.keeper.RepayAsset(ctx, borrowerAddr, msg.Asset)
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

	return &types.MsgRepayAssetResponse{
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

	repaid, collateral, reward, err := s.keeper.LiquidateBorrow(ctx, liquidatorAddr, borrowerAddr, msg.Repayment, msg.RewardDenom)
	if err != nil {
		return nil, err
	}

	s.keeper.Logger(ctx).Debug(
		"borrowed assets repaid by liquidator",
		"liquidator", liquidatorAddr.String(),
		"borrower", borrowerAddr.String(),
		"attempted", msg.Repayment.String(),
		"repaid", repaid.String(),
		"collateral", collateral.String(),
		"reward", reward.String(),
	)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeLiquidate,
			sdk.NewAttribute(types.EventAttrLiquidator, liquidatorAddr.String()),
			sdk.NewAttribute(types.EventAttrBorrower, borrowerAddr.String()),
			sdk.NewAttribute(types.EventAttrAttempted, msg.Repayment.String()),
			sdk.NewAttribute(types.EventAttrRepaid, reward.String()),
			sdk.NewAttribute(types.EventAttrCollateral, collateral.String()),
			sdk.NewAttribute(types.EventAttrReward, reward.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.EventAttrModule),
			sdk.NewAttribute(sdk.AttributeKeySender, liquidatorAddr.String()),
		),
	})

	return &types.MsgLiquidateResponse{
		Repaid:     repaid,
		Collateral: collateral,
		Reward:     reward,
	}, nil
}
