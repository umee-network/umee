package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	gov1b1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

var (
	amino = codec.NewLegacyAmino()

	// ModuleCdc references the global x/leverage module codec. Note, the codec
	// should ONLY be used in certain instances of tests and for JSON encoding as
	// Amino is still used for that purpose.
	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	amino.Seal()
}

// RegisterLegacyAminoCodec registers the necessary x/leverage interfaces and
// concrete types on the provided LegacyAmino codec. These types are used for
// Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgSupply{}, "umee/leverage/MsgSupply", nil)
	cdc.RegisterConcrete(&MsgWithdraw{}, "umee/leverage/MsgWithdraw", nil)
	cdc.RegisterConcrete(&UpdateRegistryProposal{}, "umee/leverage/UpdateRegistryProposal", nil)
	cdc.RegisterConcrete(&MsgCollateralize{}, "umee/leverage/MsgCollateralize", nil)
	cdc.RegisterConcrete(&MsgDecollateralize{}, "umee/leverage/MsgDecollateralize", nil)
	cdc.RegisterConcrete(&MsgBorrow{}, "umee/leverage/MsgBorrow", nil)
	cdc.RegisterConcrete(&MsgRepay{}, "umee/leverage/MsgRepay", nil)
	cdc.RegisterConcrete(&MsgLiquidate{}, "umee/leverage/MsgLiquidate", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgSupply{},
		&MsgWithdraw{},
		&MsgCollateralize{},
		&MsgDecollateralize{},
		&MsgBorrow{},
		&MsgRepay{},
		&MsgLiquidate{},
	)

	registry.RegisterImplementations(
		(*gov1b1.Content)(nil),
		&UpdateRegistryProposal{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
