package ugov

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

// Amino codecs
// Note, the ModuleCdc should ONLY be used in certain instances of tests and for JSON
// encoding as Amino is still used for that purpose.
var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(amino) //nolint
)

func init() {
	cryptocodec.RegisterCrypto(amino)
	sdk.RegisterLegacyAminoCodec(amino)
	RegisterLegacyAminoCodec(amino)

	amino.Seal()
}

// RegisterLegacyAminoCodec registers the necessary x/ugov interfaces and
// concrete types on the provided LegacyAmino codec. These types are used for
// Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgGovUpdateMinGasPrice{}, "umee/ugov/MsgGovUpdateMinGasPrice", nil)
	cdc.RegisterConcrete(&MsgGovSetEmergencyGroup{}, "umee/ugov/MsgGovSetEmergencyGroup", nil)
	cdc.RegisterConcrete(&MsgGovUpdateInflationParams{}, "umee/ugov/MsgGovUpdateInflationParams", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgGovUpdateMinGasPrice{},
		&MsgGovSetEmergencyGroup{},
		&MsgGovUpdateInflationParams{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
