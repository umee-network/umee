package keeper

import (
	context "context"

	"github.com/umee-network/umee/v4/util/sdkutil"
	"github.com/umee-network/umee/v4/x/uibc"
)

var _ uibc.MsgServer = msgServer{}

type msgServer struct {
	kb KeeperBuilder
}

// NewMsgServerImpl returns an implementation of MsgServer for the x/uibc
// module.
func NewMsgServerImpl(kb KeeperBuilder) uibc.MsgServer {
	return &msgServer{kb: kb}
}

// GovUpdateQuota implements types.MsgServer
func (m msgServer) GovUpdateQuota(goCtx context.Context, msg *uibc.MsgGovUpdateQuota) (
	*uibc.MsgGovUpdateQuotaResponse, error,
) {
	ctx, err := sdkutil.StartMsg(goCtx, msg)
	if err != nil {
		return nil, err
	}

	k := m.kb.Keeper(&ctx)
	if err := k.UpdateQuotaParams(msg.Total, msg.PerDenom, msg.QuotaDuration); err != nil {
		return nil, err
	}
	return &uibc.MsgGovUpdateQuotaResponse{}, nil
}

// GovSetIBCStatus implements types.MsgServer
func (m msgServer) GovSetIBCStatus(
	goCtx context.Context, msg *uibc.MsgGovSetIBCStatus,
) (*uibc.MsgGovSetIBCStatusResponse, error) {
	ctx, err := sdkutil.StartMsg(goCtx, msg)
	if err != nil {
		return nil, err
	}

	k := m.kb.Keeper(&ctx)
	if err := k.SetIBCStatus(msg.IbcStatus); err != nil {
		return &uibc.MsgGovSetIBCStatusResponse{}, err
	}
	sdkutil.Emit(&ctx, &uibc.EventIBCTransferStatus{
		Status: msg.IbcStatus,
	})

	return &uibc.MsgGovSetIBCStatusResponse{}, nil
}
