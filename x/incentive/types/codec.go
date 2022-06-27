package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
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

// RegisterLegacyAminoCodec registers the necessary x/incentive interfaces and
// concrete types on the provided LegacyAmino codec. These types are used for
// Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgLock{}, "umee/incentive/MsgLock", nil)
	cdc.RegisterConcrete(&MsgUnlock{}, "umee/incentive/MsgUnlock", nil)
	cdc.RegisterConcrete(&MsgClaim{}, "umee/incentive/MsgClaim", nil)
	cdc.RegisterConcrete(&MsgSponsor{}, "umee/incentive/MsgSponsor", nil)
	cdc.RegisterConcrete(&CreateIncentiveProgramProposal{}, "umee/incentive/CreateIncentiveProgramProposal", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgLock{},
		&MsgUnlock{},
		&MsgClaim{},
		&MsgSponsor{},
	)

	registry.RegisterImplementations(
		(*govtypes.Content)(nil),
		&CreateIncentiveProgramProposal{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
