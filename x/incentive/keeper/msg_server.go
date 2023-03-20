package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"

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
	goCtx context.Context,
	msg *incentive.MsgGovSetParams,
) (*incentive.MsgGovSetParamsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// todo: check GetSigners security, other things

	if err := msg.Params.Validate(); err != nil {
		return &incentive.MsgGovSetParamsResponse{}, err
	}

	if err := s.keeper.SetParams(ctx, msg.Params); err != nil {
		return &incentive.MsgGovSetParamsResponse{}, err
	}

	return &incentive.MsgGovSetParamsResponse{}, nil
}

func (s msgServer) GovCreatePrograms(
	goCtx context.Context,
	msg *incentive.MsgGovCreatePrograms,
) (*incentive.MsgGovCreateProgramsResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// todo: check GetSigners security, other things

	// For each program being created, create it with the next available ID
	for _, program := range msg.Programs {
		if err := s.keeper.createIncentiveProgram(ctx, program, msg.FromCommunityFund); err != nil {
			return &incentive.MsgGovCreateProgramsResponse{}, err
		}
	}

	return &incentive.MsgGovCreateProgramsResponse{}, nil
}
