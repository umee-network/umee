package tests

import (
	"testing"

	"github.com/umee-network/umee/v6/x/metoken"
	"github.com/umee-network/umee/v6/x/metoken/mocks"

	umeeapp "github.com/umee-network/umee/v6/app"
)

func TestIntegrationSuite(t *testing.T) {
	t.Parallel()
	cfg := umeeapp.IntegrationTestNetworkConfig()
	cfg.NumValidators = 2
	cfg.Mnemonics = []string{
		"empower ridge mystery shrimp predict alarm swear brick across funny vendor essay antique vote place lava proof gaze crush head east arch twin lady",
		"clean target advice dirt onion correct original vibrant actor upon waste eternal color barely shrimp aspect fall material wait repeat bench demise length seven",
	}

	var metokenGenState metoken.GenesisState
	if err := cfg.Codec.UnmarshalJSON(cfg.GenesisState[metoken.ModuleName], &metokenGenState); err != nil {
		panic(err)
	}

	metokenGenState.Registry = []metoken.Index{mocks.BondIndex()}
	metokenGenState.Balances = []metoken.IndexBalances{mocks.BondBalance()}

	bz, err := cfg.Codec.MarshalJSON(&metokenGenState)
	if err != nil {
		panic(err)
	}
	cfg.GenesisState[metoken.ModuleName] = bz

	// init the integration test and start the network
	s := NewIntegrationTestSuite(cfg, t)
	s.SetupSuite()
	defer s.TearDownSuite()

	// test cli queries
	s.TestInvalidQueries()
	s.TestValidQueries()

	// test cli transactions
	s.TestTransactions()
}
