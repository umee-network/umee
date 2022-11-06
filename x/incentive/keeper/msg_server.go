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

func (s msgServer) Bond(
	goCtx context.Context,
	msg *types.MsgBond,
) (*types.MsgBondResponse, error) {
	// TODO: Implement

	return &types.MsgBondResponse{}, types.ErrNotImplemented
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

func (s msgServer) GovSetParams(
	goCtx context.Context,
	msg *types.MsgGovSetParams,
) (*types.MsgGovSetParamsResponse, error) {
	// TODO: Implement

	return &types.MsgGovSetParamsResponse{}, types.ErrNotImplemented
}

func (s msgServer) GovCreateProgram(
	goCtx context.Context,
	msg *types.MsgGovCreateProgram,
) (*types.MsgGovCreateProgramResponse, error) {
	// TODO: Implement

	return &types.MsgGovCreateProgramResponse{}, types.ErrNotImplemented
}

func (s msgServer) GovCreateAndSponsorProgram(
	goCtx context.Context,
	msg *types.MsgGovCreateAndSponsorProgram,
) (*types.MsgGovCreateAndSponsorProgramResponse, error) {
	// TODO: Implement

	return &types.MsgGovCreateAndSponsorProgramResponse{}, types.ErrNotImplemented
}
