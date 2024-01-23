package sdkclient

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// BankSend creates and broadcasts bank send tx. `fromIdx` is an account index in the client
// keyring.
func (c *Client) BankSend(fromIdx int, toAddress string, amount sdk.Coins) (*sdk.TxResponse, error) {
	msg := &banktypes.MsgSend{
		FromAddress: c.KeyringAddress(fromIdx).String(),
		ToAddress:   toAddress,
		Amount:      amount,
	}
	return c.BroadcastTx(fromIdx, msg)
}
