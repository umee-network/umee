package quota

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
	sdkutil.Emit(&sdkCtx, &uibc.EventIBCTransferStatus{Status: msg.IbcStatus})

	return &uibc.MsgGovSetIBCStatusResponse{}, nil
}

// GovToggleICS20Hooks implements types.MsgServer
func (m msgServer) GovToggleICS20Hooks(ctx context.Context, msg *uibc.MsgGovToggleICS20Hooks,
) (*uibc.MsgGovToggleICS20HooksResponse, error) {
	sdkCtx, err := sdkutil.StartMsg(ctx, msg)
	if err != nil {
		return nil, err
	}

	k := m.kb.Keeper(&sdkCtx)
	// emergency group can change status to any valid value
	if _, err = checkers.EmergencyGroupAuthority(msg.Authority, k.ugov); err != nil {
		return nil, err
	}

	if err := k.SetICS20HooksStatus(msg.Enabled); err != nil {
		return &uibc.MsgGovToggleICS20HooksResponse{}, err
	}
	sdkutil.Emit(&sdkCtx, &uibc.EventICS20Hooks{Enabled: msg.Enabled})

	return &uibc.MsgGovToggleICS20HooksResponse{}, nil
}
