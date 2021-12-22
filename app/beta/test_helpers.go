package beta

import (
	"encoding/json"
	"fmt"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	dbm "github.com/tendermint/tm-db"

	"github.com/umee-network/umee/app"

	sdk "github.com/cosmos/cosmos-sdk/types"
	leveragetypes "github.com/umee-network/umee/x/leverage/types"
)

func Setup(t *testing.T, isCheckTx bool, invCheckPeriod uint) *UmeeApp {
	t.Helper()

	betaApp, genesisState := setup(!isCheckTx, invCheckPeriod)
	if !isCheckTx {
		// InitChain must be called to stop deliverState from being nil
		stateBytes, err := json.MarshalIndent(genesisState, "", " ")
		require.NoError(t, err)

		// Initialize the chain
		betaApp.InitChain(
			abci.RequestInitChain{
				Validators:      []abci.ValidatorUpdate{},
				ConsensusParams: app.DefaultConsensusParams,
				AppStateBytes:   stateBytes,
			},
		)
	}

	return betaApp
}

func setup(withGenesis bool, invCheckPeriod uint) (*UmeeApp, app.GenesisState) {
	db := dbm.NewMemDB()
	encCdc := MakeEncodingConfig()
	betaApp := New(
		log.NewNopLogger(),
		db,
		nil,
		true,
		map[int64]bool{},
		app.DefaultNodeHome,
		invCheckPeriod,
		encCdc,
		app.EmptyAppOptions{},
	)
	if withGenesis {
		return betaApp, NewDefaultGenesisState(encCdc.Marshaler)
	}

	return betaApp, app.GenesisState{}
}

// IntegrationTestNetworkConfig returns a networking configuration used for
// integration tests using the SDK's in-process network test suite.
func IntegrationTestNetworkConfig() network.Config {
	cfg := network.DefaultConfig()
	encCfg := MakeEncodingConfig()
	cdc := encCfg.Marshaler

	// Start with the default genesis state
	appGenState := ModuleBasics.DefaultGenesis(encCfg.Marshaler)

	// Extract the x/leverage part of the genesis state to be modified
	var leverageGenState leveragetypes.GenesisState
	if err := cdc.UnmarshalJSON(appGenState[leveragetypes.ModuleName], &leverageGenState); err != nil {
		panic(err)
	}

	// Modify the x/leverage genesis state
	leverageGenState.Registry = append(leverageGenState.Registry, leveragetypes.Token{
		BaseDenom:            app.BondDenom,
		SymbolDenom:          "UMEE",
		Exponent:             6,
		ReserveFactor:        sdk.MustNewDecFromStr("0.100000000000000000"),
		CollateralWeight:     sdk.MustNewDecFromStr("0.050000000000000000"),
		BaseBorrowRate:       sdk.MustNewDecFromStr("0.020000000000000000"),
		KinkBorrowRate:       sdk.MustNewDecFromStr("0.200000000000000000"),
		MaxBorrowRate:        sdk.MustNewDecFromStr("1.50000000000000000"),
		KinkUtilizationRate:  sdk.MustNewDecFromStr("0.200000000000000000"),
		LiquidationIncentive: sdk.MustNewDecFromStr("0.180000000000000000"),
	})

	// Marshal the modified state and add it back into appGenState
	bz, err := cdc.MarshalJSON(&leverageGenState)
	if err != nil {
		panic(err)
	}
	appGenState[leveragetypes.ModuleName] = bz

	cfg.Codec = encCfg.Marshaler
	cfg.TxConfig = encCfg.TxConfig
	cfg.LegacyAmino = encCfg.Amino
	cfg.InterfaceRegistry = encCfg.InterfaceRegistry
	cfg.GenesisState = appGenState
	cfg.BondDenom = BondDenom
	cfg.MinGasPrices = fmt.Sprintf("0.000006%s", BondDenom)
	cfg.AppConstructor = func(val network.Validator) servertypes.Application {
		return New(
			val.Ctx.Logger,
			dbm.NewMemDB(),
			nil,
			true,
			make(map[int64]bool),
			val.Ctx.Config.RootDir,
			0,
			encCfg,
			app.EmptyAppOptions{},
			baseapp.SetPruning(storetypes.NewPruningOptionsFromString(val.AppConfig.Pruning)),
			baseapp.SetMinGasPrices(val.AppConfig.MinGasPrices),
		)
	}

	return cfg
}
