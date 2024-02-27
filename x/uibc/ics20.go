package uibc

import (
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
	ics20types "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"
)

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (m ICS20Memo) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	return tx.UnpackInterfaces(unpacker, m.Messages)
}

// ExtractDenomFromPacketOnRecv takes a packet with a valid ICS20 token data in the Data field
// and returns the denom as represented in the local chain.
func ExtractDenomFromPacketOnRecv(packet ibcexported.PacketI, denom string) string {
	if ics20types.ReceiverChainIsSource(packet.GetSourcePort(), packet.GetSourceChannel(), denom) {
		// if we receive back a token, that was originally sent from UMEE, then we need to remove
		// prefix added by the sender chain: port/channel/base_denom -> base_denom.

		voucherPrefix := ics20types.GetDenomPrefix(packet.GetSourcePort(), packet.GetSourceChannel())
		unprefixedDenom := denom[len(voucherPrefix):]

		// coin denomination used in sending from the escrow address
		denom = unprefixedDenom

		// The denomination used to send the coins is either the native denom or the hash of the path
		// if the denomination is not native.
		denomTrace := ics20types.ParseDenomTrace(unprefixedDenom)
		if !denomTrace.IsNativeDenom() {
			denom = denomTrace.IBCDenom()
		}
	} else {
		prefixedDenom := ics20types.GetDenomPrefix(packet.GetDestPort(), packet.GetDestChannel()) + denom
		denom = ics20types.ParseDenomTrace(prefixedDenom).IBCDenom()
	}
	return denom
}
