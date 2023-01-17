package keeper

import (
	context "context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v4/x/uibc"
)

var _ uibc.MsgServer = msgServer{}

type msgServer struct {
	keeper Keeper
}

// NewMsgServerImpl returns an implementation of MsgServer for the x/uibc
// module.
func NewMsgServerImpl(keeper Keeper) uibc.MsgServer {
	return &msgServer{keeper: keeper}
}

// GovUpdateQuota implements types.MsgServer
func (m msgServer) GovUpdateQuota(goCtx context.Context, msg *uibc.MsgGovUpdateQuota) (
	*uibc.MsgGovUpdateQuotaResponse, error,
) {
	ctx := sdk.UnwrapSDKContext(goCtx)
	if err := m.keeper.UpdateQuotaParams(ctx, msg.Total, msg.PerDenom, msg.QuotaDuration); err != nil {
		return nil, err
	}
	// save the new ibc rate limits
	return &uibc.MsgGovUpdateQuotaResponse{}, nil
}

// GovSetIBCPause implements types.MsgServer
func (m msgServer) GovSetIBCPause(
	goCtx context.Context, msg *uibc.MsgGovSetIBCPause,
) (*uibc.MsgGovSetIBCPauseResponse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := m.keeper.SetIBCPause(ctx, msg.IbcPauseStatus); err != nil {
		return &uibc.MsgGovSetIBCPauseResponse{}, err
	}

	return &uibc.MsgGovSetIBCPauseResponse{}, nil
}
