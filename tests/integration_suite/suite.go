package integrationsuite_test

import (
	"testing"

	"github.com/cosmos/cosmos-sdk/testutil/network"
	"gotest.tools/v3/assert"
)

type IntegrationTestSuite struct {
	T       *testing.T
	Cfg     network.Config
	Network *network.Network
}

func (s *IntegrationTestSuite) SetupSuite() {
	s.T.Log("setting up integration test suite")

	network, err := network.New(s.T, s.T.TempDir(), s.Cfg)
	assert.NilError(s.T, err)
	s.Network = network

	_, err = s.Network.WaitForHeight(1)
	assert.NilError(s.T, err)
}

func (s *IntegrationTestSuite) TearDownSuite() {
	s.T.Log("tearing down integration test suite")

	s.Network.Cleanup()
}

// runTestQuery
func (s *IntegrationTestSuite) RunTestQueries(tqs ...TestQuery) {
	for _, tq := range tqs {
		// since steps of this test suite depend on previous transactions, we want to stop
		// on the first failure, rather than continue producing potentially inaccurate
		// errors as an effect of the first.
		// t.Run(s) stops properly, whereas s.Run would not
		s.T.Run(tq.Msg, func(t *testing.T) {
			tq.Run(s)
		})
	}
}

// runTestCases runs test transactions or queries, stopping early if an error occurs
func (s *IntegrationTestSuite) RunTestTransactions(txs ...TestTransaction) {
	for _, tx := range txs {
		// since steps of this test suite depend on previous transactions, we want to stop
		// on the first failure, rather than continue producing potentially inaccurate
		// errors as an effect of the first.
		// t.Run(s) stops properly, whereas s.Run would not
		s.T.Run(tx.Msg, func(t *testing.T) {
			tx.Run(s)
		})
	}
}
