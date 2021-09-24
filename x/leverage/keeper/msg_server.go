package keeper

import (
	"context"

	"google.golang.org/grpc/codes"
	"google.golang.org/grpc/status"

	"github.com/umee-network/umee/x/leverage/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	Keeper
}

// NewMsgServerImpl returns an implementation of MsgServer for the x/leverage
// module.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{Keeper: keeper}
}

func (s msgServer) LendAsset(
	goCtx context.Context,
	req *types.MsgLendAsset,
) (*types.MsgLendAssetResponse, error) {

	// ctx := sdk.UnwrapSDKContext(goCtx)
	// TODO: Implement...

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
