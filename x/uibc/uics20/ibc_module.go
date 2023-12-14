package uics20

import (
	"cosmossdk.io/errors"
	ics20types "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v7/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v7/modules/core/exported"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"

	"github.com/umee-network/umee/v6/x/uibc"
	"github.com/umee-network/umee/v6/x/uibc/quota"
)

var _ porttypes.IBCModule = ICS20Module{}

// ICS20Module implements ibcporttypes.IBCModule for ICS20 transfer middleware.
// It overwrites OnAcknowledgementPacket and OnTimeoutPacket to revert
// quota update on acknowledgement error or timeout.
type ICS20Module struct {
	porttypes.IBCModule
	kb  quota.KeeperBuilder
	cdc codec.JSONCodec
}

// NewICS20Module is an IBCMiddlware constructor.
// `app` must be an ICS20 app.
func NewICS20Module(app porttypes.IBCModule, k quota.KeeperBuilder, cdc codec.JSONCodec) ICS20Module {
	return ICS20Module{
		IBCModule: app,
		kb:        k,
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

	if ftData.Memo != "" {
		msgs, err := DeserializeMemoMsgs(im.cdc, []byte(ftData.Memo))
		if err != nil {
			// TODO: need to verify if we want to stop the handle the error or revert the ibc transerf
			ctx.Logger().Error("can't JSON deserialize ftData Memo, expecting list of Msg", "err", err)
		} else {
			// TODO: need to handle fees!
			im.dispatchMemoMsgs(ctx, msgs)
		}
	}

	return im.IBCModule.OnRecvPacket(ctx, packet, relayer)
}

// OnAcknowledgementPacket is called on the packet sender chain, once the receiver acknowledged
// the packet reception.
func (im ICS20Module) OnAcknowledgementPacket(
	ctx sdk.Context, packet channeltypes.Packet, acknowledgement []byte, relayer sdk.AccAddress,
) error {
	var ack channeltypes.Acknowledgement
	if err := im.cdc.UnmarshalJSON(acknowledgement, &ack); err != nil {
		return errors.Wrap(err, "cannot unmarshal ICS-20 transfer packet acknowledgement")
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

// runs messages encoded in the ICS20 memo.
// NOTE: storage is forked, and only committed (flushed) if all messages pass and if all
// messages are supported. Otherwise the fork storage is discarded.
func (im ICS20Module) dispatchMemoMsgs(ctx sdk.Context, msgs []sdk.Msg) {
	// Caching context so that we don't update the store in case of failure.
	///cacheCtx, flush := ctx.CacheContext()
	// TODO: call flush on success

	for _, m := range msgs {
		ctx.Logger().Info("dispatching", "msg", m)
	}

}

func deserializeFTData(cdc codec.JSONCodec, packet channeltypes.Packet,
) (d ics20types.FungibleTokenPacketData, err error) {

	if err = cdc.UnmarshalJSON(packet.GetData(), &d); err != nil {
		err = errors.Wrap(err,
			"cannot unmarshal ICS-20 transfer packet data")
	}
	return
}

func DeserializeMemoMsgs(cdc codec.JSONCodec, data []byte) ([]sdk.Msg, error) {
	var m uibc.ICS20Memo
	if err := cdc.UnmarshalJSON(data, &m); err != nil {
		return nil, err
	}
	return tx.GetMsgs(m.Messages, "memo messages")
}
