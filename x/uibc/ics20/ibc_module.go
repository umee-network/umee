package ics20

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	ibctransfer "github.com/cosmos/ibc-go/v6/modules/apps/transfer"
	ibctransfertypes "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v6/modules/core/04-channel/types"
	ibcporttypes "github.com/cosmos/ibc-go/v6/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v6/modules/core/exported"

	ltypes "github.com/umee-network/umee/v4/x/leverage/types"
	"github.com/umee-network/umee/v4/x/uibc"
	"github.com/umee-network/umee/v4/x/uibc/ics20/keeper"
)

// IBCModule wraps ICS-20 IBC module to limit token transfer inflows.
type IBCModule struct {
	// leverage keeper
	lkeeper uibc.Leverage
	// embed the ICS-20 transfer's AppModule: ibctransfer.IBCModule
	ibcporttypes.IBCModule

	keeper keeper.Keeper
}

func NewIBCModule(leverageKeeper uibc.Leverage, am ibctransfer.IBCModule, k keeper.Keeper) IBCModule {
	return IBCModule{
		lkeeper:   leverageKeeper,
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
	var data ibctransfertypes.FungibleTokenPacketData
	if err := ibctransfertypes.ModuleCdc.UnmarshalJSON(packet.GetData(), &data); err != nil {
		ackErr := sdkerrors.ErrInvalidType.Wrap("cannot unmarshal ICS-20 transfer packet data")
		return channeltypes.NewErrorAcknowledgement(ackErr)
	}

	// Allowing only registered token for ibc transfer
	// TODO: re-enable inflow checks
	// isSourceChain := ibctransfertypes.SenderChainIsSource(packet.GetSourcePort(), packet.GetSourceChannel(), data.Denom)
	// ackErr := CheckIBCInflow(ctx, packet, am.lkeeper, data.Denom, isSourceChain)
	// if ackErr != nil {
	// 	return ackErr
	// }

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

func CheckIBCInflow(ctx sdk.Context,
	packet channeltypes.Packet,
	lkeeper uibc.Leverage,
	dataDenom string, isSourceChain bool,
) ibcexported.Acknowledgement {
	// if chain is recevier and sender chain is source then we need create ibc_denom (ibc/hash(channel,denom)) to
	// check ibc_denom is exists in leverage token registry
	if isSourceChain {
		// since SendPacket did not prefix the denomination, we must prefix denomination here
		sourcePrefix := ibctransfertypes.GetDenomPrefix(packet.GetDestPort(), packet.GetDestChannel())
		// NOTE: sourcePrefix contains the trailing "/"
		prefixedDenom := sourcePrefix + dataDenom
		// construct the denomination trace from the full raw denomination and get the ibc_denom
		ibcDenom := ibctransfertypes.ParseDenomTrace(prefixedDenom).IBCDenom()
		_, err := lkeeper.GetTokenSettings(ctx, ibcDenom)
		if err != nil {
			if ltypes.ErrNotRegisteredToken.Is(err) {
				return channeltypes.NewErrorAcknowledgement(err)
			}
			// other leverage keeper error -> log the error  and allow the inflow transfer.
			ctx.Logger().Error("IBC inflows: can't load token registry", "err", err)
		}
	}

	return nil
}
