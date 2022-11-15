package ratelimits

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	sdkerrors "github.com/cosmos/cosmos-sdk/types/errors"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	"github.com/cosmos/ibc-go/v5/modules/core/exported"
)

// GetAppVersion implements types.Middleware
func (im IBCMiddleware) GetAppVersion(ctx sdk.Context, portID string, channelID string) (string, bool) {
	return im.keeper.GetAppVersion(ctx, portID, channelID)
}

// SendPacket implements types.Middleware
func (im IBCMiddleware) SendPacket(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet exported.PacketI) error {
	amount, denom, err := im.keeper.GetFundsFromPacket(packet)
	if err != nil {
		return sdkerrors.Wrap(err, "bad packet in rate limit's SendPacket")
	}

	if err := im.keeper.CheckAndUpdateRateLimits(ctx, denom, amount); err != nil {
		return sdkerrors.Wrap(err, "bad packet in rate limit's SendPacket")
	}

	return im.keeper.SendPacket(ctx, chanCap, packet)
}

// WriteAcknowledgement implements types.Middleware
func (im IBCMiddleware) WriteAcknowledgement(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet exported.PacketI, ack exported.Acknowledgement) error {
	return im.keeper.WriteAcknowledgement(ctx, chanCap, packet, ack)
}
