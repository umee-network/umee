package tests

import (
	"testing"

	gravitytypes "github.com/Gravity-Bridge/Gravity-Bridge/module/x/gravity/types"
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

	var gravityGenState gravitytypes.GenesisState
	assert.NilError(t, cfg.Codec.UnmarshalJSON(cfg.GenesisState[gravitytypes.ModuleName], &gravityGenState))

	gravityGenState.DelegateKeys = []gravitytypes.MsgSetOrchestratorAddress{
		{
			Validator:    "umeevaloper1t3ass54lpra0klz047k0dax33ckahym5phdrpg",
			Orchestrator: "umee1t3ass54lpra0klz047k0dax33ckahym5pn2vsz",
			EthAddress:   "0x9fc56f2e851e1ab2b4c0fc4f6344800f29652ffe",
		},
		{
			Validator:    "umeevaloper1kqh6nt4f48vptvq4j5cgr0nfd2x4z9ulvrtqrh",
			Orchestrator: "umee1kqh6nt4f48vptvq4j5cgr0nfd2x4z9ulv8v0ja",
			EthAddress:   "0xddfda961410b2815b48679377baa0009ace173a2",
		},
	}

	bz, err := cfg.Codec.MarshalJSON(&gravityGenState)
	assert.NilError(t, err)
	cfg.GenesisState[gravitytypes.ModuleName] = bz

	var uibcGenState uibc.GenesisState
	assert.NilError(t, cfg.Codec.UnmarshalJSON(cfg.GenesisState[uibc.ModuleName], &uibcGenState))
	uibcGenState.Outflows = sdk.DecCoins{sdk.NewInt64DecCoin("uumee", 0)}
	uibcGenState.TotalOutflowSum = sdk.NewDec(10)

	bz, err = cfg.Codec.MarshalJSON(&uibcGenState)
	assert.NilError(t, err)
	cfg.GenesisState[uibc.ModuleName] = bz

	// init the integration test and start the network
	s := NewIntegrationTestSuite(cfg, t)
	s.Suite.SetT(t)
	s.SetupSuite()

	// test cli queries
	s.TestGetQuota(t)
	s.TestQueryParams(t)

	s.TearDownSuite()
}
