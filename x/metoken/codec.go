package metoken

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

var (
	amino = codec.NewLegacyAmino()

	// ModuleCdc references the global x/metoken module codec. Note, the codec
	// should ONLY be used in certain instances of tests and for JSON encoding as
	// Amino is still used for that purpose.
	ModuleCdc = codec.NewAminoCodec(amino) //nolint
)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	amino.Seal()
}

// RegisterLegacyAminoCodec registers the necessary x/metoken interfaces and
// concrete types on the provided LegacyAmino codec. These types are used for
// Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgGovSetParams{}, "umee/metoken/MsgGovSetParams", nil)
	cdc.RegisterConcrete(&MsgGovUpdateRegistry{}, "umee/metoken/MsgGovUpdateRegistry", nil)
	cdc.RegisterConcrete(&MsgSwap{}, "umee/metoken/MsgSwap", nil)
	cdc.RegisterConcrete(&MsgRedeem{}, "umee/metoken/MsgRedeem", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgGovSetParams{},
		&MsgGovUpdateRegistry{},
		&MsgSwap{},
		&MsgRedeem{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
