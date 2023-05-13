package util

import (
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/std"
)

// NewCodec creates Codec instance and registers standard Cosmos SDK interfaces (message
// interface, tx, pub & private keys) as well as all types provided through the registrar (
// typically a RegisterInterfaces function in Cosmos SDK modules)
func NewCodec(registrars ...func(types.InterfaceRegistry)) codec.Codec {
	interfaceRegistry := types.NewInterfaceRegistry()
	std.RegisterInterfaces(interfaceRegistry)
	for _, f := range registrars {
		f(interfaceRegistry)
	}

	return codec.NewProtoCodec(interfaceRegistry)
}
