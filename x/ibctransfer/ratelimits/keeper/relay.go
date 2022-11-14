package keeper

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	ibcexported "github.com/cosmos/ibc-go/v5/modules/core/exported"

	"github.com/umee-network/umee/v3/x/ibctransfer/types"
)

// SendPacket wraps IBC ChannelKeeper's SendPacket function
func (k Keeper) SendPacket(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet ibcexported.PacketI) error {
	return k.ics4Wrapper.SendPacket(ctx, chanCap, packet)
}

// WriteAcknowledgement wraps IBC ChannelKeeper's WriteAcknowledgement function
// ICS29 WriteAcknowledgement is used for asynchronous acknowledgements
func (k Keeper) WriteAcknowledgement(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet ibcexported.PacketI, acknowledgement ibcexported.Acknowledgement) error {
	// ics4Wrapper may be core IBC or higher-level middleware
	return k.ics4Wrapper.WriteAcknowledgement(ctx, chanCap, packet, acknowledgement)
}

// GetAppVersion returns the underlying application version.
func (k Keeper) GetAppVersion(ctx sdk.Context, portID, channelID string) (string, bool) {
	version, found := k.ics4Wrapper.GetAppVersion(ctx, portID, channelID)
	if !found {
		return "", false
	}

	var metadata types.Metadata
	if err := types.ModuleCdc.UnmarshalJSON([]byte(version), &metadata); err != nil {
		panic(fmt.Errorf("unable to unmarshal metadata for fee enabled channel: %w", err))
	}

	return metadata.AppVersion, true
}
