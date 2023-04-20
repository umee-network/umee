package integrationsuite_test

import (
	"github.com/cosmos/cosmos-sdk/testutil/network"
	"github.com/stretchr/testify/suite"
)

type IntegrationTestSuite struct {
	suite.Suite

	Cfg     network.Config
	Network *network.Network
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T().Log("setting up integration test suite")

	network, err := network.New(s.T(), s.T().TempDir(), s.Cfg)
	s.Require().NoError(err)
	s.Network = network

	_, err = s.Network.WaitForHeight(1)
	s.Require().NoError(err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T().Log("tearing down integration test suite")

	s.Network.Cleanup()
}

// runTestQuery
func (s *IntegrationTestSuite) RunTestQueries(tqs ...TestQuery) {
	for _, t := range tqs {
		// since steps of this test suite depend on previous transactions, we want to stop
		// on the first failure, rather than continue producing potentially inaccurate
		// errors as an effect of the first.
		// t.Run(s) stops properly, whereas s.Run would not
		s.Run(t.Msg, func() {
			t.Run(s)
		})
	}
}

// runTestCases runs test transactions or queries, stopping early if an error occurs
func (s *IntegrationTestSuite) RunTestTransactions(txs ...TestTransaction) {
	for _, t := range txs {
		// since steps of this test suite depend on previous transactions, we want to stop
		// on the first failure, rather than continue producing potentially inaccurate
		// errors as an effect of the first.
		// t.Run(s) stops properly, whereas s.Run would not
		s.Run(t.Msg, func() {
			t.Run(s)
		})
	}
}
