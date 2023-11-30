package uibc

import (
	"cosmossdk.io/errors"
	ics20types "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	porttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"

	ibcutil "github.com/umee-network/umee/v6/util/ibc"
	"github.com/umee-network/umee/v6/x/uibc/quota"
)

// ICS4 implements porttypes.ICS4Wrapper and overwrites SendPacket to check IBC quota.
type ICS4 struct {
	porttypes.ICS4Wrapper

	kb quota.Builder
}

func NewICS4(parent porttypes.ICS4Wrapper, kb quota.Builder) ICS4 {
	return ICS4{parent, kb}
}

// SendPacket implements types.Middleware
func (q ICS4) SendPacket(
	ctx sdk.Context,
	chanCap *capabilitytypes.Capability,
	sourcePort string,
	sourceChannel string,
	timeoutHeight clienttypes.Height,
	timeoutTimestamp uint64,
	data []byte,
) (uint64, error) {
	k := q.kb.Keeper(&ctx)
	params := k.GetParams()

	if !params.IbcStatus.IBCTransferEnabled() {
		return 0, ics20types.ErrSendDisabled
	}

	funds, denom, err := ibcutil.GetFundsFromPacket(data)
	if err != nil {
		return 0, errors.Wrap(err, "bad packet in rate limit's SendPacket")
	}

	if params.IbcStatus.OutflowQuotaEnabled() {
		if err := k.CheckAndUpdateQuota(denom, funds); err != nil {
			return 0, errors.Wrap(err, "sendPacket over the IBC Quota")
		}
	}

	return q.ICS4Wrapper.SendPacket(ctx, chanCap, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, data)
}
