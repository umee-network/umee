package keeper

import (
	"context"
	"strconv"

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

func (s msgServer) SetCollateral(
	goCtx context.Context,
	msg *types.MsgSetCollateral,
) (*types.MsgSetCollateralResponse, error) {

	ctx := sdk.UnwrapSDKContext(goCtx)

	borrowerAddr, err := sdk.AccAddressFromBech32(msg.Borrower)
	if err != nil {
		return nil, err
	}

	if err := s.keeper.SetCollateralSetting(ctx, borrowerAddr, msg.Denom, msg.Enable); err != nil {
		return nil, err
	}

	s.keeper.Logger(ctx).Debug(
		"collateral setting set",
		"borrower", borrowerAddr.String(),
		"denom", msg.Denom,
		"enable", strconv.FormatBool(msg.Enable),
	)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeSetCollateralSetting,
			sdk.NewAttribute(types.EventAttrBorrower, borrowerAddr.String()),
			sdk.NewAttribute(types.EventAttrDenom, msg.Denom),
			sdk.NewAttribute(types.EventAttrEnable, strconv.FormatBool(msg.Enable)),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.EventAttrModule),
			sdk.NewAttribute(sdk.AttributeKeySender, borrowerAddr.String()),
		),
	})

	return &types.MsgSetCollateralResponse{}, nil
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

	if err := s.keeper.BorrowAsset(ctx, borrowerAddr, msg.Amount); err != nil {
		return nil, err
	}

	s.keeper.Logger(ctx).Debug(
		"assets borrowed",
		"borrower", borrowerAddr.String(),
		"amount", msg.Amount.String(),
	)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeBorrowAsset,
			sdk.NewAttribute(types.EventAttrBorrower, borrowerAddr.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.String()),
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

	repaid, err := s.keeper.RepayAsset(ctx, borrowerAddr, msg.Amount)
	if err != nil {
		return nil, err
	}

	repaidCoin := sdk.NewCoin(msg.Amount.Denom, repaid)

	s.keeper.Logger(ctx).Debug(
		"borrowed assets repaid",
		"borrower", borrowerAddr.String(),
		"amount", repaidCoin.String(),
		"attempted", msg.Amount.String(),
	)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeRepayBorrowedAsset,
			sdk.NewAttribute(types.EventAttrBorrower, borrowerAddr.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, repaidCoin.String()),
			sdk.NewAttribute(types.EventAttrAttempted, msg.Amount.String()),
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

	repaid, reward, err := s.keeper.LiquidateBorrow(ctx, liquidatorAddr, borrowerAddr, msg.Repayment, msg.RewardDenom)
	if err != nil {
		return nil, err
	}

	repaidCoin := sdk.NewCoin(msg.Repayment.Denom, repaid)
	rewardCoin := sdk.NewCoin(msg.RewardDenom, reward)

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
