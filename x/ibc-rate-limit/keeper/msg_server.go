package keeper

import (
	context "context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"

	"github.com/umee-network/umee/v3/x/ibc-rate-limit/types"
)

var _ types.MsgServer = msgServer{}

type msgServer struct {
	keeper Keeper
}

// NewMsgServerImpl returns an implementation of MsgServer for the x/ibc-rate-limit
// module.
func NewMsgServerImpl(keeper Keeper) types.MsgServer {
	return &msgServer{keeper: keeper}
}

// AddIBCDenomRateLimit implements types.MsgServer
func (m msgServer) UpdateIBCDenomsRateLimit(goCtx context.Context, msg *types.MsgUpdateIBCDenomsRateLimit) (*types.MsgUpdateIBCDenomsRateLimitResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// checking req msg authority is the gov module address
	if m.keeper.authority != msg.Authority {
		return &types.MsgUpdateIBCDenomsRateLimitResponse{},
			govtypes.ErrInvalidSigner.Wrapf(
				"invalid authority: expected %s, got %s",
				m.keeper.authority, msg.Authority,
			)
	}

	if err := m.keeper.SetRateLimitsOfIBCDenoms(ctx, msg.NewIbcDenomsRateLimits); err != nil {
		return &types.MsgUpdateIBCDenomsRateLimitResponse{}, err
	}

	return &types.MsgUpdateIBCDenomsRateLimitResponse{}, nil
}

// UpdateIBCTransferPauseStatus implements types.MsgServer
func (m msgServer) UpdateIBCTransferPauseStatus(goCtx context.Context, msg *types.MsgUpdateIBCTransferPauseStatus) (*types.MsgUpdateIBCTransferPauseStatusResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	// checking req msg authority is the gov module address
	if m.keeper.authority != msg.Authority {
		return &types.MsgUpdateIBCTransferPauseStatusResponse{},
			govtypes.ErrInvalidSigner.Wrapf(
				"invalid authority: expected %s, got %s",
				m.keeper.authority, msg.Authority,
			)
	}

	if err := m.keeper.UpdateIBCTansferStatus(ctx, msg.IbcPauseStatus); err != nil {
		return &types.MsgUpdateIBCTransferPauseStatusResponse{}, err
	}

	return &types.MsgUpdateIBCTransferPauseStatusResponse{}, nil
}
