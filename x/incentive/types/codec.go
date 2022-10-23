package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
)

var (
	amino = codec.NewLegacyAmino()

	// ModuleCdc references the global x/incentive module codec. Note, the codec
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
	cdc.RegisterConcrete(&MsgClaim{}, "umee/incentive/MsgClaim", nil)
	cdc.RegisterConcrete(&MsgLock{}, "umee/incentive/MsgLock", nil)
	cdc.RegisterConcrete(&MsgBeginUnbonding{}, "umee/incentive/MsgBeginUnbonding", nil)
	cdc.RegisterConcrete(&MsgSponsor{}, "umee/incentive/MsgSponsor", nil)
	cdc.RegisterConcrete(&MsgCreateProgram{}, "umee/incentive/MsgCreateProgram", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgClaim{},
		&MsgLock{},
		&MsgBeginUnbonding{},
		&MsgSponsor{},
		&MsgCreateProgram{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
