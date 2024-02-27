package uibc

import (
	"github.com/cosmos/cosmos-sdk/codec/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/tx"
)

// UnpackInterfaces implements UnpackInterfacesMessage.UnpackInterfaces
func (m ICS20Memo) UnpackInterfaces(unpacker types.AnyUnpacker) error {
	return tx.UnpackInterfaces(unpacker, m.Messages)
}

// GetMsgs unpacks messages into []sdk.Msg
func (m ICS20Memo) GetMsgs() ([]sdk.Msg, error) {
	return tx.GetMsgs(m.Messages, "memo messages")
}
