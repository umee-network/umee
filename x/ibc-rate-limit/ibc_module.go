package ibc_rate_limit

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	transfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v5/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v5/modules/core/exported"
	"github.com/umee-network/umee/v3/x/ibc-rate-limit/keeper"
	"github.com/umee-network/umee/v3/x/ibc-rate-limit/types"
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
func (im IBCMiddleware) OnChanOpenInit(ctx sdk.Context, order channeltypes.Order, connectionHops []string, portID string, channelID string, channelCap *capabilitytypes.Capability, counterparty channeltypes.Counterparty, version string) (string, error) {
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
	return im.app.OnChanOpenTry(ctx, order, connectionHops, portID, channelID, channelCap, counterparty, counterpartyVersion)
}

// OnChanOpenAck implements types.Middleware
func (im IBCMiddleware) OnChanOpenAck(ctx sdk.Context, portID string, channelID string, counterpartyChannelID string, counterpartyVersion string) error {
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
func (im IBCMiddleware) OnRecvPacket(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) exported.Acknowledgement {
	if err := ValidateReceiverAddress(packet); err != nil {
		return channeltypes.NewErrorAcknowledgement(err)
	}

	amount, denom, err := im.keeper.GetFundsFromPacket(packet)
	if err != nil {
		return channeltypes.NewErrorAcknowledgement(types.ErrBadPacket.Wrapf("bad packet in rate limit's OnRecvPacket"))
	}

	if err := im.keeper.CheckAndUpdateRateLimits(ctx, denom, amount); err != nil {
		return channeltypes.NewErrorAcknowledgement(types.ErrRateLimitExceeded)
	}

	// if this returns an Acknowledgement that isn't successful, all state changes are discarded
	return im.app.OnRecvPacket(ctx, packet, relayer)
}

// OnAcknowledgementPacket implements types.Middleware
func (im IBCMiddleware) OnAcknowledgementPacket(ctx sdk.Context, packet channeltypes.Packet, acknowledgement []byte, relayer sdk.AccAddress) error {
	var ack channeltypes.Acknowledgement
	if err := json.Unmarshal(acknowledgement, &ack); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-20 transfer packet acknowledgement: %v", err)
	}

	if !ack.Success() {
		// TODO: reverse sent the ibc transfer
		err := im.RevertSentPacket(ctx, packet) // If there is an error here we should still handle the timeout
		if err != nil {
			ctx.EventManager().EmitEvent(
				sdk.NewEvent(
					types.EventBadRevert,
					sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
					sdk.NewAttribute(types.AttributeKeyFailureType, "timeout"),
					sdk.NewAttribute(types.AttributeKeyPacket, string(packet.GetData())),
				),
			)
		}
	}

	return im.app.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
}

// OnTimeoutPacket implements types.Middleware
func (im IBCMiddleware) OnTimeoutPacket(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) error {
	err := im.RevertSentPacket(ctx, packet) // If there is an error here we should still handle the timeout
	if err != nil {
		ctx.EventManager().EmitEvent(
			sdk.NewEvent(
				types.EventBadRevert,
				sdk.NewAttribute(sdk.AttributeKeyModule, types.ModuleName),
				sdk.NewAttribute(types.AttributeKeyFailureType, "timeout"),
				sdk.NewAttribute(types.AttributeKeyPacket, string(packet.GetData())),
			),
		)
	}
	return im.app.OnTimeoutPacket(ctx, packet, relayer)
}

// RevertSentPacket Notifies the contract that a sent packet wasn't properly received
func (im *IBCMiddleware) RevertSentPacket(
	ctx sdk.Context,
	packet channeltypes.Packet,
) error {
	var data transfertypes.FungibleTokenPacketData
	if err := json.Unmarshal(packet.GetData(), &data); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-20 transfer packet data: %s", err.Error())
	}

	if err := im.keeper.UndoSendRateLimit(
		ctx,
		data.Denom,
		data.Amount,
	); err != nil {
		return err
	}

	return nil
}

func ValidateReceiverAddress(packet channeltypes.Packet) error {
	var packetData transfertypes.FungibleTokenPacketData
	if err := json.Unmarshal(packet.GetData(), &packetData); err != nil {
		return err
	}
	if len(packetData.Receiver) >= 4096 {
		return sdkerrors.Wrapf(sdkerrors.ErrInvalidAddress, "IBC Receiver address too long. Max supported length is %d", 4096)
	}
	return nil
}
