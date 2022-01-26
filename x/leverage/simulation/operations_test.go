package simulation_test

import (
	"fmt"
	"math/rand"
	"testing"

	"github.com/cosmos/cosmos-sdk/simapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/suite"
	abci "github.com/tendermint/tendermint/abci/types"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	umeeapp "github.com/umee-network/umee/app"
	umeeappbeta "github.com/umee-network/umee/app/beta"
	"github.com/umee-network/umee/x/leverage/simulation"
	"github.com/umee-network/umee/x/leverage/types"
)

// SimTestSuite wraps the test suite for running the simulations
type SimTestSuite struct {
	suite.Suite

	app    *umeeappbeta.UmeeApp
	simApp *simapp.SimApp
	ctx    sdk.Context
}

// SetupTest creates a new umee base app
func (s *SimTestSuite) SetupTest() {
	checkTx := false
	s.simApp = simapp.Setup(checkTx)
	s.app = umeeappbeta.Setup(s.T(), checkTx, 1)
	s.ctx = s.app.BaseApp.NewContext(checkTx, tmproto.Header{})
}

// getTestingAccounts generates
func (s *SimTestSuite) getTestingAccounts(r *rand.Rand, n int) []simtypes.Account {
	accounts := simtypes.RandomAccounts(r, n)

	initAmt := sdk.TokensFromConsensusPower(200, sdk.DefaultPowerReduction)
	initCoins := sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, initAmt), sdk.NewCoin(umeeapp.BondDenom, initAmt))

	// add coins to the accounts
	for _, account := range accounts {
		acc := s.app.AccountKeeper.NewAccountWithAddress(s.ctx, account.Address)
		s.app.AccountKeeper.SetAccount(s.ctx, acc)
		s.Require().NoError(s.app.BankKeeper.MintCoins(s.ctx, minttypes.ModuleName, initCoins))
		s.Require().NoError(
			s.app.BankKeeper.SendCoinsFromModuleToAccount(s.ctx, minttypes.ModuleName, acc.GetAddress(), initCoins),
		)
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
	accs := s.getTestingAccounts(r, 3)

	expected := []struct {
		weight     int
		opMsgRoute string
		opMsgName  string
	}{
		{simulation.DefaultWeightMsgLendAsset, types.ModuleName, types.EventTypeLoanAsset},
		{simulation.DefaultWeightMsgWithdrawAsset, types.ModuleName, types.EventTypeWithdrawLoanedAsset},
		{simulation.DefaultWeightMsgBorrowAsset, types.ModuleName, types.EventTypeBorrowAsset},
		{simulation.DefaultWeightMsgSetCollateral, types.ModuleName, types.EventTypeSetCollateralSetting},
		{simulation.DefaultWeightMsgRepayAsset, types.ModuleName, types.EventTypeRepayBorrowedAsset},
		{simulation.DefaultWeightMsgLiquidate, types.ModuleName, types.EventTypeLiquidate},
	}

	for i, w := range weightesOps {
		operationMsg, _, _ := w.Op()(r, s.app.BaseApp, s.ctx, accs, "")
		// 	// the following checks are very much dependent from the ordering of the output given
		// 	// by WeightedOperations. if the ordering in WeightedOperations changes some tests
		// 	// will fail
		s.Require().Equal(expected[i].weight, w.Weight(), "weight should be the same")
		s.Require().Equal(expected[i].opMsgRoute, operationMsg.Route, "route should be the same")
		s.Require().Equal(expected[i].opMsgName, operationMsg.Name, "operation Msg name should be the same")
	}
}

// TestWeightedOperations tests the weights of the operations.
func (s *SimTestSuite) TestSimulateMsgLendAsset() {
	// cdc := s.app.AppCodec()
	// appParams := make(simtypes.AppParams)

	// weightesOps := simulation.WeightedOperations(appParams, cdc, s.app.AccountKeeper, s.app.BankKeeper)

	// setup 3 accounts
	r := rand.New(rand.NewSource(1))
	accs := s.getTestingAccounts(r, 3)

	s.app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{Height: s.app.LastBlockHeight() + 1, AppHash: s.app.LastCommitID().Hash}})

	op := simulation.SimulateMsgLendAsset(s.app.AccountKeeper, s.app.BankKeeper)
	// test failing for check signature ??
	operationMsg, futureOperations, err := op(r, s.app.BaseApp, s.ctx, accs, "") // s.ctx.ChainID()
	s.Require().NoError(err)

	fmt.Print(operationMsg, futureOperations)

	// var msg types.MsgLendAsset
	// types.ModuleCdc.UnmarshalJSON(operationMsg.Msg, &msg)

	// s.Require().True(operationMsg.OK)
	// s.Require().Equal("cosmos1ghekyjucln7y67ntx7cf27m9dpuxxemn4c8g4r", msg.Lender)
	// s.Require().Equal("cosmos1p8wcgrjr4pjju90xg6u9cgq55dxwq8j7u4x9a0", msg.Amount.String())
	// s.Require().Len(futureOperations, 0)
}

func TestSimTestSuite(t *testing.T) {
	suite.Run(t, new(SimTestSuite))
}
