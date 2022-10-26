package keeper

import (
	"context"

	"github.com/umee-network/umee/v3/x/incentive/types"
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

	return &types.MsgClaimResponse{}, types.ErrNotImplemented
}

func (s msgServer) Lock(
	goCtx context.Context,
	msg *types.MsgLock,
) (*types.MsgLockResponse, error) {
	// TODO: Implement

	return &types.MsgLockResponse{}, types.ErrNotImplemented
}

func (s msgServer) BeginUnbonding(
	goCtx context.Context,
	msg *types.MsgBeginUnbonding,
) (*types.MsgBeginUnbondingResponse, error) {
	// TODO: Implement

	return &types.MsgBeginUnbondingResponse{}, types.ErrNotImplemented
}

func (s msgServer) Sponsor(
	goCtx context.Context,
	msg *types.MsgSponsor,
) (*types.MsgSponsorResponse, error) {
	// TODO: Implement

	return &types.MsgSponsorResponse{}, types.ErrNotImplemented
}

func (s msgServer) CreateProgram(
	goCtx context.Context,
	msg *types.MsgCreateProgram,
) (*types.MsgCreateProgramResponse, error) {
	// TODO: Implement

	return &types.MsgCreateProgramResponse{}, types.ErrNotImplemented
}
