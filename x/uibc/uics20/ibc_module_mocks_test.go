package uics20

import (
	"context"

	sdk "github.com/cosmos/cosmos-sdk/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"

	"github.com/umee-network/umee/v6/x/leverage/types"
)

type MOCKIBCModule struct {
	porttypes.IBCModule
}

func NewMockIBCModule() MOCKIBCModule {
	return MOCKIBCModule{}
}

func (m MOCKIBCModule) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) exported.Acknowledgement {
	return channeltypes.NewResultAcknowledgement([]byte("true"))
}

type MockLeverageMsgServer struct {
	types.MsgServer
}

func NewMockLeverageMsgServer() MockLeverageMsgServer {
	return MockLeverageMsgServer{}
}

// Supply implements types.MsgServer.
func (m MockLeverageMsgServer) Supply(context.Context, *types.MsgSupply) (*types.MsgSupplyResponse, error) {
	return &types.MsgSupplyResponse{}, nil
}
