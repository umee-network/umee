package simulation_test

import (
	"math/rand"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/gogo/protobuf/proto"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	umeeapp "github.com/umee-network/umee/v3/app"
	"github.com/umee-network/umee/v3/x/leverage"
	"github.com/umee-network/umee/v3/x/leverage/simulation"
	"github.com/umee-network/umee/v3/x/leverage/types"
)

// SimTestSuite wraps the test suite for running the simulations
type SimTestSuite struct {
	suite.Suite

	app *umeeapp.UmeeApp
	ctx sdk.Context
}

// SetupTest creates a new umee base app
func (s *SimTestSuite) SetupTest() {
	checkTx := false
	app := umeeapp.Setup(s.T(), checkTx, 1)
	ctx := app.NewContext(checkTx, tmproto.Header{})

	umeeToken := types.Token{
		BaseDenom:              umeeapp.BondDenom,
		ReserveFactor:          sdk.MustNewDecFromStr("0.25"),
		CollateralWeight:       sdk.MustNewDecFromStr("0.5"),
		LiquidationThreshold:   sdk.MustNewDecFromStr("0.5"),
		BaseBorrowRate:         sdk.MustNewDecFromStr("0.02"),
		KinkBorrowRate:         sdk.MustNewDecFromStr("0.2"),
		MaxBorrowRate:          sdk.MustNewDecFromStr("1.0"),
		KinkUtilization:        sdk.MustNewDecFromStr("0.8"),
		LiquidationIncentive:   sdk.MustNewDecFromStr("0.1"),
		SymbolDenom:            umeeapp.DisplayDenom,
		Exponent:               6,
		EnableMsgSupply:        true,
		EnableMsgBorrow:        true,
		Blacklist:              false,
		MaxCollateralShare:     sdk.MustNewDecFromStr("1"),
		MaxSupplyUtilization:   sdk.MustNewDecFromStr("0.9"),
		MinCollateralLiquidity: sdk.MustNewDecFromStr("0"),
		MaxSupply:              sdk.NewInt(100000000000),
	}
	atomIBCToken := types.Token{
		BaseDenom:              "ibc/CDC4587874B85BEA4FCEC3CEA5A1195139799A1FEE711A07D972537E18FDA39D",
		ReserveFactor:          sdk.MustNewDecFromStr("0.25"),
		CollateralWeight:       sdk.MustNewDecFromStr("0.8"),
		LiquidationThreshold:   sdk.MustNewDecFromStr("0.8"),
		BaseBorrowRate:         sdk.MustNewDecFromStr("0.05"),
		KinkBorrowRate:         sdk.MustNewDecFromStr("0.3"),
		MaxBorrowRate:          sdk.MustNewDecFromStr("0.9"),
		KinkUtilization:        sdk.MustNewDecFromStr("0.75"),
		LiquidationIncentive:   sdk.MustNewDecFromStr("0.11"),
		SymbolDenom:            "ATOM",
		Exponent:               6,
		EnableMsgSupply:        true,
		EnableMsgBorrow:        true,
		Blacklist:              false,
		MaxCollateralShare:     sdk.MustNewDecFromStr("1"),
		MaxSupplyUtilization:   sdk.MustNewDecFromStr("0.9"),
		MinCollateralLiquidity: sdk.MustNewDecFromStr("0"),
		MaxSupply:              sdk.NewInt(100000000000),
	}
	uabc := types.Token{
		BaseDenom:              "uabc",
		ReserveFactor:          sdk.MustNewDecFromStr("0"),
		CollateralWeight:       sdk.MustNewDecFromStr("0.1"),
		LiquidationThreshold:   sdk.MustNewDecFromStr("0.1"),
		BaseBorrowRate:         sdk.MustNewDecFromStr("0.02"),
		KinkBorrowRate:         sdk.MustNewDecFromStr("0.22"),
		MaxBorrowRate:          sdk.MustNewDecFromStr("1.52"),
		KinkUtilization:        sdk.MustNewDecFromStr("0.87"),
		LiquidationIncentive:   sdk.MustNewDecFromStr("0.1"),
		SymbolDenom:            "ABC",
		Exponent:               6,
		EnableMsgSupply:        true,
		EnableMsgBorrow:        true,
		Blacklist:              false,
		MaxCollateralShare:     sdk.MustNewDecFromStr("1"),
		MaxSupplyUtilization:   sdk.MustNewDecFromStr("0.9"),
		MinCollateralLiquidity: sdk.MustNewDecFromStr("0"),
		MaxSupply:              sdk.NewInt(100000000000),
	}

	tokens := []types.Token{umeeToken, atomIBCToken, uabc}
	leverage.InitGenesis(ctx, app.LeverageKeeper, *types.DefaultGenesis())
	for _, token := range tokens {
		s.Require().NoError(app.LeverageKeeper.SetTokenSettings(ctx, token))
		app.OracleKeeper.SetExchangeRate(ctx, token.SymbolDenom, sdk.MustNewDecFromStr("100.0"))
	}

	s.app = app
	s.ctx = ctx
}

// msg must be a pointer to a Message
func (s *SimTestSuite) unmarshal(op *simtypes.OperationMsg, msg proto.Message) {
	// TODO: use app Codec instead of Amino #1313
	// err := s.app.AppCodec().UnmarshalJSON(op.Msg, msg)

	s.Require().True(op.OK)
	err := types.ModuleCdc.UnmarshalJSON(op.Msg, msg)
	s.Require().NoError(err)
}

// getTestingAccounts generates accounts with balance in all registered token types
func (s *SimTestSuite) getTestingAccounts(r *rand.Rand, n int, cb func(fundedAccount simtypes.Account)) []simtypes.Account {
	accounts := simtypes.RandomAccounts(r, n)

	initAmt := sdk.NewInt(200000000) // 200 * 10^6
	accCoins := sdk.NewCoins()

	tokens := s.app.LeverageKeeper.GetAllRegisteredTokens(s.ctx)

	for _, token := range tokens {
		accCoins = accCoins.Add(sdk.NewCoin(token.BaseDenom, initAmt))
	}

	// add coins to the accounts
	for _, account := range accounts {
		acc := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, account.Address)
		s.app.AccountKeeper.SetAccount(s.ctx, acc)
		s.Require().NoError(s.app.BankKeeper.MintCoins(s.ctx, minttypes.ModuleName, accCoins))
		s.Require().NoError(
			s.app.BankKeeper.SendCoinsFromModuleToAccount(s.ctx, minttypes.ModuleName, acc.GetAddress(), accCoins),
		)
		cb(account)
	}

	return accounts
}

// TestWeightedOperations tests the weights of the operations.
func (s *SimTestSuite) TestWeightedOperations() {
	cdc := s.app.AppCodec()
	appParams := make(simtypes.AppParams)

	weightesOps := simulation.WeightedOperations(appParams, cdc, s.app.AccountKeeper, s.app.BankKeeper, s.app.LeverageKeeper)

	// setup 3 accounts
	r := rand.New(rand.NewSource(1))
	accs := s.getTestingAccounts(r, 3, func(acc simtypes.Account) {})
	expected := []struct {
		weight    int
		opMsgName string
	}{
		{simulation.DefaultWeightMsgSupply, sdk.MsgTypeURL(new(types.MsgSupply))},
		{simulation.DefaultWeightMsgWithdraw, sdk.MsgTypeURL(new(types.MsgWithdraw))},
		{simulation.DefaultWeightMsgBorrow, sdk.MsgTypeURL(new(types.MsgBorrow))},
		{simulation.DefaultWeightMsgCollateralize, sdk.MsgTypeURL(new(types.MsgCollateralize))},
		{simulation.DefaultWeightMsgDecollateralize, sdk.MsgTypeURL(new(types.MsgDecollateralize))},
		{simulation.DefaultWeightMsgRepay, sdk.MsgTypeURL(new(types.MsgRepay))},
		{simulation.DefaultWeightMsgLiquidate, sdk.MsgTypeURL(new(types.MsgLiquidate))},
	}

	for i, w := range weightesOps {
		operationMsg, _, err := w.Op()(r, s.app.BaseApp, s.ctx, accs, "")
		s.Require().NoError(err)
		// the following checks are very much dependent from the ordering of the output given
		// by WeightedOperations. if the ordering in WeightedOperations changes some tests
		// will fail
		s.Require().Equal(expected[i].weight, w.Weight(), "weight should be the same")
		s.Require().Equal(expected[i].opMsgName, operationMsg.Name, "operation Msg name should be the same")
	}
}

func (s *SimTestSuite) TestSimulateMsgSupply() {
	r := rand.New(rand.NewSource(1))
	accs := s.getTestingAccounts(r, 3, func(fundedAccount simtypes.Account) {})

	s.app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: s.app.LastBlockHeight() + 1, AppHash: s.app.LastCommitID().Hash}})

	op := simulation.SimulateMsgSupply(s.app.AccountKeeper, s.app.BankKeeper)
	operationMsg, futureOperations, err := op(r, s.app.BaseApp, s.ctx, accs, "")
	s.Require().NoError(err)

	var msg types.MsgSupply
	s.unmarshal(&operationMsg, &msg)
	s.Require().Equal("umee1ghekyjucln7y67ntx7cf27m9dpuxxemn8w6h33", msg.Supplier)
	s.Require().Equal("185121068uumee", msg.Asset.String())
	s.Require().Len(futureOperations, 0)
}

func (s *SimTestSuite) TestSimulateMsgWithdraw() {
	r := rand.New(rand.NewSource(1))
	supplyToken := sdk.NewCoin(umeeapp.BondDenom, sdk.NewInt(100))

	accs := s.getTestingAccounts(r, 3, func(fundedAccount simtypes.Account) {
		_, err := s.app.LeverageKeeper.Supply(s.ctx, fundedAccount.Address, supplyToken)
		s.Require().NoError(err)
	})

	s.app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: s.app.LastBlockHeight() + 1, AppHash: s.app.LastCommitID().Hash}})

	op := simulation.SimulateMsgWithdraw(s.app.AccountKeeper, s.app.BankKeeper, s.app.LeverageKeeper)
	operationMsg, futureOperations, err := op(r, s.app.BaseApp, s.ctx, accs, "")
	s.Require().NoError(err)

	var msg types.MsgWithdraw
	s.unmarshal(&operationMsg, &msg)
	s.Require().True(operationMsg.OK)
	s.Require().Equal("umee1ghekyjucln7y67ntx7cf27m9dpuxxemn8w6h33", msg.Supplier)
	s.Require().Equal("73u/uumee", msg.Asset.String())
	s.Require().Len(futureOperations, 0)
}

func (s *SimTestSuite) TestSimulateMsgBorrow() {
	r := rand.New(rand.NewSource(8))
	supplyToken := sdk.NewCoin(umeeapp.BondDenom, sdk.NewInt(1000))

	accs := s.getTestingAccounts(r, 3, func(fundedAccount simtypes.Account) {
		uToken, err := s.app.LeverageKeeper.Supply(s.ctx, fundedAccount.Address, supplyToken)
		if err != nil {
			s.Require().NoError(err)
		}
		s.Require().NoError(s.app.LeverageKeeper.Collateralize(s.ctx, fundedAccount.Address, uToken))
	})

	s.app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: s.app.LastBlockHeight() + 1, AppHash: s.app.LastCommitID().Hash}})

	op := simulation.SimulateMsgBorrow(s.app.AccountKeeper, s.app.BankKeeper, s.app.LeverageKeeper)
	operationMsg, futureOperations, err := op(r, s.app.BaseApp, s.ctx, accs, "")
	s.Require().NoError(err)

	var msg types.MsgBorrow
	s.unmarshal(&operationMsg, &msg)
	s.Require().Equal("umee1qnclgkcxtuledc8xhle4lqly2q0z96uqkks60s", msg.Borrower)
	s.Require().Equal("67uumee", msg.Asset.String())
	s.Require().Len(futureOperations, 0)
}

func (s *SimTestSuite) TestSimulateMsgCollateralize() {
	r := rand.New(rand.NewSource(1))
	supplyToken := sdk.NewInt64Coin(umeeapp.BondDenom, 100)

	accs := s.getTestingAccounts(r, 3, func(fundedAccount simtypes.Account) {
		_, err := s.app.LeverageKeeper.Supply(s.ctx, fundedAccount.Address, supplyToken)
		s.Require().NoError(err)
	})

	s.app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: s.app.LastBlockHeight() + 1, AppHash: s.app.LastCommitID().Hash}})

	op := simulation.SimulateMsgCollateralize(s.app.AccountKeeper, s.app.BankKeeper, s.app.LeverageKeeper)
	operationMsg, futureOperations, err := op(r, s.app.BaseApp, s.ctx, accs, "")
	s.Require().NoError(err)

	var msg types.MsgCollateralize
	s.unmarshal(&operationMsg, &msg)
	s.Require().Equal("umee1ghekyjucln7y67ntx7cf27m9dpuxxemn8w6h33", msg.Borrower)
	s.Require().Equal("73u/uumee", msg.Asset.String())
	s.Require().Len(futureOperations, 0)
}

func (s *SimTestSuite) TestSimulateMsgDecollateralize() {
	r := rand.New(rand.NewSource(1))
	supplyToken := sdk.NewInt64Coin(umeeapp.BondDenom, 100)

	accs := s.getTestingAccounts(r, 3, func(fundedAccount simtypes.Account) {
		uToken, err := s.app.LeverageKeeper.Supply(s.ctx, fundedAccount.Address, supplyToken)
		s.Require().NoError(err)
		s.Require().NoError(s.app.LeverageKeeper.Collateralize(s.ctx, fundedAccount.Address, uToken))
	})

	s.app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: s.app.LastBlockHeight() + 1, AppHash: s.app.LastCommitID().Hash}})

	op := simulation.SimulateMsgDecollateralize(s.app.AccountKeeper, s.app.BankKeeper, s.app.LeverageKeeper)
	operationMsg, futureOperations, err := op(r, s.app.BaseApp, s.ctx, accs, "")
	s.Require().NoError(err)

	var msg types.MsgDecollateralize
	s.unmarshal(&operationMsg, &msg)
	s.Require().Equal("umee1ghekyjucln7y67ntx7cf27m9dpuxxemn8w6h33", msg.Borrower)
	s.Require().Equal("73u/uumee", msg.Asset.String())
	s.Require().Len(futureOperations, 0)
}

func (s *SimTestSuite) TestSimulateMsgRepay() {
	r := rand.New(rand.NewSource(1))
	supplyToken := sdk.NewInt64Coin(umeeapp.BondDenom, 100)
	borrowToken := sdk.NewInt64Coin(umeeapp.BondDenom, 20)

	accs := s.getTestingAccounts(r, 3, func(fundedAccount simtypes.Account) {
		uToken, err := s.app.LeverageKeeper.Supply(s.ctx, fundedAccount.Address, supplyToken)
		s.Require().NoError(err)
		s.Require().NoError(s.app.LeverageKeeper.Collateralize(s.ctx, fundedAccount.Address, uToken))
		s.Require().NoError(s.app.LeverageKeeper.Borrow(s.ctx, fundedAccount.Address, borrowToken))
	})

	s.app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: s.app.LastBlockHeight() + 1, AppHash: s.app.LastCommitID().Hash}})

	op := simulation.SimulateMsgRepay(s.app.AccountKeeper, s.app.BankKeeper, s.app.LeverageKeeper)
	operationMsg, futureOperations, err := op(r, s.app.BaseApp, s.ctx, accs, "")
	s.Require().NoError(err)

	var msg types.MsgRepay
	s.unmarshal(&operationMsg, &msg)
	s.Require().Equal("umee1ghekyjucln7y67ntx7cf27m9dpuxxemn8w6h33", msg.Borrower)
	s.Require().Equal("9uumee", msg.Asset.String())
	s.Require().Len(futureOperations, 0)
}

func (s *SimTestSuite) TestSimulateMsgLiquidate() {
	r := rand.New(rand.NewSource(1))
	supplyToken := sdk.NewCoin(umeeapp.BondDenom, sdk.NewInt(100))
	uToken := sdk.NewCoin("u/"+umeeapp.BondDenom, sdk.NewInt(100))
	borrowToken := sdk.NewCoin(umeeapp.BondDenom, sdk.NewInt(10))

	accs := s.getTestingAccounts(r, 3, func(fundedAccount simtypes.Account) {
		_, err := s.app.LeverageKeeper.Supply(s.ctx, fundedAccount.Address, supplyToken)
		s.Require().NoError(err)
		s.Require().NoError(s.app.LeverageKeeper.Collateralize(s.ctx, fundedAccount.Address, uToken))
		s.Require().NoError(s.app.LeverageKeeper.Borrow(s.ctx, fundedAccount.Address, borrowToken))
	})

	s.app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: s.app.LastBlockHeight() + 1, AppHash: s.app.LastCommitID().Hash}})

	op := simulation.SimulateMsgLiquidate(s.app.AccountKeeper, s.app.BankKeeper, s.app.LeverageKeeper)
	operationMsg, futureOperations, err := op(r, s.app.BaseApp, s.ctx, accs, "")
	s.Require().EqualError(err,
		"failed to execute message; message index: 0: borrower not eligible for liquidation",
	)

	// While it is no longer simple to create an eligible liquidation target using exported keeper methods here,
	// we can still verify some properties of the resulting operation.
	s.Require().Empty(operationMsg.Msg)
	s.Require().False(operationMsg.OK)
	s.Require().Len(futureOperations, 0)
}

func TestSimTestSuite(t *testing.T) {
	suite.Run(t, new(SimTestSuite))
}
