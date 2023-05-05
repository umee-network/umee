package keeper

import (
	"cosmossdk.io/errors"
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	ics20types "github.com/cosmos/ibc-go/v6/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v6/modules/core/02-client/types"
	ibcexported "github.com/cosmos/ibc-go/v6/modules/core/exported"
	"github.com/umee-network/umee/v4/x/uibc"

	ibcutil "github.com/umee-network/umee/v4/util/ibc"
)

/******
 * Implementation of ICS4Wrapper interface
 ******/

// SendPacket wraps IBC ChannelKeeper's SendPacket function to record quota outflows.
func (k Keeper) SendPacket(ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	sourcePort string,
	sourceChannel string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
	data []byte) (uint64, error) {

	params := k.GetParams(ctx)
	if params.IbcStatus == uibc.IBCTransferStatus_IBC_TRANSFER_STATUS_TRANSFERS_PAUSED {
		return 0, ics20types.ErrSendDisabled
	}

	funds, denom, err := ibcutil.GetFundsFromPacket(data)
	if err != nil {
		return 0, errors.Wrap(err, "bad packet in rate limit's SendPacket")
	}
	if params.IbcStatus == uibc.IBCTransferStatus_IBC_TRANSFER_STATUS_QUOTA_ENABLED {
		if err := k.CheckAndUpdateQuota(ctx, denom, funds); err != nil {
			return 0, errors.Wrap(err, "SendPacket over the IBC Quota")
		}
	}

	return k.ics4Wrapper.SendPacket(ctx, chanCap, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, data)
}

// WriteAcknowledgement wraps IBC ChannelKeeper's WriteAcknowledgement function
// ICS29 WriteAcknowledgement is used for asynchronous acknowledgements
func (k Keeper) WriteAcknowledgement(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet ibcexported.PacketI,
	acknowledgement ibcexported.Acknowledgement,
) error {
	// ics4Wrapper may be core IBC or higher-level middleware
	return k.ics4Wrapper.WriteAcknowledgement(ctx, chanCap, packet, acknowledgement)
}

// GetAppVersion returns the underlying application version.
func (k Keeper) GetAppVersion(ctx sdk.Context, portID, channelID string) (string, bool) {
	return k.ics4Wrapper.GetAppVersion(ctx, portID, channelID)
}
