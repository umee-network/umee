package keeper

import (
	context "context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v3/x/uibc"
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
	if err := m.keeper.UpdateQuota(ctx, msg.Total, msg.PerDenom, msg.QuotaDuration); err != nil {
		return nil, err
	}
	// save the new ibc rate limits
	return &uibc.MsgGovUpdateQuotaResponse{}, nil
}

// GovUpdateTransferStatus implements types.MsgServer
func (m msgServer) GovUpdateTransferStatus(
	goCtx context.Context, msg *uibc.MsgGovUpdateTransferStatus,
) (*uibc.MsgGovUpdateTransferStatusResonse, error) {
	ctx := sdk.UnwrapSDKContext(goCtx)

	if err := m.keeper.UpdateTansferStatus(ctx, msg.IbcPauseStatus); err != nil {
		return &uibc.MsgGovUpdateTransferStatusResonse{}, err
	}

	return &uibc.MsgGovUpdateTransferStatusResonse{}, nil
}
