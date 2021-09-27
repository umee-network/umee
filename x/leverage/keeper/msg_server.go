package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

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

	// TODO: Events + Logging

	return nil, status.Errorf(codes.Unimplemented, "method LendAsset not implemented")
}

func (s msgServer) WithdrawAsset(
	goCtx context.Context,
	req *types.MsgWithdrawAsset,
) (*types.MsgWithdrawAssetResponse, error) {

	// ctx := sdk.UnwrapSDKContext(goCtx)
	// TODO: Implement...

	return nil, status.Errorf(codes.Unimplemented, "method WithdrawAsset not implemented")
}
