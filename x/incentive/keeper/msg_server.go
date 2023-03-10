package keeper

import (
	"context"

	"github.com/umee-network/umee/v4/x/incentive"
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
	_ context.Context,
	_ *incentive.MsgClaim,
) (*incentive.MsgClaimResponse, error) {
	// TODO: Implement

	return &incentive.MsgClaimResponse{}, incentive.ErrNotImplemented
}

func (s msgServer) Bond(
	_ context.Context,
	_ *incentive.MsgBond,
) (*incentive.MsgBondResponse, error) {
	// TODO: Implement

	return &incentive.MsgBondResponse{}, incentive.ErrNotImplemented
}

func (s msgServer) BeginUnbonding(
	_ context.Context,
	_ *incentive.MsgBeginUnbonding,
) (*incentive.MsgBeginUnbondingResponse, error) {
	// TODO: Implement

	return &incentive.MsgBeginUnbondingResponse{}, incentive.ErrNotImplemented
}

func (s msgServer) Sponsor(
	_ context.Context,
	_ *incentive.MsgSponsor,
) (*incentive.MsgSponsorResponse, error) {
	// TODO: Implement

	return &incentive.MsgSponsorResponse{}, incentive.ErrNotImplemented
}

func (s msgServer) GovSetParams(
	_ context.Context,
	_ *incentive.MsgGovSetParams,
) (*incentive.MsgGovSetParamsResponse, error) {
	// TODO: Implement

	return &incentive.MsgGovSetParamsResponse{}, incentive.ErrNotImplemented
}

func (s msgServer) GovCreatePrograms(
	_ context.Context,
	_ *incentive.MsgGovCreatePrograms,
) (*incentive.MsgGovCreateProgramsResponse, error) {
	// TODO: Implement

	return &incentive.MsgGovCreateProgramsResponse{}, incentive.ErrNotImplemented
}
