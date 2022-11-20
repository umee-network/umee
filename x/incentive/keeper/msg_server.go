package keeper

import (
	"context"

	"github.com/umee-network/umee/v3/x/incentive"
)

var _ incentive.MsgServer = msgServer{}

type msgServer struct {
	keeper Keeper
}

// NewMsgServerImpl returns an implementation of MsgServer for the x/incentive
// module.
func NewMsgServerImpl(keeper Keeper) incentive.MsgServer {
	return &msgServer{keeper: keeper}
}

func (s msgServer) Claim(
	goCtx context.Context,
	msg *incentive.MsgClaim,
) (*incentive.MsgClaimResponse, error) {
	// TODO: Implement

	return &incentive.MsgClaimResponse{}, incentive.ErrNotImplemented
}

func (s msgServer) Bond(
	goCtx context.Context,
	msg *incentive.MsgBond,
) (*incentive.MsgBondResponse, error) {
	// TODO: Implement

	return &incentive.MsgBondResponse{}, incentive.ErrNotImplemented
}

func (s msgServer) BeginUnbonding(
	goCtx context.Context,
	msg *incentive.MsgBeginUnbonding,
) (*incentive.MsgBeginUnbondingResponse, error) {
	// TODO: Implement

	return &incentive.MsgBeginUnbondingResponse{}, incentive.ErrNotImplemented
}

func (s msgServer) Sponsor(
	goCtx context.Context,
	msg *incentive.MsgSponsor,
) (*incentive.MsgSponsorResponse, error) {
	// TODO: Implement

	return &incentive.MsgSponsorResponse{}, incentive.ErrNotImplemented
}

func (s msgServer) GovSetParams(
	goCtx context.Context,
	msg *incentive.MsgGovSetParams,
) (*incentive.MsgGovSetParamsResponse, error) {
	// TODO: Implement

	return &incentive.MsgGovSetParamsResponse{}, incentive.ErrNotImplemented
}

func (s msgServer) GovCreateProgram(
	goCtx context.Context,
	msg *incentive.MsgGovCreateProgram,
) (*incentive.MsgGovCreateProgramResponse, error) {
	// TODO: Implement

	return &incentive.MsgGovCreateProgramResponse{}, incentive.ErrNotImplemented
}

func (s msgServer) GovCreateAndSponsorProgram(
	goCtx context.Context,
	msg *incentive.MsgGovCreateAndSponsorProgram,
) (*incentive.MsgGovCreateAndSponsorProgramResponse, error) {
	// TODO: Implement

	return &incentive.MsgGovCreateAndSponsorProgramResponse{}, incentive.ErrNotImplemented
}
