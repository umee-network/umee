package tx

import (
	"fmt"
	"os"

	"github.com/CosmWasm/wasmd/x/wasm/ioutils"
	"github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/umee-network/umee/v5/app/params"
)

func (c *Client) TxSubmitWasmContract(contractPath string) (*sdk.TxResponse, error) {
	fromAddr, err := c.keyringRecord[0].GetAddress()
	if err != nil {
		return nil, err
	}

	msg, err := readWasmCode(contractPath, fromAddr)
	if err != nil {
		return nil, err
	}

	return c.BroadcastTx(&msg)
}

func (c *Client) TxWasmInstantiateContract(storeCode uint64, initMsg []byte) (*sdk.TxResponse, error) {
	fromAddr, err := c.keyringRecord[0].GetAddress()
	if err != nil {
		return nil, err
	}
	amount := sdk.NewCoins(sdk.NewCoin(appparams.BondDenom, sdk.NewInt(1)))
	msg := types.MsgInstantiateContract{
		Sender: fromAddr.String(),
		CodeID: storeCode,
		Label:  "label",
		Funds:  amount,
		Msg:    initMsg,
		Admin:  "",
	}

	return c.BroadcastTx(&msg)
}

func (c *Client) TxWasmExecuteContractByAccSeq(contractAddr string, execMsg []byte,
	accSeq uint64) (*sdk.TxResponse, error) {
	fromAddr, err := c.keyringRecord[0].GetAddress()
	if err != nil {
		return nil, err
	}
	amount := sdk.NewCoins(sdk.NewCoin(appparams.BondDenom, sdk.NewInt(1)))
	msg := types.MsgExecuteContract{
		Sender:   fromAddr.String(),
		Contract: contractAddr,
		Funds:    amount,
		Msg:      execMsg,
	}
	if accSeq != 0 {
		return c.BroadcastTxWithAccSeq(accSeq).BroadcastTxWithAsyncBlock().BroadcastTx(&msg)
	}
	return c.BroadcastTxWithAsyncBlock().BroadcastTx(&msg)
}

func (c *Client) TxWasmExecuteContract(contractAddr string, execMsg []byte) (*sdk.TxResponse, error) {
	return c.TxWasmExecuteContractByAccSeq(contractAddr, execMsg, 0)
}

// Prepares MsgStoreCode object from flags with gzipped wasm byte code field
func readWasmCode(file string, sender sdk.AccAddress) (types.MsgStoreCode, error) {
	wasm, err := os.ReadFile(file)
	if err != nil {
		return types.MsgStoreCode{}, err
	}

	// gzip the wasm file
	if ioutils.IsWasm(wasm) {
		wasm, err = ioutils.GzipIt(wasm)

		if err != nil {
			return types.MsgStoreCode{}, err
		}
	} else if !ioutils.IsGzip(wasm) {
		return types.MsgStoreCode{}, fmt.Errorf("invalid input file. Use wasm binary or gzip")
	}

	msg := types.MsgStoreCode{
		Sender:                sender.String(),
		WASMByteCode:          wasm,
		InstantiatePermission: &types.AllowEverybody,
	}
	return msg, nil
}
