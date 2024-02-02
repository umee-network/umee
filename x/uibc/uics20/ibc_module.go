package uics20

import (
	sdkerrors "cosmossdk.io/errors"
	"github.com/cometbft/cometbft/libs/log"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	ics20types "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"

	ltypes "github.com/umee-network/umee/v6/x/leverage/types"
	"github.com/umee-network/umee/v6/x/uibc/quota"
)

var _ porttypes.IBCModule = ICS20Module{}

// ICS20Module implements ibcporttypes.IBCModule for ICS20 transfer middleware.
// It overwrites OnAcknowledgementPacket and OnTimeoutPacket to revert
// quota update on acknowledgement error or timeout.
type ICS20Module struct {
	porttypes.IBCModule
	kb       quota.KeeperBuilder
	leverage ltypes.MsgServer
	cdc      codec.JSONCodec
}

// NewICS20Module is an IBCMiddlware constructor.
// `app` must be an ICS20 app.
func NewICS20Module(app porttypes.IBCModule, cdc codec.JSONCodec, k quota.KeeperBuilder, l ltypes.MsgServer,
) ICS20Module {
	return ICS20Module{
		IBCModule: app,
		kb:        k,
		leverage:  l,
		cdc:       cdc,
	}
}

// OnRecvPacket is called when a receiver chain receives a packet from SendPacket.
func (im ICS20Module) OnRecvPacket(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress,
) exported.Acknowledgement {
	ftData, err := deserializeFTData(im.cdc, packet)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}
	qk := im.kb.Keeper(&ctx)
	if ackResp := qk.IBCOnRecvPacket(ftData, packet); ackResp != nil && !ackResp.Success() {
		return ackResp
	}

	ack := im.IBCModule.OnRecvPacket(ctx, packet, relayer)
	if !ack.Success() {
		return ack
	}
	if ftData.Memo != "" {
		logger := recvPacketLogger(&ctx)
		mh := MemoHandler{im.cdc, im.leverage}
		if err := mh.onRecvPacket(&ctx, ftData); err != nil {
			logger.Error("can't handle ICS20 memo", "err", err)
		}
	}

	return ack
}

// OnAcknowledgementPacket is called on the packet sender chain, once the receiver acknowledged
// the packet reception.
func (im ICS20Module) OnAcknowledgementPacket(
	ctx sdk.Context, packet channeltypes.Packet, acknowledgement []byte, relayer sdk.AccAddress,
) error {
	var ack channeltypes.Acknowledgement
	if err := im.cdc.UnmarshalJSON(acknowledgement, &ack); err != nil {
		return sdkerrors.Wrap(err, "cannot unmarshal ICS-20 transfer packet acknowledgement")
	}
	if _, isErr := ack.Response.(*channeltypes.Acknowledgement_Error); isErr {
		// we don't return to propagate the ack error to the other layers
		im.onAckErr(&ctx, packet)
	}

	return im.IBCModule.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
}

// OnTimeoutPacket implements types.Middleware
func (im ICS20Module) OnTimeoutPacket(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) error {
	im.onAckErr(&ctx, packet)
	return im.IBCModule.OnTimeoutPacket(ctx, packet, relayer)
}

func (im ICS20Module) onAckErr(ctx *sdk.Context, packet channeltypes.Packet) {
	ftData, err := deserializeFTData(im.cdc, packet)
	if err != nil {
		// we only log error, because we want to propagate the ack to other layers.
		ctx.Logger().Error("can't revert quota update", "err", err)
	}
	qk := im.kb.Keeper(ctx)
	qk.IBCRevertQuotaUpdate(ftData.Amount, ftData.Denom)
}

func deserializeFTData(cdc codec.JSONCodec, packet channeltypes.Packet,
) (d ics20types.FungibleTokenPacketData, err error) {
	if err = cdc.UnmarshalJSON(packet.GetData(), &d); err != nil {
		err = sdkerrors.Wrap(err, "cannot unmarshal ICS-20 transfer packet data")
	}
	return
}

func recvPacketLogger(ctx *sdk.Context) log.Logger {
	return ctx.Logger().With("scope", "ics20-OnRecvPacket")
}
