package keeper_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/ibc-go/modules/apps/transfer/types"
	ibctesting "github.com/cosmos/ibc-go/testing"
	"github.com/stretchr/testify/suite"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"

	"github.com/umee-network/umee/app"
)

type KeeperTestSuite struct {
	suite.Suite

	coordinator *ibctesting.Coordinator
	chainA      *ibctesting.TestChain
	chainB      *ibctesting.TestChain

	queryClient types.QueryClient
}

func (s *KeeperTestSuite) SetupTest() {
	s.coordinator = ibctesting.NewCoordinator(s.T(), 0)
	ibctesting.DefaultTestingAppInit = setupTestingApp

	chains := make(map[string]*ibctesting.TestChain)
	for i := 0; i < 2; i++ {
		// create a chain with the temporary coordinator that we'll later override
		chainID := ibctesting.GetChainID(i)
		chain := ibctesting.NewTestChain(s.T(), ibctesting.NewCoordinator(s.T(), 0), chainID)

		balance := banktypes.Balance{
			Address: chain.SenderAccount.GetAddress().String(),
			Coins:   sdk.NewCoins(sdk.NewCoin(sdk.DefaultBondDenom, sdk.NewInt(100000000000000))),
		}

		// create application and override files in the IBC test chain
		app := ibctesting.SetupWithGenesisValSet(
			s.T(),
			chain.Vals,
			[]authtypes.GenesisAccount{
				chain.SenderAccount.(authtypes.GenesisAccount),
			},
			balance,
		)

		chain.App = app
		chain.QueryServer = app.GetIBCKeeper()
		chain.TxConfig = app.GetTxConfig()
		chain.Codec = app.AppCodec()
		chain.CurrentHeader = tmproto.Header{
			ChainID: chainID,
			Height:  1,
			Time:    s.coordinator.CurrentTime.UTC(),
		}

		chain.Coordinator = s.coordinator
		s.coordinator.CommitBlock(chain)

		chains[chainID] = chain
	}

	s.coordinator.Chains = chains
	s.chainA = s.coordinator.GetChain(ibctesting.GetChainID(0))
	s.chainB = s.coordinator.GetChain(ibctesting.GetChainID(1))

	umeeApp := s.getUmeeApp(s.chainA)

	queryHelper := baseapp.NewQueryServerTestHelper(s.chainA.GetContext(), umeeApp.InterfaceRegistry())
	types.RegisterQueryServer(queryHelper, umeeApp.TransferKeeper)
	s.queryClient = types.NewQueryClient(queryHelper)
}

func (suite *KeeperTestSuite) getUmeeApp(c *ibctesting.TestChain) *app.UmeeApp {
	umeeApp, ok := c.App.(*app.UmeeApp)
	suite.Require().True(ok)

	return umeeApp
}

func TestKeeperTestSuite(t *testing.T) {
	suite.Run(t, new(KeeperTestSuite))
}

func (s *KeeperTestSuite) TestSendTransfer() {

}

func (s *KeeperTestSuite) TestOnRecvPacket() {

}
