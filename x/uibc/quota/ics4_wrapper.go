package quota

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	"github.com/cosmos/ibc-go/v6/modules/core/exported"
)

// GetAppVersion implements types.Middleware
func (im ICS20Middleware) GetAppVersion(ctx sdk.Context, portID string, channelID string) (string, bool) {
	return im.kb.GetAppVersion(ctx, portID, channelID)
}

// SendPacket implements types.Middleware
func (im ICS20Middleware) SendPacket(
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	sourcePort string,
	sourceChannel string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
	data []byte,
) (uint64, error) {
	return im.kb.SendPacket(ctx, chanCap, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, data)
}

// WriteAcknowledgement implements types.Middleware
func (im ICS20Middleware) WriteAcknowledgement(ctx sdk.Context, chanCap *capabilitytypes.Capability,
	packet exported.PacketI, ack exported.Acknowledgement,
) error {
	return im.kb.WriteAcknowledgement(ctx, chanCap, packet, ack)
}
