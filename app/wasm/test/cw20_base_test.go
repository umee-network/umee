package wasm_test

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"cosmossdk.io/math"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	wasmvmtypes "github.com/CosmWasm/wasmvm/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	"gotest.tools/v3/assert"

	umeeapp "github.com/umee-network/umee/v5/app"
	appparams "github.com/umee-network/umee/v5/app/params"
	wm "github.com/umee-network/umee/v5/app/wasm/msg"
	wq "github.com/umee-network/umee/v5/app/wasm/query"
	"github.com/umee-network/umee/v5/x/oracle/types"
)

const (
	initialPower = int64(10000000000000)
	cw20Artifact = "../../../tests/artifacts/cw20_base.wasm"
	umeeArtifact = "../../../tests/artifacts/umee_cosmwasm-aarch64.wasm"
	cw20Label    = "cw20InstanceTest"
)

// Test addresses
var (
	valPubKeys = simapp.CreateTestPubKeys(2)

	valPubKey = valPubKeys[0]
	pubKey    = secp256k1.GenPrivKey().PubKey()
	addr      = sdk.AccAddress(pubKey.Address())
	valAddr   = sdk.ValAddress(pubKey.Address())

	valPubKey2 = valPubKeys[1]
	pubKey2    = secp256k1.GenPrivKey().PubKey()
	addr2      = sdk.AccAddress(pubKey2.Address())
	valAddr2   = sdk.ValAddress(pubKey2.Address())

	initTokens = sdk.TokensFromConsensusPower(initialPower, sdk.DefaultPowerReduction)
	initCoins  = sdk.NewCoins(sdk.NewCoin(appparams.BondDenom, initTokens))
)

type cw20InitMsg struct {
	Name            string    `json:"name"`
	Symbol          string    `json:"symbol"`
	Decimals        uint8     `json:"decimals"`
	InitialBalances []Balance `json:"initial_balances"`
}

type InstantiateMsg struct {
}

type Address struct {
	Address string `json:"address"`
}

type Balance struct {
	Address
	Amount uint64 `json:"amount,string"`
}

type cw20QueryBalance struct {
	Balance struct {
		Address
	} `json:"balance"`
}

type cw20QueryBalanceResp struct {
	Balance string `json:"balance"`
}

type cw20ExecMsg struct {
	Transfer *transferMsg `json:"transfer,omitempty"`
}

type transferMsg struct {
	Recipient string `json:"recipient"`
	Amount    uint64 `json:"amount,string"`
}

type CustomQuery struct {
	Chain struct {
		Custom wq.UmeeQuery `json:"custom"`
	} `json:"chain"`
}

type StargateQuery struct {
	Chain struct {
		Stargate wasmvmtypes.StargateQuery `json:"stargate"`
	} `json:"chain"`
}

type ExecuteMsg struct {
	Umee struct {
		Leverage wm.UmeeMsg `json:"leverage"`
	} `json:"umee"`
}

func UmeeCwCustomQuery(umeeCWQuery wq.UmeeQuery) CustomQuery {
	c := CustomQuery{}
	c.Chain.Custom = umeeCWQuery
	return c
}

func UmeeCwCustomTx(customMsg wm.UmeeMsg) ExecuteMsg {
	c := ExecuteMsg{}
	c.Umee.Leverage = customMsg
	return c
}

type IntegrationTestSuite struct {
	T   *testing.T
	ctx sdk.Context
	app *umeeapp.UmeeApp

	wasmMsgServer       wasmtypes.MsgServer
	wasmQueryClient     wasmtypes.QueryClient
	wasmProposalHandler govv1.Handler

	codeID       uint64
	contractAddr string
	encfg        params.EncodingConfig
}

func (s *IntegrationTestSuite) SetupTest(t *testing.T) {
	app := umeeapp.Setup(t)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{
		ChainID: fmt.Sprintf("test-chain-%s", tmrand.Str(4)),
		Height:  9,
		Time:    time.Date(2022, 4, 20, 10, 20, 15, 1, time.UTC),
	})

	// mint and send coins to addrs
	assert.NilError(t, app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, initCoins.MulInt(sdk.NewInt(10))))
	assert.NilError(t, app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, addr, initCoins))
	assert.NilError(t, app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, initCoins))
	assert.NilError(t, app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, addr2, initCoins))
	s.T = t
	s.app = app
	s.ctx = ctx
	s.wasmMsgServer = wasmkeeper.NewMsgServerImpl(wasmkeeper.NewDefaultPermissionKeeper(app.WasmKeeper))
	querier := app.GRPCQueryRouter()
	wasmtypes.RegisterMsgServer(querier, s.wasmMsgServer)

	queryHelper := baseapp.NewQueryServerTestHelper(ctx, app.InterfaceRegistry())
	grpc := wasmkeeper.Querier(&app.WasmKeeper)
	wasmtypes.RegisterQueryServer(queryHelper, grpc)
	s.wasmQueryClient = wasmtypes.NewQueryClient(queryHelper)
	s.wasmProposalHandler = wasmkeeper.NewWasmProposalHandler(app.WasmKeeper, umeeapp.GetWasmEnabledProposals())
	s.encfg = umeeapp.MakeEncodingConfig()
}

// NewTestMsgCreateValidator test msg creator
func NewTestMsgCreateValidator(address sdk.ValAddress, pubKey cryptotypes.PubKey, amt math.Int) *stakingtypes.MsgCreateValidator {
	commission := stakingtypes.NewCommissionRates(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec())
	msg, _ := stakingtypes.NewMsgCreateValidator(
		address, pubKey, sdk.NewCoin(types.UmeeDenom, amt),
		stakingtypes.Description{}, commission, sdk.OneInt(),
	)

	return msg
}

func (s *IntegrationTestSuite) cw20StoreCode(sender sdk.AccAddress, cwArtifacePath string) (codeId uint64) {
	cw20Code, err := os.ReadFile(cwArtifacePath)
	assert.NilError(s.T, err)
	storeCodeProposal := wasmtypes.StoreCodeProposal{
		Title:                 cwArtifacePath,
		Description:           cwArtifacePath,
		RunAs:                 sender.String(),
		WASMByteCode:          cw20Code,
		InstantiatePermission: &wasmtypes.AllowEverybody,
	}

	err = s.wasmProposalHandler(s.ctx, &storeCodeProposal)
	assert.NilError(s.T, err)

	codes, err := s.wasmQueryClient.PinnedCodes(sdk.WrapSDKContext(s.ctx), &wasmtypes.QueryPinnedCodesRequest{})
	assert.NilError(s.T, err)
	assert.Equal(s.T, true, len(codes.CodeIDs) > 0)

	return codes.CodeIDs[len(codes.CodeIDs)-1]
}

func (s *IntegrationTestSuite) transfer(contracAddr string, amount uint64, from, to sdk.AccAddress) {
	s.contractAddr = contracAddr
	msg := cw20ExecMsg{Transfer: &transferMsg{
		Recipient: to.String(),
		Amount:    amount,
	}}
	transferBz, err := json.Marshal(msg)
	assert.NilError(s.T, err)

	s.execContract(from, transferBz)
}

func (s *IntegrationTestSuite) execContract(sender sdk.AccAddress, msg []byte) {
	_, err := s.wasmMsgServer.ExecuteContract(sdk.WrapSDKContext(s.ctx), &wasmtypes.MsgExecuteContract{
		Sender:   sender.String(),
		Contract: s.contractAddr,
		Msg:      msg,
	})
	assert.NilError(s.T, err)
}

func (s *IntegrationTestSuite) queryContract(q []byte) *wasmtypes.QuerySmartContractStateResponse {
	resp, err := s.wasmQueryClient.SmartContractState(sdk.WrapSDKContext(s.ctx), &wasmtypes.QuerySmartContractStateRequest{
		Address: s.contractAddr, QueryData: q,
	})
	assert.NilError(s.T, err)
	return resp
}

func (s *IntegrationTestSuite) queryBalance(contracAddr string, address sdk.AccAddress) uint64 {
	queryBobBalance := cw20QueryBalance{
		Balance: struct{ Address }{
			Address: Address{
				Address: address.String(),
			},
		},
	}
	queryBobBalanceBz, err := json.Marshal(queryBobBalance)
	assert.NilError(s.T, err)

	queryBobBalanceResp, err := s.wasmQueryClient.SmartContractState(sdk.WrapSDKContext(s.ctx), &wasmtypes.QuerySmartContractStateRequest{Address: contracAddr, QueryData: queryBobBalanceBz})
	assert.NilError(s.T, err)

	var bobBalanceResp cw20QueryBalanceResp
	err = json.Unmarshal(queryBobBalanceResp.Data, &bobBalanceResp)
	assert.NilError(s.T, err)

	bobBalanceUint, err := strconv.ParseUint(bobBalanceResp.Balance, 10, 64)
	assert.NilError(s.T, err)
	return bobBalanceUint
}

func intMsgCw20(addr2Amount uint64) cw20InitMsg {
	return cw20InitMsg{
		Name:     "Cw20TestToken",
		Symbol:   "CashSymbol",
		Decimals: 4,
		InitialBalances: []Balance{
			{
				Address: Address{
					Address: addr.String(),
				},
				Amount: 1003,
			},
			{
				Address: Address{
					Address: addr2.String(),
				},
				Amount: addr2Amount,
			},
		},
	}
}

func (s *IntegrationTestSuite) cw20InitiateCode(sender sdk.AccAddress, init interface{}) (*wasmtypes.MsgInstantiateContractResponse, error) {
	initBz, err := json.Marshal(init)
	assert.NilError(s.T, err)

	initMsg := wasmtypes.MsgInstantiateContract{
		Sender: sender.String(),
		CodeID: s.codeID,
		Label:  cw20Label,
		Funds:  sdk.Coins{sdk.NewCoin(appparams.BondDenom, sdk.NewIntFromUint64(10))},
		Msg:    initBz,
		Admin:  sender.String(),
	}
	return s.wasmMsgServer.InstantiateContract(sdk.WrapSDKContext(s.ctx), &initMsg)
}

func (s *IntegrationTestSuite) TestCw20Store() {
	codeID := s.cw20StoreCode(addr, cw20Artifact)
	s.codeID = codeID
	assert.Equal(s.T, uint64(1), codeID)
}

func (s *IntegrationTestSuite) TestCw20Instantiate() {
	intantiateResp, err := s.cw20InitiateCode(addr, intMsgCw20(200000))
	assert.NilError(s.T, err)
	assert.Equal(s.T, "umee14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9scsdqqx", intantiateResp.Address)
	s.contractAddr = intantiateResp.Address
}

func (s *IntegrationTestSuite) TestCw20ContractInfo() {
	sender := addr
	intantiateResp, err := s.cw20InitiateCode(sender, intMsgCw20(200))
	assert.NilError(s.T, err)

	cw20ContractInfo, err := s.wasmQueryClient.ContractInfo(sdk.WrapSDKContext(s.ctx), &wasmtypes.QueryContractInfoRequest{Address: intantiateResp.Address})
	assert.NilError(s.T, err)
	assert.Equal(s.T, uint64(1), cw20ContractInfo.CodeID)
	assert.Equal(s.T, sender.String(), cw20ContractInfo.Admin)
	assert.Equal(s.T, cw20Label, cw20ContractInfo.Label)
	assert.Equal(s.T, "umee1suhgf5svhu4usrurvxzlgn54ksxmn8gljarjtxqnapv8kjnp4nrs89p4g0", cw20ContractInfo.Address)
}

func (s *IntegrationTestSuite) TestCw20CheckBalance() {
	sender, bobAddr, bobAmount := addr, addr2, uint64(2500)

	intantiateResp, err := s.cw20InitiateCode(sender, intMsgCw20(bobAmount))
	assert.NilError(s.T, err)

	bobBalanceUint := s.queryBalance(intantiateResp.Address, bobAddr)
	assert.Equal(s.T, bobAmount, bobBalanceUint)
}

func (s *IntegrationTestSuite) TestCw20Transfer() {
	sender, bobAddr, bobAmount := addr, addr2, uint64(2500)

	intantiateResp, err := s.cw20InitiateCode(sender, intMsgCw20(bobAmount))
	assert.NilError(s.T, err)

	contracAddr := intantiateResp.Address
	amountToTransfer := uint64(100)

	s.transfer(contracAddr, amountToTransfer, addr, bobAddr)

	bobBalanceUint := s.queryBalance(contracAddr, bobAddr)
	assert.Equal(s.T, bobAmount+amountToTransfer, bobBalanceUint)
}

func (s *IntegrationTestSuite) InitiateUmeeCosmwasm() {
	// umee cosmwasm contract upload
	codeID := s.cw20StoreCode(addr2, umeeArtifact)
	assert.Equal(s.T, uint64(2), codeID)
	s.codeID = codeID

	// initiate the umee cw contract
	intantiateResp, err := s.cw20InitiateCode(addr2, new(InstantiateMsg))
	assert.NilError(s.T, err)
	s.contractAddr = intantiateResp.Address
}

func (s *IntegrationTestSuite) genCustomQuery(umeeQuery wq.UmeeQuery) []byte {
	cq, err := json.Marshal(UmeeCwCustomQuery(umeeQuery))
	assert.NilError(s.T, err)
	return cq
}

func (s *IntegrationTestSuite) genCustomTx(msg wm.UmeeMsg) []byte {
	cq, err := json.Marshal(UmeeCwCustomTx(msg))
	assert.NilError(s.T, err)
	return cq
}
