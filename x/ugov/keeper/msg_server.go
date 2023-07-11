package keeper

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v5/util/sdkutil"
	"github.com/umee-network/umee/v5/x/ugov"
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

func (m msgServer) GovSetEmergencyGroup(_ context.Context, _ *ugov.MsgGovSetEmergencyGroup,
) (*ugov.MsgGovSetEmergencyGroupResponse, error) {
	panic("not implemented")
}
