package tx

import (
	"fmt"
	"os"

	"github.com/CosmWasm/wasmd/x/wasm/ioutils"
	"github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/umee-network/umee/v4/app/params"
)

func (c *Client) TxSubmitWasmContract(contractPath string) (*sdk.TxResponse, error) {
	fromAddr, err := c.keyringRecord[0].GetAddress()
	if err != nil {
		return nil, err
	}

	msg, err := parseStoreCodeArgs(contractPath, fromAddr)
	if err != nil {
		return nil, err
	}

	return c.BroadcastTx(&msg)
}

func (c *Client) WasmInstantiateContract(storeCode uint64, initMsg string) (*sdk.TxResponse, error) {
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
		Msg:    []byte(initMsg),
		Admin:  "",
	}

	return c.BroadcastTx(&msg)
}

func (c *Client) WasmExecuteContract(contractAddr, execMsg string) (*sdk.TxResponse, error) {
	fromAddr, err := c.keyringRecord[0].GetAddress()
	if err != nil {
		return nil, err
	}
	amount := sdk.NewCoins(sdk.NewCoin(appparams.BondDenom, sdk.NewInt(1)))
	msg := types.MsgExecuteContract{
		Sender:   fromAddr.String(),
		Contract: contractAddr,
		Funds:    amount,
		Msg:      []byte(execMsg),
	}

	return c.BroadcastTx(&msg)
}

// Prepares MsgStoreCode object from flags with gzipped wasm byte code field
func parseStoreCodeArgs(file string, sender sdk.AccAddress) (types.MsgStoreCode, error) {
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
