package uibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/x/bank/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	ibcexported "github.com/cosmos/ibc-go/v5/modules/core/exported"

	leveragetypes "github.com/umee-network/umee/v3/x/leverage/types"
)

// BankKeeper defines the expected x/bank keeper interface.
type BankKeeper interface {
	GetDenomMetaData(ctx sdk.Context, denom string) (types.Metadata, bool)
	SetDenomMetaData(ctx sdk.Context, denomMetaData types.Metadata)
	IterateAllDenomMetaData(ctx sdk.Context, cb func(types.Metadata) bool)
}

// ICS4Wrapper defines the expected ICS4Wrapper for middleware
type ICS4Wrapper interface {
	WriteAcknowledgement(ctx sdk.Context, chanCap *capabilitytypes.Capability, packet ibcexported.PacketI,
		acknowledgement ibcexported.Acknowledgement,
	) error
	SendPacket(ctx sdk.Context, channelCap *capabilitytypes.Capability, packet ibcexported.PacketI) error
	GetAppVersion(ctx sdk.Context, portID, channelID string) (string, bool)
}

type OracleKeeper interface {
	GetExchangeRate(ctx sdk.Context, symbol string) (sdk.Dec, error)
}

type LeverageKeeper interface {
	GetTokenSettings(ctx sdk.Context, baseDenom string) (leveragetypes.Token, error)
}
