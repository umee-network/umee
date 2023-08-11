package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v6/util/sdkutil"
	"github.com/umee-network/umee/v6/x/ugov"
)

type msgServer struct {
	kb Builder
}

// NewMsgServer returns an implementation of uibc.MsgServer
func NewMsgServer(kb Builder) ugov.MsgServer {
	return msgServer{kb: kb}
}

// GovUpdateMinFees sets protocol controlled tx min fees.
func (m msgServer) GovUpdateMinGasPrice(ctx context.Context, msg *ugov.MsgGovUpdateMinGasPrice,
) (*ugov.MsgGovUpdateMinGasPriceResponse, error) {
	sdkCtx, err := sdkutil.StartMsg(ctx, msg)
	if err != nil {
		return nil, err
	}

	k := m.kb.Keeper(&sdkCtx)
	if err := k.SetMinGasPrice(msg.MinGasPrice); err != nil {
		return nil, err
	}
	sdkutil.Emit(&sdkCtx, &ugov.EventMinGasPrice{
		MinGasPrices: sdk.NewDecCoins(msg.MinGasPrice),
	})

	return &ugov.MsgGovUpdateMinGasPriceResponse{}, nil
}

func (m msgServer) GovSetEmergencyGroup(ctx context.Context, msg *ugov.MsgGovSetEmergencyGroup,
) (*ugov.MsgGovSetEmergencyGroupResponse, error) {
	sdkCtx, err := sdkutil.StartMsg(ctx, msg)
	if err != nil {
		return nil, err
	}
	addr, err := sdk.AccAddressFromBech32(msg.EmergencyGroup)
	if err != nil {
		return nil, err
	}
	m.kb.Keeper(&sdkCtx).SetEmergencyGroup(addr)
	sdkutil.Emit(&sdkCtx, &ugov.EventEmergencyGroup{
		EmergencyGroup: msg.EmergencyGroup,
	})

	return &ugov.MsgGovSetEmergencyGroupResponse{}, nil
}

// GovUpdateInflationParams implements ugov.MsgServer.
func (m msgServer) GovUpdateInflationParams(ctx context.Context, msg *ugov.MsgGovUpdateInflationParams) (
	*ugov.GovUpdateInflationParamsResponse, error) {

	sdkCtx, err := sdkutil.StartMsg(ctx, msg)
	if err != nil {
		return nil, err
	}

	err = m.kb.Keeper(&sdkCtx).SetInflationParams(msg.Params)
	if err != nil {
		return nil, err
	}

	return &ugov.GovUpdateInflationParamsResponse{}, nil
}
