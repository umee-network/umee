package tsdk

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/testutil"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
	"github.com/umee-network/umee/v6/app/params"
)

// NewCodec creates Codec instance and registers standard Cosmos SDK interfaces (message
// interface, tx, pub & private keys) as well as all types provided through the registrar (
// typically a RegisterInterfaces function in Cosmos SDK modules)
func NewCodec(registrars ...func(types.InterfaceRegistry)) codec.Codec {
	interfaceRegistry := testutil.CodecOptions{
		AccAddressPrefix: params.AccountAddressPrefix,
		ValAddressPrefix: params.ValidatorAddressPrefix,
	}.NewInterfaceRegistry()
	std.RegisterInterfaces(interfaceRegistry) // register SDK interfaces
	for _, f := range registrars {
		f(interfaceRegistry)
	}

	return codec.NewProtoCodec(interfaceRegistry)
}
