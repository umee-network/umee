package keeper

import (
	"context"

	"github.com/umee-network/umee/v4/util/sdkutil"
	"github.com/umee-network/umee/v4/x/uibc"
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
	if err := k.UpdateQuotaParams(msg.Total, msg.PerDenom, msg.QuotaDuration); err != nil {
		return nil, err
	}
	return &uibc.MsgGovUpdateQuotaResponse{}, nil
}

// GovSetIBCQuotaStatus implements types.MsgServer
func (m msgServer) GovSetIBCQuotaStatus(
	ctx context.Context, msg *uibc.MsgGovSetIBCSQuotaStatus,
) (*uibc.MsgGovSetIBCQuotaStatusResponse, error) {
	sdkCtx, err := sdkutil.StartMsg(ctx, msg)
	if err != nil {
		return nil, err
	}

	k := m.kb.Keeper(&sdkCtx)
	if err := k.SetIBCTransferQuotaStatus(msg.QuotaStatus); err != nil {
		return &uibc.MsgGovSetIBCQuotaStatusResponse{}, err
	}
	sdkutil.Emit(&sdkCtx, &uibc.EventIBCTransferQuotaStatus{
		Status: msg.QuotaStatus,
	})

	return &uibc.MsgGovSetIBCQuotaStatusResponse{}, nil
}
