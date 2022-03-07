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

// RegisterLegacyAminoCodec registers the necessary x/leverage interfaces and
// concrete types on the provided LegacyAmino codec. These types are used for
// Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgLendAsset{}, "umee/leverage/MsgLendAsset", nil)
	cdc.RegisterConcrete(&MsgWithdrawAsset{}, "umee/leverage/MsgWithdrawAsset", nil)
	cdc.RegisterConcrete(&UpdateRegistryProposal{}, "umee/leverage/UpdateRegistryProposal", nil)
	cdc.RegisterConcrete(&MsgSetCollateral{}, "umee/leverage/MsgSetCollateral", nil)
	cdc.RegisterConcrete(&MsgBorrowAsset{}, "umee/leverage/MsgBorrowAsset", nil)
	cdc.RegisterConcrete(&MsgRepayAsset{}, "umee/leverage/MsgRepayAsset", nil)
	cdc.RegisterConcrete(&MsgLiquidate{}, "umee/leverage/MsgLiquidate", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgLendAsset{},
		&MsgWithdrawAsset{},
		&MsgSetCollateral{},
		&MsgBorrowAsset{},
		&MsgRepayAsset{},
		&MsgLiquidate{},
	)

	registry.RegisterImplementations(
		(*govtypes.Content)(nil),
		&UpdateRegistryProposal{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
