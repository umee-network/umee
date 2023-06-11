package types

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/msgservice"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
)

var (
	amino = codec.NewLegacyAmino()

	// ModuleCdc references the global x/refileverage module codec. Note, Amino
	// is required for ledger signing of messages, and Kepler signing.
	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	amino.Seal()
}

// RegisterLegacyAminoCodec registers the necessary x/refileverage interfaces and
// concrete types on the provided LegacyAmino codec. These types are used for
// Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgDecollateralize{}, "umee/refileverage/MsgDecollateralize", nil)
	cdc.RegisterConcrete(&MsgBorrow{}, "umee/refileverage/MsgBorrow", nil)
	cdc.RegisterConcrete(&MsgRepay{}, "umee/refileverage/MsgRepay", nil)
	cdc.RegisterConcrete(&MsgLiquidate{}, "umee/refileverage/MsgLiquidate", nil)
	cdc.RegisterConcrete(&MsgGovUpdateRegistry{}, "umee/refileverage/MsgGovUpdateRegistry", nil)
	cdc.RegisterConcrete(&MsgSupplyCollateral{}, "umee/refileverage/MsgSupplyCollateral", nil)
	cdc.RegisterConcrete(&MsgMaxWithdraw{}, "umee/refileverage/MsgMaxWithdraw", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		&MsgDecollateralize{},
		&MsgBorrow{},
		&MsgRepay{},
		&MsgLiquidate{},
		&MsgGovUpdateRegistry{},
		&MsgSupplyCollateral{},
		&MsgMaxWithdraw{},
	)

	registry.RegisterImplementations(
		(*govtypes.Content)(nil),
		&MsgGovUpdateRegistry{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
