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

	if err := s.keeper.RepayAsset(ctx, borrowerAddr, msg.Amount); err != nil {
		return nil, err
	}

	s.keeper.Logger(ctx).Debug(
		"borrowed assets repaid",
		"borrower", borrowerAddr.String(),
		"amount", msg.Amount.String(),
	)

	ctx.EventManager().EmitEvents(sdk.Events{
		sdk.NewEvent(
			types.EventTypeRepayBorrowedAsset,
			sdk.NewAttribute(types.EventAttrBorrower, borrowerAddr.String()),
			sdk.NewAttribute(sdk.AttributeKeyAmount, msg.Amount.String()),
		),
		sdk.NewEvent(
			sdk.EventTypeMessage,
			sdk.NewAttribute(sdk.AttributeKeyModule, types.EventAttrModule),
			sdk.NewAttribute(sdk.AttributeKeySender, borrowerAddr.String()),
		),
	})

	return &types.MsgRepayAssetResponse{}, nil
}
