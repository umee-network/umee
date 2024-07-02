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

	// ModuleCdc references the global x/leverage module codec. Note, Amino
	// is required for ledger signing of messages, and Kepler signing.
	ModuleCdc = codec.NewAminoCodec(amino) //nolint
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
	cdc.RegisterConcrete(&MsgCollateralize{}, "umee/leverage/MsgCollateralize", nil)
	cdc.RegisterConcrete(&MsgDecollateralize{}, "umee/leverage/MsgDecollateralize", nil)
	cdc.RegisterConcrete(&MsgBorrow{}, "umee/leverage/MsgBorrow", nil)
	cdc.RegisterConcrete(&MsgRepay{}, "umee/leverage/MsgRepay", nil)
	cdc.RegisterConcrete(&MsgLiquidate{}, "umee/leverage/MsgLiquidate", nil)
	cdc.RegisterConcrete(&MsgSupplyCollateral{}, "umee/leverage/MsgSupplyCollateral", nil)
	cdc.RegisterConcrete(&MsgMaxWithdraw{}, "umee/leverage/MsgMaxWithdraw", nil)
	cdc.RegisterConcrete(&MsgMaxBorrow{}, "umee/leverage/MsgMaxBorrow", nil)
	cdc.RegisterConcrete(&MsgLeveragedLiquidate{}, "umee/leverage/MsgLeveragedLiquidate", nil)

	cdc.RegisterConcrete(&MsgGovUpdateRegistry{}, "umee/leverage/MsgGovUpdateRegistry", nil)
	cdc.RegisterConcrete(&MsgGovSetParams{}, "umee/leverage/MsgGovSetParams", nil)
	cdc.RegisterConcrete(&MsgGovUpdateSpecialAssets{}, "umee/leverage/MsgGovUpdateSpecialAssets", nil)
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
		&MsgSupplyCollateral{},
		&MsgMaxWithdraw{},
		&MsgMaxBorrow{},
		&MsgLeveragedLiquidate{},

		&MsgGovUpdateRegistry{},
		&MsgGovUpdateSpecialAssets{},
		&MsgGovSetParams{},
	)

	msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
