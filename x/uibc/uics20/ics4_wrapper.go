package uics20

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	clienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	porttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"

	"github.com/umee-network/umee/v6/x/uibc/quota"
)

// ICS4 implements porttypes.ICS4Wrapper (middleware to send packets and acknowledgements) and
// overwrites SendPacket to check IBC quota.
type ICS4 struct {
	porttypes.ICS4Wrapper

	quotaKB quota.KeeperBuilder
}

func NewICS4(parent porttypes.ICS4Wrapper, kb quota.KeeperBuilder) ICS4 {
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
	k := q.quotaKB.Keeper(&ctx)
	if err := k.IBCOnSendPacket(data); err != nil {
		return 0, err
	}
	return q.ICS4Wrapper.SendPacket(ctx, chanCap, sourcePort, sourceChannel, timeoutHeight, timeoutTimestamp, data)
}
