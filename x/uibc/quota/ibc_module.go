package quota

import (
	"encoding/json"

	sdkmath "cosmossdk.io/math"
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	transfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v5/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v5/modules/core/exported"

	"github.com/umee-network/umee/v4/x/uibc"
	"github.com/umee-network/umee/v4/x/uibc/quota/keeper"
)

type IBCMiddleware struct {
	app    porttypes.IBCModule
	keeper keeper.Keeper
}

// NewIBCMiddleware creates a new IBCMiddlware given the keeper and underlying application
func NewIBCMiddleware(app porttypes.IBCModule, k keeper.Keeper) IBCMiddleware {
	return IBCMiddleware{
		app:    app,
		keeper: k,
	}
}

// OnChanOpenInit implements types.Middleware
func (im IBCMiddleware) OnChanOpenInit(ctx sdk.Context, order channeltypes.Order, connectionHops []string,
	portID string, channelID string, channelCap *capabilitytypes.Capability, counterparty channeltypes.Counterparty,
	version string,
) (string, error) {
	return im.app.OnChanOpenInit(
		ctx,
		order,
		connectionHops,
		portID,
		channelID,
		channelCap,
		counterparty,
		version,
	)
}

// OnChanOpenTry implements types.Middleware
func (im IBCMiddleware) OnChanOpenTry(
	ctx sdk.Context,
	order channeltypes.Order,
	connectionHops []string,
	portID string, channelID string,
	channelCap *capabilitytypes.Capability, counterparty channeltypes.Counterparty,
	counterpartyVersion string) (version string, err error) {
	return im.app.OnChanOpenTry(
		ctx, order, connectionHops, portID, channelID, channelCap,
		counterparty, counterpartyVersion,
	)
}

// OnChanOpenAck implements types.Middleware
func (im IBCMiddleware) OnChanOpenAck(ctx sdk.Context, portID string, channelID string, counterpartyChannelID string,
	counterpartyVersion string,
) error {
	return im.app.OnChanOpenAck(ctx, portID, channelID, counterpartyChannelID, counterpartyVersion)
}

// OnChanOpenConfirm implements types.Middleware
func (im IBCMiddleware) OnChanOpenConfirm(ctx sdk.Context, portID string, channelID string) error {
	return im.app.OnChanOpenConfirm(ctx, portID, channelID)
}

// OnChanCloseInit implements types.Middleware
func (im IBCMiddleware) OnChanCloseInit(ctx sdk.Context, portID string, channelID string) error {
	return im.app.OnChanCloseInit(ctx, portID, channelID)
}

// OnChanCloseConfirm implements types.Middleware
func (im IBCMiddleware) OnChanCloseConfirm(ctx sdk.Context, portID string, channelID string) error {
	return im.app.OnChanCloseConfirm(ctx, portID, channelID)
}

// OnRecvPacket implements types.Middleware
func (im IBCMiddleware) OnRecvPacket(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress,
) exported.Acknowledgement {
	// if this returns an Acknowledgement that isn't successful, all state changes are discarded
	return im.app.OnRecvPacket(ctx, packet, relayer)
}

// OnAcknowledgementPacket implements types.Middleware
func (im IBCMiddleware) OnAcknowledgementPacket(ctx sdk.Context, packet channeltypes.Packet, acknowledgement []byte,
	relayer sdk.AccAddress,
) error {
	var ack channeltypes.Acknowledgement
	if err := json.Unmarshal(acknowledgement, &ack); err != nil {
		return sdkerrors.ErrUnknownRequest.Wrapf("cannot unmarshal ICS-20 transfer packet acknowledgement: %v", err)
	}

	if isAckError(acknowledgement) {
		err := im.RevertSentPacket(ctx, packet) // If there is an error here we should still handle the timeout
		if err != nil {
			if err = ctx.EventManager().EmitTypedEvent(&uibc.EventBadRevert{
				Module:      uibc.ModuleName,
				FailureType: "acknowledgment",
				Packet:      string(packet.GetData()),
			}); err != nil {
				return err
			}
		}
	}

	return im.app.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
}

// IsAckError checks an IBC acknowledgement to see if it's an error.
// This is a replacement for ack.Success() which is currently not working on some circumstances
func isAckError(acknowledgement []byte) bool {
	var ackErr channeltypes.Acknowledgement_Error
	if err := json.Unmarshal(acknowledgement, &ackErr); err == nil && len(ackErr.Error) > 0 {
		return true
	}
	return false
}

// OnTimeoutPacket implements types.Middleware
func (im IBCMiddleware) OnTimeoutPacket(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) error {
	err := im.RevertSentPacket(ctx, packet) // If there is an error here we should still handle the timeout
	if err != nil {
		if err = ctx.EventManager().EmitTypedEvent(&uibc.EventBadRevert{
			Module:      uibc.ModuleName,
			FailureType: "timeout",
			Packet:      string(packet.GetData()),
		}); err != nil {
			return err
		}
	}
	return im.app.OnTimeoutPacket(ctx, packet, relayer)
}

// RevertSentPacket Notifies the contract that a sent packet wasn't properly received
func (im IBCMiddleware) RevertSentPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
) error {
	var data transfertypes.FungibleTokenPacketData
	if err := json.Unmarshal(packet.GetData(), &data); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest,
			"cannot unmarshal ICS-20 transfer packet data: %s", err.Error(),
		)
	}

	amount, ok := sdkmath.NewIntFromString(data.Amount)
	if !ok {
		return uibc.ErrInvalidAmount.Wrapf("amount %s", data.Amount)
	}

	return im.keeper.UndoUpdateQuota(
		ctx,
		data.Denom,
		amount,
	)
}

func ValidateReceiverAddress(packet channeltypes.Packet) error {
	var packetData transfertypes.FungibleTokenPacketData
	if err := json.Unmarshal(packet.GetData(), &packetData); err != nil {
		return err
	}
	if len(packetData.Receiver) >= 4096 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress,
			"IBC Receiver address too long. Max supported length is %d", 4096,
		)
	}
	return nil
}
