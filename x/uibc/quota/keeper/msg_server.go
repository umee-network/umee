package keeper

import (
	"context"

	"github.com/umee-network/umee/v6/util/checkers"
	"github.com/umee-network/umee/v6/util/sdkutil"
	"github.com/umee-network/umee/v6/x/uibc"
)

var _ uibc.MsgServer = msgServer{}

type msgServer struct {
	kb Builder
}

// NewMsgServerImpl returns an implementation of uibc.MsgServer
func NewMsgServerImpl(kb Builder) uibc.MsgServer {
	return &msgServer{kb: kb}
}

// GovUpdateQuota implements types.MsgServer
func (m msgServer) GovUpdateQuota(ctx context.Context, msg *uibc.MsgGovUpdateQuota) (
	*uibc.MsgGovUpdateQuotaResponse, error,
) {
	sdkCtx, err := sdkutil.StartMsg(ctx, msg)
	if err != nil {
		return nil, err
	}

	k := m.kb.Keeper(&sdkCtx)
	byEmergencyGroup, err := checkers.EmergencyGroupAuthority(msg.Authority, k.ugov)
	if err != nil {
		return nil, err
	}

	if err := k.UpdateQuotaParams(msg.Total, msg.PerDenom, msg.QuotaDuration, byEmergencyGroup); err != nil {
		return nil, err
	}
	return &uibc.MsgGovUpdateQuotaResponse{}, nil
}

// GovSetIBCStatus implements types.MsgServer
func (m msgServer) GovSetIBCStatus(
	ctx context.Context, msg *uibc.MsgGovSetIBCStatus,
) (*uibc.MsgGovSetIBCStatusResponse, error) {
	sdkCtx, err := sdkutil.StartMsg(ctx, msg)
	if err != nil {
		return nil, err
	}

	k := m.kb.Keeper(&sdkCtx)
	if err := k.SetIBCStatus(msg.IbcStatus); err != nil {
		return &uibc.MsgGovSetIBCStatusResponse{}, err
	}
	sdkutil.Emit(&sdkCtx, &uibc.EventIBCTransferStatus{
		Status: msg.IbcStatus,
	})

	return &uibc.MsgGovSetIBCStatusResponse{}, nil
}
