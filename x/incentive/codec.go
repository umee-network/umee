package incentive

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	proto "github.com/gogo/protobuf/proto"
)

var (
	amino     = codec.NewLegacyAmino()
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
	cdc.RegisterConcrete(&MsgClaim{}, proto.MessageName(&MsgClaim{}), nil)
	cdc.RegisterConcrete(&MsgBond{}, proto.MessageName(&MsgBond{}), nil)
	cdc.RegisterConcrete(&MsgBeginUnbonding{}, proto.MessageName(&MsgBeginUnbonding{}), nil)
	cdc.RegisterConcrete(&MsgSponsor{}, proto.MessageName(&MsgSponsor{}), nil)
	cdc.RegisterConcrete(&MsgGovCreateProgram{}, proto.MessageName(&MsgGovCreateProgram{}), nil)
	cdc.RegisterConcrete(&MsgGovCreateAndSponsorProgram{}, proto.MessageName(&MsgGovCreateAndSponsorProgram{}), nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgClaim{},
		&MsgBond{},
		&MsgBeginUnbonding{},
		&MsgSponsor{},
		&MsgGovCreateProgram{},
		&MsgGovCreateAndSponsorProgram{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
