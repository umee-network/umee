package simulation_test

import (
	"fmt"
	"math/rand"
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"

	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/stretchr/testify/suite"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	umeeapp "github.com/umee-network/umee/app"
	umeeappbeta "github.com/umee-network/umee/app/beta"
	"github.com/umee-network/umee/x/leverage/simulation"
	"github.com/umee-network/umee/x/leverage/types"
)

// SimTestSuite wraps the test suite for running the simulations
type SimTestSuite struct {
	suite.Suite

	app *umeeappbeta.UmeeApp
	ctx sdk.Context
}

// SetupTest creates a new umee base app
func (s *SimTestSuite) SetupTest() {
	s.app = umeeappbeta.Setup(s.T(), false, 1)
	s.ctx = s.app.BaseApp.NewContext(false, tmproto.Header{
		ChainID: fmt.Sprintf("test-chain-%s", tmrand.Str(4)),
		Height:  1,
	})
}

// TestWeightedOperations tests the weights of the operations.
func (s *SimTestSuite) TestWeightedOperations() {
	cdc := s.app.AppCodec()
	appParams := make(simtypes.AppParams)

	weightesOps := simulation.WeightedOperations(appParams, cdc, s.app.AccountKeeper, s.app.BankKeeper)

	// setup 3 accounts
	r := rand.New(rand.NewSource(1))
	accs := s.getTestingAccounts(r, 3)

	expected := []struct {
		weight     int
		opMsgRoute string
		opMsgName  string
	}{
		{simulation.DefaultWeightMsgLendAsset, types.ModuleName, types.EventTypeLoanAsset},
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

func TestSimTestSuite(t *testing.T) {
	suite.Run(t, new(SimTestSuite))
}
