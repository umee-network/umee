package ibc_rate_limit

import (
	"encoding/json"

	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	channeltypes "github.com/cosmos/ibc-go/v5/modules/core/04-channel/types"
	porttypes "github.com/cosmos/ibc-go/v5/modules/core/05-port/types"
	"github.com/cosmos/ibc-go/v5/modules/core/exported"
	"github.com/umee-network/umee/v3/x/ibc-rate-limit/keeper"
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
	if err := im.app.OnChanCloseInit(ctx, portID, channelID); err != nil {
		return err
	}

	// TODO: do ourn own logic
	return nil
}

// OnChanCloseConfirm implements types.Middleware
func (im IBCMiddleware) OnChanCloseConfirm(ctx sdk.Context, portID string, channelID string) error {
	return im.app.OnChanCloseInit(ctx, portID, channelID)
}

// OnRecvPacket implements types.Middleware
func (im IBCMiddleware) OnRecvPacket(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) exported.Acknowledgement {
	panic("unimplemented")
}

// OnAcknowledgementPacket implements types.Middleware
func (im IBCMiddleware) OnAcknowledgementPacket(ctx sdk.Context, packet channeltypes.Packet, acknowledgement []byte, relayer sdk.AccAddress) error {
	var ack channeltypes.Acknowledgement
	if err := json.Unmarshal(acknowledgement, &ack); err != nil {
		return sdkerrors.Wrapf(sdkerrors.ErrUnknownRequest, "cannot unmarshal ICS-20 transfer packet acknowledgement: %v", err)
	}

	if !ack.Success() {
		// TODO: reverse sent the ibc transfer
	}

	return im.app.OnAcknowledgementPacket(ctx, packet, acknowledgement, relayer)
}

// OnTimeoutPacket implements types.Middleware
func (im IBCMiddleware) OnTimeoutPacket(ctx sdk.Context, packet channeltypes.Packet, relayer sdk.AccAddress) error {
	// TODO: do reverse logic
	return im.app.OnTimeoutPacket(ctx, packet, relayer)
}

// GetAppVersion implements types.Middleware
func (im IBCMiddleware) GetAppVersion(ctx sdk.Context, portID string, channelID string) (string, bool) {
	return im.keeper.GetAppVersion(ctx, portID, channelID)
}

// SendPacket implements types.Middleware
func (im IBCMiddleware) SendPacket(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet exported.PacketI) error {
	return im.keeper.SendPacket(ctx, chanCap, packet)
}

// WriteAcknowledgement implements types.Middleware
func (im IBCMiddleware) WriteAcknowledgement(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet exported.PacketI, ack exported.Acknowledgement) error {
	return im.keeper.WriteAcknowledgement(ctx, chanCap, packet, ack)
}
