package quota

import (
	"context"

	"github.com/umee-network/umee/v6/util/checkers"
	"github.com/umee-network/umee/v6/util/sdkutil"
	"github.com/umee-network/umee/v6/x/uibc"
)

var _ uibc.MsgServer = msgServer{}

type msgServer struct {
	kb KeeperBuilder
}

// NewMsgServerImpl returns an implementation of uibc.MsgServer
func NewMsgServerImpl(kb KeeperBuilder) uibc.MsgServer {
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

	if err := k.UpdateQuotaParams(msg, byEmergencyGroup); err != nil {
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
	// emergency group can change status to any valid value
	if _, err = checkers.EmergencyGroupAuthority(msg.Authority, k.ugov); err != nil {
		return nil, err
	}

	if err := k.SetIBCStatus(msg.IbcStatus); err != nil {
		return &uibc.MsgGovSetIBCStatusResponse{}, err
	}
	sdkutil.Emit(&sdkCtx, &uibc.EventIBCTransferStatus{
		Status: msg.IbcStatus,
	})

	return &uibc.MsgGovSetIBCStatusResponse{}, nil
}
