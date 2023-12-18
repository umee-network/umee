package uibc

import (
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (m ICS20Memo) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	return tx.UnpackInterfaces(unpacker, m.Messages)
}
