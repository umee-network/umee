package app

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/std"
	mtestuti "github.com/cosmos/cosmos-sdk/types/module/testutil"
	"github.com/cosmos/cosmos-sdk/x/auth/tx"
	"github.com/umee-network/umee/v6/app/params"
)

// MakeEncodingConfig returns the application's encoding configuration with all
// types and interfaces registered.
func MakeEncodingConfig() mtestuti.TestEncodingConfig {
	interfaceRegistry := testutil.CodecOptions{
		AccAddressPrefix: params.AccountAddressPrefix,
		ValAddressPrefix: params.ValidatorAddressPrefix,
	}.NewInterfaceRegistry()

	appCodec := codec.NewProtoCodec(interfaceRegistry)
	aminoCodec := codec.NewLegacyAmino()

	// cosmos-sdk module
	std.RegisterLegacyAminoCodec(aminoCodec)
	std.RegisterInterfaces(interfaceRegistry)

	// umee app modules
	ModuleBasics.RegisterLegacyAminoCodec(aminoCodec)
	ModuleBasics.RegisterInterfaces(interfaceRegistry)

	encCfg := mtestuti.TestEncodingConfig{
		InterfaceRegistry: interfaceRegistry,
		Codec:             appCodec,
		TxConfig:          tx.NewTxConfig(appCodec, tx.DefaultSignModes),
		Amino:             aminoCodec,
	}
	return encCfg
}
