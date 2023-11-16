package tx

import (
	"fmt"
	"os"

	"github.com/CosmWasm/wasmd/x/wasm/ioutils"
	"github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	appparams "github.com/umee-network/umee/v6/app/params"
	"github.com/umee-network/umee/v6/util/coin"
)

func (c *Client) WasmDeployContract(contractPath string) (*sdk.TxResponse, error) {
	fromIdx := 0
	msg, err := readWasmCode(contractPath, c.KeyringAddress(fromIdx))
	if err != nil {
		return nil, err
	}

	return c.BroadcastTx(fromIdx, &msg)
}

func (c *Client) WasmInitContract(storeCode uint64, initMsg []byte) (*sdk.TxResponse, error) {
	fromIdx := 0
	amount := sdk.NewCoins(sdk.NewCoin(appparams.BaseDenom, sdk.NewInt(1)))
	msg := types.MsgInstantiateContract{
		Sender: c.KeyringAddress(fromIdx).String(),
		CodeID: storeCode,
		Label:  "label",
		Funds:  amount,
		Msg:    initMsg,
		Admin:  "",
	}

	return c.BroadcastTx(fromIdx, &msg)
}

func (c *Client) WasmExecContractWithAccSeq(contractAddr string, execMsg []byte, accSeq uint64,
) (*sdk.TxResponse, error) {
	fromIdx := 0
	amount := sdk.NewCoins(coin.Umee1)
	msg := types.MsgExecuteContract{
		Sender:   c.KeyringAddress(fromIdx).String(),
		Contract: contractAddr,
		Funds:    amount,
		Msg:      execMsg,
	}
	if accSeq != 0 {
		return c.WithAccSeq(accSeq).BroadcastTx(fromIdx, &msg)
	}
	return c.BroadcastTx(fromIdx, &msg)
}

func (c *Client) WasmExecuteContract(contractAddr string, execMsg []byte) (*sdk.TxResponse, error) {
	return c.WasmExecContractWithAccSeq(contractAddr, execMsg, 0)
}

// Prepares MsgStoreCode object from flags with gzipped wasm byte code field
func readWasmCode(file string, sender sdk.AccAddress) (types.MsgStoreCode, error) {
	wasm, err := os.ReadFile(file)
	if err != nil {
		return types.MsgStoreCode{}, err
	}

	if ioutils.IsWasm(wasm) {
		wasm, err = ioutils.GzipIt(wasm)
		if err != nil {
			return types.MsgStoreCode{}, err
		}
	} else if !ioutils.IsGzip(wasm) {
		return types.MsgStoreCode{},
			fmt.Errorf("invalid input file. Wasm file must be a binary or gzip of a binary")
	}

	return types.MsgStoreCode{
		Sender:                sender.String(),
		WASMByteCode:          wasm,
		InstantiatePermission: &types.AllowEverybody,
	}, nil
}
