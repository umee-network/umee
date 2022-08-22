package ibctransfer

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	ibctransfer "github.com/cosmos/ibc-go/v3/modules/apps/transfer"
	ibctransfertypes "github.com/cosmos/ibc-go/v3/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v3/modules/core/04-channel/types"
	ibcexported "github.com/cosmos/ibc-go/v3/modules/core/exported"

	"github.com/umee-network/umee/v2/x/ibctransfer/keeper"
)

// IBCModule embeds the ICS-20 transfer IBCModule where we only override specific
// methods.
type IBCModule struct {
	// embed the ICS-20 transfer's AppModule
	ibctransfer.IBCModule

	keeper keeper.Keeper
}

func NewIBCModule(am ibctransfer.IBCModule, k keeper.Keeper) IBCModule {
	return IBCModule{
		IBCModule: am,
		keeper:    k,
	}
}

// OnRecvPacket delegates the OnRecvPacket call to the embedded ICS-20 transfer
// IBCModule and updates metadata if successful.
func (am IBCModule) OnRecvPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
	relayer sdk.AccAddress,
) ibcexported.Acknowledgement {
	ack := am.IBCModule.OnRecvPacket(ctx, packet, relayer)
	if ack.Success() {
		var data ibctransfertypes.FungibleTokenPacketData
		if err := ibctransfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err == nil {
			// track metadata
			am.keeper.PostOnRecvPacket(ctx, packet, data)
		}
	}

	return ack
}
