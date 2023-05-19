package cw_test

import (
	"encoding/json"
	"os"
	"strconv"
	"testing"

	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"gotest.tools/v3/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v4/app"
	"github.com/umee-network/umee/v4/client"
	"github.com/umee-network/umee/v4/tests/qa/cw"
)

const (
	cwBaseTransferPath = "../../artifacts/cw4_group-aarch64.wasm"
)

var (
	SucceessRespCode = uint32(0)
	TotalAccs        = 1000
)

func TestCWTransfer(t *testing.T) {
	privateKeys := make([]*secp256k1.PrivKey, 0)
	accAddrs := make([]sdk.AccAddress, 0)
	for i := 0; i < TotalAccs; i++ {
		privateKey := secp256k1.GenPrivKey()
		privateKeys = append(privateKeys, privateKey)
		pubKey := privateKey.PubKey()
		accAddrs = append(accAddrs, sdk.AccAddress(pubKey.Address()))
	}
	// remove if old keyring exists for testing
	os.RemoveAll("./keyring-test")
	encConfig := app.MakeEncodingConfig()
	cc, err := cw.ReadConfig("./config_example.yaml")
	assert.NilError(t, err)
	client, err := client.NewClient(cc.ChainID, cc.RPC, cc.GRPC, cc.Mnemonics, 1.5, encConfig)
	assert.NilError(t, err)

	resp, err := client.Tx.TxSubmitWasmContract(cwBaseTransferPath)
	assert.NilError(t, err)
	assert.Equal(t, SucceessRespCode, resp.Code)
	respStoreCode := cw.GetAttributeValue(*resp, "store_code", "code_id")
	assert.Equal(t, uint32(0), resp.Code)
	storeCode, err := strconv.ParseUint(respStoreCode, 10, 64)
	assert.NilError(t, err)

	admin, err := client.Tx.KeyringRecord[0].GetAddress()
	assert.NilError(t, err)

	// instantiate Contract
	initMsg := InitMsg{
		Admin:   admin.String(),
		Members: []Member{{Addr: admin.String(), Weight: 1}},
	}

	for i := 0; i < TotalAccs; i++ {
		initMsg.Members = append(initMsg.Members, Member{
			Addr:   accAddrs[i].String(),
			Weight: 1,
		})
	}

	msg, err := json.Marshal(initMsg)
	assert.NilError(t, err)
	initResp, err := client.Tx.WasmInstantiateContract(storeCode, msg)
	assert.NilError(t, err)
	assert.Equal(t, SucceessRespCode, initResp.Code)
	contractAddr := cw.GetAttributeValue(*initResp, "instantiate", "_contract_address")

	// query the contract
	cwGroupQuery := CWGroupQuery{
		Admin:       nil,
		Hooks:       nil,
		ListMembers: &ListMembers{},
	}

	queryMsg, err := json.Marshal(cwGroupQuery)
	assert.NilError(t, err)
	_, err = client.QueryContract(contractAddr, queryMsg)
	assert.NilError(t, err)
	// assert.Equal(t, true, false)
}
