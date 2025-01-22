package uics20

import (
	"encoding/json"
	"errors"

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
	kb       quota.Builder
	leverage ltypes.MsgServer
	cdc      codec.JSONCodec
}

// NewICS20Module is an IBCMiddlware constructor.
// `app` must be an ICS20 app.
func NewICS20Module(app porttypes.IBCModule, cdc codec.JSONCodec, k quota.Builder, l ltypes.MsgServer,
) ICS20Module {
	return ICS20Module{
		IBCModule: app,
		kb:        k,
		leverage:  l,
		cdc:       cdc,
	}
}

// OnRecvPacket is called when a receiver chain receives a packet from SendPacket.
//  1. record IBC quota
//  2. Try to unpack and prepare memo. If memo has a correct structure, and fallback addr is
//     defined but malformed, we cancel the transfer (otherwise would not be able to use it
//     correctly).
//  3. If memo has a correct structure, but memo.messages can't be unpack or don't pass
//     validation, then we continue with the transfer and overwrite the original receiver to
//     fallback_addr if it's defined.
//  4. Execute the downstream middleware and the transfer app.
//  5. Execute hooks. If hook execution fails, and the fallback_addr is defined, then we revert
//     the transfer (and all related state changes and events) and use send the tokens to the
//     `fallback_addr` instead.
func (im ICS20Module) OnRecvPacket(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress,
) exported.Acknowledgement {
	ftData, err := deserializeFTData(im.cdc, packet)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}
	quotaKeeper := im.kb.Keeper(&ctx)
	if ackResp := quotaKeeper.IBCOnRecvPacket(ftData, packet); ackResp != nil && !ackResp.Success() {
		return ackResp
	}

	params := quotaKeeper.GetParams()

	// NOTE: IBC hooks must be the last middleware - just the transfer app.
	// MemoHandler may update amoount in the message, because the received token amount may be
	// smaller than the amount originally sent (various fees). We need to be sure that there is
	// no other middleware that can change packet data or amounts.

	mh := MemoHandler{executeEnabled: params.Ics20Hooks, cdc: im.cdc, leverage: im.leverage}
	events, err := mh.onRecvPacketPrepare(&ctx, packet, ftData)
	if err != nil {
		if !errors.Is(err, errMemoValidation{}) {
			return channeltypes.NewErrorAcknowledgement(err)
		}
		if mh.fallbackReceiver != nil {
			ftData.Receiver = mh.fallbackReceiver.String()
			events = append(events, "overwrite receiver to fallback_addr="+ftData.Receiver)
			if packet.Data, err = json.Marshal(ftData); err != nil {
				return channeltypes.NewErrorAcknowledgement(err)
			}
		}
		im.emitEvents(ctx.EventManager(), recvPacketLogger(&ctx), "ics20-memo-hook", events)
		return im.IBCModule.OnRecvPacket(ctx, packet, relayer)
	}

	var transferCtxFlush func()
	var transferCtx = ctx
	if mh.fallbackReceiver != nil {
		// create a new cache context for the fallback receiver. We will discard it when
		// the execution fails.
		transferCtx, transferCtxFlush = ctx.CacheContext()
	}
	execCtx, execCtxFlush := transferCtx.CacheContext()

	// call transfer module app
	// ack is nil if acknowledgement is asynchronous
	ack := im.IBCModule.OnRecvPacket(transferCtx, packet, relayer)
	if ack != nil && !ack.Success() {
		goto end
	}

	if err = mh.execute(&execCtx); err != nil {
		events = append(events, "can't handle ICS20 memo err = "+err.Error())
		// if we created a new cache context, then we can discard it, and repeate the transfer to
		// the fallback address
		if mh.fallbackReceiver != nil {
			transferCtxFlush = nil // discard the context
			ftData.Receiver = mh.fallbackReceiver.String()
			events = append(events, "overwrite receiver to fallback_addr="+ftData.Receiver)
			if packet.Data, err = json.Marshal(ftData); err != nil {
				return channeltypes.NewErrorAcknowledgement(err)
			}
			ack = im.IBCModule.OnRecvPacket(ctx, packet, relayer)
		}
	} else {
		execCtxFlush()
	}

end:
	if transferCtxFlush != nil {
		transferCtxFlush()
	}
	im.emitEvents(ctx.EventManager(), recvPacketLogger(&ctx), "ics20-memo-hook", events)
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
		ctx.Logger().With("scope", "ics20-OnAckErr").Error("can't revert quota update", "err", err)
	}
	qk := im.kb.Keeper(ctx)
	qk.IBCRevertQuotaUpdate(ftData.Amount, ftData.Denom)
}

func (im ICS20Module) emitEvents(em *sdk.EventManager, logger log.Logger, topic string, events []string) {
	attributes := make([]sdk.Attribute, len(events))
	key := topic + "-context"
	for i, s := range events {
		// it's ok that all events have the same key. This is how ibc-apps are dealing with events.
		attributes[i] = sdk.NewAttribute(key, s)
	}
	logger.Debug("Handle ICS20 memo", "events", events)

	em.EmitEvents(sdk.Events{
		sdk.NewEvent(
			topic,
			attributes...,
		),
	})
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
