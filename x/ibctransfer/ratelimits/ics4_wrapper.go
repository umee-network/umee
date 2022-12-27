package ratelimits

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	"github.com/cosmos/ibc-go/v5/modules/core/exported"
)

// GetAppVersion implements types.Middleware
func (im IBCMiddleware) GetAppVersion(ctx sdk.Context, portID string, channelID string) (string, bool) {
	return im.keeper.GetAppVersion(ctx, portID, channelID)
}

// SendPacket implements types.Middleware
func (im IBCMiddleware) SendPacket(ctx sdk.Context, chanCap *capabilitytypes.Capability,
	packet exported.PacketI,
) error {
	return im.keeper.SendPacket(ctx, chanCap, packet)
}

// WriteAcknowledgement implements types.Middleware
func (im IBCMiddleware) WriteAcknowledgement(ctx sdk.Context, chanCap *capabilitytypes.Capability,
	packet exported.PacketI, ack exported.Acknowledgement,
) error {
	return im.keeper.WriteAcknowledgement(ctx, chanCap, packet, ack)
}
