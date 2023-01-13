package client

import (
	"fmt"

	sdk "github.com/cosmos/cosmos-sdk/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
)

// func (bt *BankTx) Send(fromAddress string, toAddress string, amount sdk.Coins) {
// 	msg := banktypes.MsgSend{
// 		FromAddress: fromAddress,
// 		ToAddress:   toAddress,
// 		Amount:      amount,
// 	}
// 	resp, err := BroadcastTx()
// }

func (tc *TxClient) Send(fromAddress string, toAddress string, amount sdk.Coins) {
	msg := &banktypes.MsgSend{
		FromAddress: fromAddress,
		ToAddress:   toAddress,
		Amount:      amount,
	}
	resp, err := tc.BroadcastTx(msg)
	fmt.Printf("%+v\n", resp)
	fmt.Println(err)
}
