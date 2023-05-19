package e2e

import (
	"encoding/json"
	"strconv"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/suite"
)

type TestCosmwasm struct {
	IntegrationTestSuite
	StoreCode    uint64
	ContractAddr string
	Sender       string
}

func TestCosmwasmSuite(t *testing.T) {
	suite.Run(t, new(TestCosmwasm))
}

// TODO: re-enable this tests when we do dockernet integration
// func (cw *TestCosmwasm) TestCW20() {
// 	// TODO: needs to add contracts
// 	accAddr, err := cw.chain.validators[0].keyInfo.GetAddress()
// 	cw.Require().NoError(err)
// 	cw.Sender = accAddr.String()

// 	// path := ""
// 	path := "/Users/gsk967/Projects/umee-network/umee-cosmwasm/artifacts/umee_cosmwasm-aarch64.wasm"
// 	cw.DeployWasmContract(path)

// 	// InstantiateContract
// 	cw.InstantiateContract()

// 	// execute contract
// 	tx := "{\"umee\":{\"leverage\":{\"supply_collateral\":{\"supplier\":\"umee19ppq83qzzy3f0fftdp2p3t5eyp44nm33we37n3\",\"asset\":{\"amount\":\"1000\",\"denom\":\"uumee\"}}}}}"
// 	cw.CWExecute(tx)

// 	// query the contract
// 	query := "{\"chain\":{\"custom\":{\"leverage_params\":{},\"assigned_query\":\"0\"}}}"
// 	cw.CWQuery(query)
// 	cw.Require().False(true)
// }

func (cw *TestCosmwasm) DeployWasmContract(path string) {
	resp, err := cw.umee.Tx.TxSubmitWasmContract(path)
	cw.Require().NoError(err)
	storeCode := cw.GetAttributeValue(*resp, "store_code", "code_id")
	cw.StoreCode, err = strconv.ParseUint(storeCode, 10, 64)
	cw.Require().NoError(err)
}

func (cw *TestCosmwasm) MarshalAny(any interface{}) []byte {
	data, err := json.Marshal(any)
	cw.Require().NoError(err)
	return data
}

func (cw *TestCosmwasm) InstantiateContract() {
	resp, err := cw.umee.Tx.WasmInstantiateContract(cw.StoreCode, []byte("{}"))
	cw.ContractAddr = cw.GetAttributeValue(*resp, "wasm", "_contract_address")
	cw.Require().NoError(err)
}

func (cw *TestCosmwasm) CWQuery(query string) {
	_, err := cw.umee.QueryContract(cw.ContractAddr, []byte(query))
	cw.Require().NoError(err)
}

func (cw *TestCosmwasm) CWExecute(execMsg string) {
	_, err := cw.umee.Tx.WasmExecuteContract(cw.ContractAddr, execMsg)
	cw.Require().NoError(err)
}

func (cw *TestCosmwasm) GetAttributeValue(resp sdk.TxResponse, eventName, attrKey string) string {
	var attrVal string
	for _, event := range resp.Logs[0].Events {
		if event.Type == eventName {
			for _, attribute := range event.Attributes {
				if attribute.Key == attrKey {
					attrVal = attribute.Value
				}
			}
		}
	}
	return attrVal
}
