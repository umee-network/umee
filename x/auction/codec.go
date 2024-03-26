package auction

import (
	"github.com/cosmos/cosmos-sdk/codec"
	cdctypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	// "github.com/cosmos/cosmos-sdk/types/msgservice"
)

// Amino codecs
// Note, the ModuleCdc should ONLY be used in certain instances of tests and for JSON
// encoding as Amino is still used for that purpose.
var (
	amino     = codec.NewLegacyAmino()
	ModuleCdc = codec.NewAminoCodec(amino)
)

func init() {
	RegisterLegacyAminoCodec(amino)
	cryptocodec.RegisterCrypto(amino)
	amino.Seal()
}

// RegisterLegacyAminoCodec registers the necessary x/auction interfaces and
// concrete types on the provided LegacyAmino codec. These types are used for
// Amino JSON serialization.
func RegisterLegacyAminoCodec(cdc *codec.LegacyAmino) {
	cdc.RegisterConcrete(&MsgGovSetRewardsParams{}, "umee/auction/MsgGovSetRewardsParams", nil)
	cdc.RegisterConcrete(&MsgRewardsBid{}, "umee/auction/MsgRewardsBid", nil)
}

func RegisterInterfaces(registry cdctypes.InterfaceRegistry) {
	registry.RegisterImplementations(
		(*sdk.Msg)(nil),
		// &MsgGovSetRewardsParams{},
		// &MsgRewardsBid{},
	)

	// TODO
	// msgservice.RegisterMsgServiceDesc(registry, &_Msg_serviceDesc)
}
