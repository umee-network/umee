package util

import (
	"encoding/json"
	"strconv"
	"testing"

	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"gotest.tools/v3/assert"

	"github.com/umee-network/umee/v6/client"
	"github.com/umee-network/umee/v6/tests/grpc"
)

const (
	SucceessRespCode = uint32(0)
)

type Cosmwasm struct {
	StoreCode    uint64
	ContractAddr string
	Sender       string
	T            *testing.T
	umee         client.Client
}

func NewCosmwasmTestSuite(t *testing.T, umee client.Client) *Cosmwasm {
	return &Cosmwasm{
		T:    t,
		umee: umee,
	}
}

func (cw *Cosmwasm) DeployWasmContract(path string) {
	cw.T.Logf("ℹ️ deploying smart contract %s", path)
	resp, err := cw.umee.Tx.WasmDeployContract(path)
	assert.NilError(cw.T, err)
	resp, err = grpc.GetTxResponse(cw.umee, resp.TxHash, 1)
	assert.NilError(cw.T, err)
	storeCode := cw.GetAttributeValue(*resp, "store_code", "code_id")
	cw.StoreCode, err = strconv.ParseUint(storeCode, 10, 64)
	assert.NilError(cw.T, err)
	cw.T.Logf("✅ smart contract is deployed and store code is %d", cw.StoreCode)
}

func (cw *Cosmwasm) MarshalAny(any interface{}) []byte {
	data, err := json.Marshal(any)
	assert.NilError(cw.T, err)
	return data
}

func (cw *Cosmwasm) InstantiateContract(initMsg []byte) {
	cw.T.Log("ℹ️ smart contract is instantiating...")
	resp, err := cw.umee.Tx.WasmInitContract(cw.StoreCode, initMsg)
	assert.NilError(cw.T, err)
	resp, err = grpc.GetTxResponse(cw.umee, resp.TxHash, 1)
	assert.NilError(cw.T, err)
	cw.ContractAddr = cw.GetAttributeValue(*resp, "instantiate", "_contract_address")
	assert.Equal(cw.T, SucceessRespCode, resp.Code)
	cw.T.Log("✅ smart contract is instantiating is done.")
	cw.T.Logf("smart contract address is %s", cw.ContractAddr)
}

func (cw *Cosmwasm) CWQuery(query []byte) wasmtypes.QuerySmartContractStateResponse {
	resp, err := cw.umee.QueryContract(cw.ContractAddr, query)
	assert.NilError(cw.T, err)
	return *resp
}

func (cw *Cosmwasm) CWExecute(execMsg []byte) {
	resp, err := cw.umee.Tx.WasmExecuteContract(cw.ContractAddr, execMsg)
	assert.NilError(cw.T, err)
	assert.Equal(cw.T, SucceessRespCode, resp.Code)
}

func (cw *Cosmwasm) CWExecuteWithSeqAndAsync(execMsg []byte, accSeq uint64) {
	resp, err := cw.umee.Tx.WasmExecContractWithAccSeq(cw.ContractAddr, execMsg, accSeq)
	assert.NilError(cw.T, err)
	assert.Equal(cw.T, SucceessRespCode, resp.Code)
}

func (cw *Cosmwasm) CWExecuteWithSeqAndAsyncResp(execMsg []byte, accSeq uint64) (*sdk.TxResponse, error) {
	return cw.umee.Tx.WasmExecContractWithAccSeq(cw.ContractAddr, execMsg, accSeq)
}

func (cw *Cosmwasm) GetAttributeValue(resp sdk.TxResponse, eventName, attrKey string) string {
	var attrVal string
	for _, event := range resp.Logs[0].Events {
		if event.Type == eventName {
			for _, attribute := range event.Attributes {
				if attribute.Key == attrKey {
					attrVal = attribute.Value
					break
				}
			}
		}
	}
	return attrVal
}
