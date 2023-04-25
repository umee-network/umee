package wasm_test

import (
	"encoding/json"
	"fmt"
	"os"
	"strconv"
	"testing"
	"time"

	"gotest.tools/v3/assert"

	"cosmossdk.io/math"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	umeeapp "github.com/umee-network/umee/v4/app"
	appparams "github.com/umee-network/umee/v4/app/params"
	"github.com/umee-network/umee/v4/x/oracle/types"
)

const (
	initialPower = int64(10000000000)
	cw20Artifact = "../../tests/artifacts/cw20_base.wasm"
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

type IntegrationTestSuite struct {
	T   *testing.T
	ctx sdk.Context
	app *umeeapp.UmeeApp

	wasmMsgServer       wasmtypes.MsgServer
	wasmQueryClient     wasmtypes.QueryClient
	wasmProposalHandler govv1.Handler
}

func (s *IntegrationTestSuite) SetupTest(t *testing.T) {
	app := umeeapp.Setup(t)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{
		ChainID: fmt.Sprintf("test-chain-%s", tmrand.Str(4)),
		Height:  9,
		Time:    time.Date(2022, 4, 20, 10, 20, 15, 1, time.UTC),
	})

	// mint and send coins to addrs
	assert.NilError(t, app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, initCoins))
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

func (s *IntegrationTestSuite) cw20StoreCode(sender sdk.AccAddress) (codeId uint64) {
	cw20Code, err := os.ReadFile(cw20Artifact)
	assert.NilError(s.T, err)
	storeCodeProposal := wasmtypes.StoreCodeProposal{
		Title:                 "Store cw20",
		Description:           "Store brand new contract",
		RunAs:                 sender.String(),
		WASMByteCode:          cw20Code,
		InstantiatePermission: &wasmtypes.AllowEverybody,
	}

	s.wasmProposalHandler(s.ctx, &storeCodeProposal)

	codes, err := s.wasmQueryClient.PinnedCodes(sdk.WrapSDKContext(s.ctx), &wasmtypes.QueryPinnedCodesRequest{})
	assert.NilError(s.T, err)
	assert.Equal(s.T, true, len(codes.CodeIDs) > 0)

	return codes.CodeIDs[len(codes.CodeIDs)-1]
}

func (s *IntegrationTestSuite) transfer(contracAddr string, amount uint64, from, to sdk.AccAddress) {
	transfer := cw20ExecMsg{Transfer: &transferMsg{
		Recipient: to.String(),
		Amount:    amount,
	}}
	transferBz, err := json.Marshal(transfer)
	assert.NilError(s.T, err)

	_, err = s.wasmMsgServer.ExecuteContract(sdk.WrapSDKContext(s.ctx), &wasmtypes.MsgExecuteContract{
		Sender:   from.String(),
		Contract: contracAddr,
		Msg:      transferBz,
	})
	assert.NilError(s.T, err)
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

func (s *IntegrationTestSuite) cw20InitiateCode(sender sdk.AccAddress, addr2Amount uint64) (*wasmtypes.MsgInstantiateContractResponse, error) {
	codeID := s.cw20StoreCode(addr)

	init := cw20InitMsg{
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

	initBz, err := json.Marshal(init)
	assert.NilError(s.T, err)

	initMsg := wasmtypes.MsgInstantiateContract{
		Sender: sender.String(),
		CodeID: codeID,
		Label:  cw20Label,
		Funds:  sdk.Coins{sdk.NewCoin(appparams.BondDenom, sdk.NewIntFromUint64(10))},
		Msg:    initBz,
		Admin:  sender.String(),
	}
	return s.wasmMsgServer.InstantiateContract(sdk.WrapSDKContext(s.ctx), &initMsg)
}

func (s *IntegrationTestSuite) TestCw20Store() {
	codeID := s.cw20StoreCode(addr)
	assert.Equal(s.T, uint64(1), codeID)
}

func (s *IntegrationTestSuite) TestCw20Instantiate() {
	msgIntantiateResponse, err := s.cw20InitiateCode(addr, 200)
	assert.NilError(s.T, err)
	assert.Equal(s.T, "umee1wug8sewp6cedgkmrmvhl3lf3tulagm9hnvy8p0rppz9yjw0g4wtqqpllkv", msgIntantiateResponse.Address)
}

func (s *IntegrationTestSuite) TestCw20ContractInfo() {
	sender := addr
	msgIntantiateResponse, err := s.cw20InitiateCode(sender, 200)
	assert.NilError(s.T, err)

	cw20ContractInfo, err := s.wasmQueryClient.ContractInfo(sdk.WrapSDKContext(s.ctx), &wasmtypes.QueryContractInfoRequest{Address: msgIntantiateResponse.Address})
	assert.NilError(s.T, err)
	assert.Equal(s.T, uint64(3), cw20ContractInfo.CodeID)
	assert.Equal(s.T, sender.String(), cw20ContractInfo.Admin)
	assert.Equal(s.T, cw20Label, cw20ContractInfo.Label)
	assert.Equal(s.T, "umee1qg5ega6dykkxc307y25pecuufrjkxkaggkkxh7nad0vhyhtuhw3shmlsys", cw20ContractInfo.Address)
}

func (s *IntegrationTestSuite) TestCw20CheckBalance() {
	sender, bobAddr, bobAmount := addr, addr2, uint64(2500)

	msgIntantiateResponse, err := s.cw20InitiateCode(sender, bobAmount)
	assert.NilError(s.T, err)

	bobBalanceUint := s.queryBalance(msgIntantiateResponse.Address, bobAddr)
	assert.Equal(s.T, bobAmount, bobBalanceUint)
}

func (s *IntegrationTestSuite) TestCw20Transfer() {
	sender, bobAddr, bobAmount := addr, addr2, uint64(2500)

	msgIntantiateResponse, err := s.cw20InitiateCode(sender, bobAmount)
	assert.NilError(s.T, err)

	contracAddr := msgIntantiateResponse.Address
	amountToTransfer := uint64(100)

	s.transfer(contracAddr, amountToTransfer, addr, bobAddr)

	bobBalanceUint := s.queryBalance(contracAddr, bobAddr)
	assert.Equal(s.T, bobAmount+amountToTransfer, bobBalanceUint)
}

func TestCosmwasmCW20(t *testing.T) {
	its := new(IntegrationTestSuite)
	// setup the test configuration
	its.SetupTest(t)

	its.TestCw20Store()
	its.TestCw20Instantiate()
	its.TestCw20ContractInfo()
	its.TestCw20CheckBalance()
	its.TestCw20Transfer()
}
