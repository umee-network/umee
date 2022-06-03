package keeper

import (
	"context"

	"github.com/umee-network/umee/v2/x/incentive/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	keeper Keeper
}

// NewMsgServerImpl returns an implementation of MsgServer for the x/incentive
// module.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{keeper: keeper}
}

func (s msgServer) Claim(
	goCtx context.Context,
	msg *types.MsgClaim,
) (*types.MsgClaimResponse, error) {
	// TODO: Implement

	return &types.MsgClaimResponse{}, nil
}

func (s msgServer) Lock(
	goCtx context.Context,
	msg *types.MsgLock,
) (*types.MsgLockResponse, error) {
	// TODO: Implement

	return &types.MsgLockResponse{}, nil
}

func (s msgServer) Unlock(
	goCtx context.Context,
	msg *types.MsgUnlock,
) (*types.MsgUnlockResponse, error) {
	// TODO: Implement

	return &types.MsgUnlockResponse{}, nil
}

func (s msgServer) Sponsor(
	goCtx context.Context,
	msg *types.MsgSponsor,
) (*types.MsgSponsorResponse, error) {
	// TODO: Implement

	return &types.MsgSponsorResponse{}, nil
}
