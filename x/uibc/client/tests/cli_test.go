package tests

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	umeeapp "github.com/umee-network/umee/v4/app"
	"github.com/umee-network/umee/v4/x/uibc"
	"gotest.tools/v3/assert"
)

func TestIntegrationSuite(t *testing.T) {
	t.Parallel()
	cfg := umeeapp.IntegrationTestNetworkConfig()
	cfg.NumValidators = 2
	cfg.Mnemonics = []string{
		"empower ridge mystery shrimp predict alarm swear brick across funny vendor essay antique vote place lava proof gaze crush head east arch twin lady",
		"clean target advice dirt onion correct original vibrant actor upon waste eternal color barely shrimp aspect fall material wait repeat bench demise length seven",
	}

	var uibcGenState uibc.GenesisState
	assert.NilError(t, cfg.Codec.UnmarshalJSON(cfg.GenesisState[uibc.ModuleName], &uibcGenState))
	uibcGenState.Quotas = sdk.DecCoins{sdk.NewInt64DecCoin("uumee", 0)}
	uibcGenState.TotalOutflowSum = sdk.NewDec(10)

	bz, err = cfg.Codec.MarshalJSON(&uibcGenState)
	assert.NilError(t, err)
	cfg.GenesisState[uibc.ModuleName] = bz

	// init the integration test and start the network
	s := initIntegrationTestSuite(cfg, t)

	// test cli queries
	s.TestGetQuota(t)
	s.TestQueryParams(t)

	// tear down netowkr
	tearDownSuite(s, t)
}
