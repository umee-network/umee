package tx

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

func (c *Client) TxSend(fromAddress, toAddress string, amount sdk.Coins) (*sdk.TxResponse, error) {
	msg := &banktypes.MsgSend{
		FromAddress: fromAddress,
		ToAddress:   toAddress,
		Amount:      amount,
	}
	return c.BroadcastTx(msg)
}
