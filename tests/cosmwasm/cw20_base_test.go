package cosmwasm_test

import (
	"encoding/json"
	"fmt"
	"io/ioutil"
	"strconv"
	"testing"
	"time"

	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	"github.com/cosmos/cosmos-sdk/baseapp"
	cryptotypes "github.com/cosmos/cosmos-sdk/crypto/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"

	tmrand "github.com/tendermint/tendermint/libs/rand"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/stretchr/testify/suite"
	"github.com/tendermint/tendermint/crypto/secp256k1"

	umeeapp "github.com/umee-network/umee/v2/app"
	"github.com/umee-network/umee/v2/x/oracle/types"
)

const (
	initialPower = int64(10000000000)
	cw20Artifact = "../artifacts/cw20_base.wasm"
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
	initCoins  = sdk.NewCoins(sdk.NewCoin(umeeapp.BondDenom, initTokens))
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
	suite.Suite

	ctx sdk.Context
	app *umeeapp.UmeeApp

	wasmMsgServer       wasmtypes.MsgServer
	wasmQueryClient     wasmtypes.QueryClient
	wasmProposalHandler govtypes.Handler
}

func (s *IntegrationTestSuite) SetupTest() {
	app := umeeapp.Setup(s.T(), false, 1)
	ctx := app.BaseApp.NewContext(false, tmproto.Header{
		ChainID: fmt.Sprintf("test-chain-%s", tmrand.Str(4)),
		Height:  9,
		Time:    time.Date(2022, 4, 20, 10, 20, 15, 1, time.UTC),
	})

	// mint and send coins to addrs
	s.Require().NoError(app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, initCoins))
	s.Require().NoError(app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, addr, initCoins))
	s.Require().NoError(app.BankKeeper.MintCoins(ctx, minttypes.ModuleName, initCoins))
	s.Require().NoError(app.BankKeeper.SendCoinsFromModuleToAccount(ctx, minttypes.ModuleName, addr2, initCoins))

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
func NewTestMsgCreateValidator(address sdk.ValAddress, pubKey cryptotypes.PubKey, amt sdk.Int) *stakingtypes.MsgCreateValidator {
	commission := stakingtypes.NewCommissionRates(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec())
	msg, _ := stakingtypes.NewMsgCreateValidator(
		address, pubKey, sdk.NewCoin(types.UmeeDenom, amt),
		stakingtypes.Description{}, commission, sdk.OneInt(),
	)

	return msg
}

func (s *IntegrationTestSuite) cw20StoreCode(sender sdk.AccAddress) (codeId uint64) {
	cw20Code, err := ioutil.ReadFile(cw20Artifact)
	s.Require().NoError(err)
	storeCodeProposal := wasmtypes.StoreCodeProposal{
		Title:                 "Store cw20",
		Description:           "Store brand new contract",
		RunAs:                 sender.String(),
		WASMByteCode:          cw20Code,
		InstantiatePermission: &wasmtypes.AllowEverybody,
	}

	s.wasmProposalHandler(s.ctx, &storeCodeProposal)

	codes, err := s.wasmQueryClient.PinnedCodes(sdk.WrapSDKContext(s.ctx), &wasmtypes.QueryPinnedCodesRequest{})
	s.Require().NoError(err)
	s.Require().True(len(codes.CodeIDs) > 0)

	return codes.CodeIDs[len(codes.CodeIDs)-1]
}

func (s *IntegrationTestSuite) transfer(contracAddr string, amount uint64, from, to sdk.AccAddress) {
	transfer := cw20ExecMsg{Transfer: &transferMsg{
		Recipient: to.String(),
		Amount:    amount,
	}}
	transferBz, err := json.Marshal(transfer)
	s.Require().NoError(err)

	execContractResp, err := s.wasmMsgServer.ExecuteContract(sdk.WrapSDKContext(s.ctx), &wasmtypes.MsgExecuteContract{
		Sender:   from.String(),
		Contract: contracAddr,
		Msg:      transferBz,
	})
	s.Require().NoError(err)
	s.Require().NotNil(execContractResp)
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
	s.Require().NoError(err)

	queryBobBalanceResp, err := s.wasmQueryClient.SmartContractState(sdk.WrapSDKContext(s.ctx), &wasmtypes.QuerySmartContractStateRequest{Address: contracAddr, QueryData: queryBobBalanceBz})
	s.Require().NoError(err)

	var bobBalanceResp cw20QueryBalanceResp
	err = json.Unmarshal(queryBobBalanceResp.Data, &bobBalanceResp)
	s.Require().NoError(err)

	bobBalanceUint, err := strconv.ParseUint(bobBalanceResp.Balance, 10, 64)
	s.Require().NoError(err)
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
	s.Require().NoError(err)

	initMsg := wasmtypes.MsgInstantiateContract{
		Sender: sender.String(),
		CodeID: codeID,
		Label:  cw20Label,
		Funds:  sdk.Coins{sdk.NewCoin(umeeapp.BondDenom, sdk.NewIntFromUint64(10))},
		Msg:    initBz,
		Admin:  sender.String(),
	}
	return s.wasmMsgServer.InstantiateContract(sdk.WrapSDKContext(s.ctx), &initMsg)
}

func (s *IntegrationTestSuite) TestCw20Store() {
	codeID := s.cw20StoreCode(addr)
	s.Require().Equal(uint64(1), codeID)
}

func (s *IntegrationTestSuite) TestCw20ErrorStoreWithoutGovProposal() {
	cw20Code, err := ioutil.ReadFile(cw20Artifact)
	s.Require().NoError(err)

	msgStoreCodeResp, err := s.wasmMsgServer.StoreCode(sdk.WrapSDKContext(s.ctx), &wasmtypes.MsgStoreCode{
		Sender:       addr.String(),
		WASMByteCode: cw20Code,
	})
	s.Require().ErrorContains(err, "can not create code: unauthorized")
	s.Require().Nil(msgStoreCodeResp)
}

func (s *IntegrationTestSuite) TestCw20Instantiate() {
	msgIntantiateResponse, err := s.cw20InitiateCode(addr, 200)
	s.Require().NoError(err)
	s.Require().Equal("umee14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9scsdqqx", msgIntantiateResponse.Address)
}

func (s *IntegrationTestSuite) TestCw20ContractInfo() {
	sender := addr
	msgIntantiateResponse, err := s.cw20InitiateCode(sender, 200)
	s.Require().NoError(err)

	cw20ContractInfo, err := s.wasmQueryClient.ContractInfo(sdk.WrapSDKContext(s.ctx), &wasmtypes.QueryContractInfoRequest{Address: msgIntantiateResponse.Address})
	s.Require().NoError(err)
	s.Require().Equal(uint64(1), cw20ContractInfo.CodeID)
	s.Require().Equal(sender.String(), cw20ContractInfo.Admin)
	s.Require().Equal(cw20Label, cw20ContractInfo.Label)
	s.Require().Equal("umee14hj2tavq8fpesdwxxcu44rty3hh90vhujrvcmstl4zr3txmfvw9scsdqqx", cw20ContractInfo.Address)
}

func (s *IntegrationTestSuite) TestCw20CheckBalance() {
	sender, bobAddr, bobAmount := addr, addr2, uint64(2500)

	msgIntantiateResponse, err := s.cw20InitiateCode(sender, bobAmount)
	s.Require().NoError(err)

	bobBalanceUint := s.queryBalance(msgIntantiateResponse.Address, bobAddr)
	s.Require().Equal(bobAmount, bobBalanceUint)
}

func (s *IntegrationTestSuite) TestCw20Transfer() {
	sender, bobAddr, bobAmount := addr, addr2, uint64(2500)

	msgIntantiateResponse, err := s.cw20InitiateCode(sender, bobAmount)
	s.Require().NoError(err)

	contracAddr := msgIntantiateResponse.Address
	amountToTransfer := uint64(100)

	s.transfer(contracAddr, amountToTransfer, addr, bobAddr)

	bobBalanceUint := s.queryBalance(contracAddr, bobAddr)
	s.Require().Equal(bobAmount+amountToTransfer, bobBalanceUint)
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(IntegrationTestSuite))
}
