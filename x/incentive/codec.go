package incentive

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
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

// RegisterLegacyAminoCodec registers the necessary x/incentive interfaces and
// concrete types on the provided LegacyAmino codec. These types are used for
// Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgClaim{}, "umee/incentive/MsgClaim", nil)
	cdc.RegisterConcrete(&MsgBond{}, "umee/incentive/MsgBond", nil)
	cdc.RegisterConcrete(&MsgBeginUnbonding{}, "umee/incentive/MsgBeginUnbonding", nil)
	cdc.RegisterConcrete(&MsgSponsor{}, "umee/incentive/MsgSponsor", nil)
	cdc.RegisterConcrete(&MsgGovSetParams{}, "umee/incentive/MsgGovSetParams", nil)
	cdc.RegisterConcrete(&MsgGovCreatePrograms{}, "umee/incentive/MsgGovCreatePrograms", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgClaim{},
		&MsgBond{},
		&MsgBeginUnbonding{},
		&MsgSponsor{},
		&MsgGovSetParams{},
		&MsgGovCreatePrograms{},
	)

	registry.RegisterImplementations(
		(*govtypes.Content)(nil),
		&MsgGovSetParams{},
		&MsgGovCreatePrograms{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
