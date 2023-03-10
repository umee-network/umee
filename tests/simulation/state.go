package simulation

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"os"
	"testing"
	"time"

	sdkmath "cosmossdk.io/math"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/simapp"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	"github.com/cosmos/cosmos-sdk/store"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	tmjson "github.com/tendermint/tendermint/libs/json"
	"github.com/tendermint/tendermint/libs/log"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"
	"gotest.tools/v3/assert"

	umeeapp "github.com/umee-network/umee/v4/app"
	appparams "github.com/umee-network/umee/v4/app/params"
)

// GenesisState of the blockchain is represented here as a map of raw json
// messages key'd by a identifier string.
// The identifier is used to determine which module genesis information belongs
// to so it may be appropriately routed during init chain.
// Within this application default genesis information is retrieved from
// the ModuleBasicManager which populates json from each BasicModule
// object provided to it during init.
type GenesisState map[string]json.RawMessage

type StoreKeysPrefixes struct {
	A        storetypes.StoreKey
	B        storetypes.StoreKey
	Prefixes [][]byte
}

// interBlockCacheOpt returns a BaseApp option function that sets the persistent
// inter-block write-through cache.
func interBlockCacheOpt() func(*baseapp.BaseApp) {
	return baseapp.SetInterBlockCache(store.NewCommitKVStoreCacheManager())
}

// fauxMerkleModeOpt returns a BaseApp option to use a dbStoreAdapter instead of
// an IAVLStore for faster simulation speed.
func fauxMerkleModeOpt(bapp *baseapp.BaseApp) {
	bapp.SetFauxMerkleMode()
}

// appStateFn returns the initial application state using a genesis file or
// simulation parameters. It panics if the user provides files for both of them.
// If a file is not given for the genesis or the sim params, it creates a
// randomized one.
func appStateFn(cdc codec.JSONCodec, simManager *module.SimulationManager) simtypes.AppStateFn {
	return func(
		r *rand.Rand,
		accs []simtypes.Account,
		config simtypes.Config,
	) (appState json.RawMessage, simAccs []simtypes.Account, chainID string, genesisTimestamp time.Time) {
		if simapp.FlagGenesisTimeValue == 0 {
			genesisTimestamp = simtypes.RandTimestamp(r)
		} else {
			genesisTimestamp = time.Unix(simapp.FlagGenesisTimeValue, 0)
		}

		chainID = config.ChainID
		switch {
		case config.ParamsFile != "" && config.GenesisFile != "":
			panic("cannot provide both a genesis file and a params file")

		case config.GenesisFile != "":
			// override the default chain-id from simapp to set it later to the config
			genesisDoc, accounts := appStateFromGenesisFileFn(r, cdc, config.GenesisFile)

			if simapp.FlagGenesisTimeValue == 0 {
				// use genesis timestamp if no custom timestamp is provided (i.e no random timestamp)
				genesisTimestamp = genesisDoc.GenesisTime
			}

			appState = genesisDoc.AppState
			chainID = genesisDoc.ChainID
			simAccs = accounts

		case config.ParamsFile != "":
			appParams := make(simtypes.AppParams)
			bz, err := ioutil.ReadFile(config.ParamsFile)
			if err != nil {
				panic(err)
			}

			if err := json.Unmarshal(bz, &appParams); err != nil {
				panic(err)
			}

			appState, simAccs = appStateRandomizedFn(simManager, r, cdc, accs, genesisTimestamp, appParams)

		default:
			appParams := make(simtypes.AppParams)
			appState, simAccs = appStateRandomizedFn(simManager, r, cdc, accs, genesisTimestamp, appParams)
		}

		rawState := make(map[string]json.RawMessage)
		if err := json.Unmarshal(appState, &rawState); err != nil {
			panic(err)
		}

		stakingStateBz, ok := rawState[stakingtypes.ModuleName]
		if !ok {
			panic("staking genesis state is missing")
		}

		stakingState := new(stakingtypes.GenesisState)
		if err := cdc.UnmarshalJSON(stakingStateBz, stakingState); err != nil {
			panic(err)
		}

		// compute not bonded balance
		notBondedTokens := sdk.ZeroInt()
		for _, val := range stakingState.Validators {
			if val.Status != stakingtypes.Unbonded {
				continue
			}

			notBondedTokens = notBondedTokens.Add(val.GetTokens())
		}

		notBondedCoins := sdk.NewCoin(stakingState.Params.BondDenom, notBondedTokens)

		// edit bank state to make it have the not bonded pool tokens
		bankStateBz, ok := rawState[banktypes.ModuleName]
		if !ok {
			panic("bank genesis state is missing")
		}

		bankState := new(banktypes.GenesisState)
		if err := cdc.UnmarshalJSON(bankStateBz, bankState); err != nil {
			panic(err)
		}

		stakingAddr := authtypes.NewModuleAddress(stakingtypes.NotBondedPoolName).String()

		var found bool
		for _, balance := range bankState.Balances {
			if balance.Address == stakingAddr {
				found = true
				break
			}
		}

		if !found {
			bankState.Balances = append(bankState.Balances, banktypes.Balance{
				Address: stakingAddr,
				Coins:   sdk.NewCoins(notBondedCoins),
			})
		}

		rawState[stakingtypes.ModuleName] = cdc.MustMarshalJSON(stakingState)
		rawState[banktypes.ModuleName] = cdc.MustMarshalJSON(bankState)

		appState, err := json.Marshal(rawState)
		if err != nil {
			panic(err)
		}

		return appState, simAccs, chainID, genesisTimestamp
	}
}

// appStateRandomizedFn creates calls each module's GenesisState generator
// function and creates the simulation params.
func appStateRandomizedFn(
	simManager *module.SimulationManager,
	r *rand.Rand,
	cdc codec.JSONCodec,
	accs []simtypes.Account,
	genesisTimestamp time.Time,
	appParams simtypes.AppParams,
) (json.RawMessage, []simtypes.Account) {
	numAccs := int64(len(accs))
	genesisState := umeeapp.NewDefaultGenesisState(cdc)

	// Generate a random amount of initial stake coins and a random initial
	// number of bonded accounts.
	var numInitiallyBonded int64
	var initialStake sdkmath.Int
	appParams.GetOrGenerate(
		cdc,
		simappparams.StakePerAccount,
		&initialStake,
		r,
		func(r *rand.Rand) { initialStake = sdkmath.NewIntFromUint64(uint64(r.Int63n(1e12))) },
	)
	appParams.GetOrGenerate(
		cdc,
		simappparams.InitiallyBondedValidators,
		&numInitiallyBonded,
		r,
		func(r *rand.Rand) { numInitiallyBonded = int64(r.Intn(300)) },
	)

	if numInitiallyBonded > numAccs {
		numInitiallyBonded = numAccs
	}

	fmt.Printf(
		`Selected randomly generated parameters for simulated genesis:
{
  stake_per_account: "%d",
  initially_bonded_validators: "%d"
}
`, initialStake, numInitiallyBonded,
	)

	simState := &module.SimulationState{
		AppParams:    appParams,
		Cdc:          cdc,
		Rand:         r,
		GenState:     genesisState,
		Accounts:     accs,
		InitialStake: initialStake,
		NumBonded:    numInitiallyBonded,
		GenTimestamp: genesisTimestamp,
	}

	simManager.GenerateGenesisStates(simState)

	appState, err := json.Marshal(genesisState)
	if err != nil {
		panic(err)
	}

	return appState, accs
}

// appStateFromGenesisFileFn generates genesis AppState from a genesis.json file.
func appStateFromGenesisFileFn(
	r io.Reader,
	cdc codec.JSONCodec,
	genesisFile string,
) (tmtypes.GenesisDoc, []simtypes.Account) {
	bytes, err := ioutil.ReadFile(genesisFile)
	if err != nil {
		panic(err)
	}

	// NOTE: Tendermint uses a custom JSON decoder for GenesisDoc
	var genesis tmtypes.GenesisDoc
	if err := tmjson.Unmarshal(bytes, &genesis); err != nil {
		panic(err)
	}

	var appState umeeapp.GenesisState
	if err := json.Unmarshal(genesis.AppState, &appState); err != nil {
		panic(err)
	}

	var authGenesis authtypes.GenesisState
	if appState[authtypes.ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[authtypes.ModuleName], &authGenesis)
	}

	newAccs := make([]simtypes.Account, len(authGenesis.Accounts))
	for i, acc := range authGenesis.Accounts {
		// Pick a random private key, since we don't know the actual key
		// This should be fine as it's only used for mock Tendermint validators
		// and these keys are never actually used to sign by mock Tendermint.
		privkeySeed := make([]byte, 15)
		if _, err := r.Read(privkeySeed); err != nil {
			panic(err)
		}

		privKey := secp256k1.GenPrivKeyFromSecret(privkeySeed)

		a, ok := acc.GetCachedValue().(authtypes.AccountI)
		if !ok {
			panic("expected account")
		}

		// create simulator accounts
		simAcc := simtypes.Account{PrivKey: privKey, PubKey: privKey.PubKey(), Address: a.GetAddress()}
		newAccs[i] = simAcc
	}

	return genesis, newAccs
}

func appExportAndImport(t *testing.T) (
	dbm.DB, string, *umeeapp.UmeeApp, log.Logger, servertypes.ExportedApp, bool, dbm.DB, string, *umeeapp.UmeeApp,
	simtypes.Config,
) {
	config, db, dir, logger, skip, err := simapp.SetupSimulation("leveldb-app-sim", "Simulation")
	if skip {
		t.Skip("skipping application simulation")
	}

	assert.NilError(t, err, "simulation setup failed")

	app := umeeapp.New(
		logger,
		db,
		nil,
		true,
		map[int64]bool{},
		dir,
		simapp.FlagPeriodValue,
		umeeapp.MakeEncodingConfig(),
		umeeapp.EmptyAppOptions{},
		umeeapp.GetWasmEnabledProposals(),
		umeeapp.EmptyWasmOpts,
		fauxMerkleModeOpt,
	)
	assert.Equal(t, appparams.Name, app.Name())

	// run randomized simulation
	stopEarly, simParams, simErr := simulation.SimulateFromSeed(
		t,
		os.Stdout,
		app.BaseApp,
		appStateFn(app.AppCodec(), app.StateSimulationManager),
		simtypes.RandomAccounts,
		simapp.SimulationOperations(app, app.AppCodec(), config),
		app.ModuleAccountAddrs(),
		config,
		app.AppCodec(),
	)

	// export state and simParams before the simulation error is checked
	err = simapp.CheckExportSimulation(app, config, simParams)
	assert.NilError(t, err)
	assert.NilError(t, simErr)

	if config.Commit {
		simapp.PrintStats(db)
	}

	fmt.Printf("exporting genesis...\n")

	exported, err := app.ExportAppStateAndValidators(false, []string{})
	assert.NilError(t, err)

	fmt.Printf("importing genesis...\n")

	config, newDB, newDir, _, _, err := simapp.SetupSimulation("leveldb-app-sim-2", "Simulation-2")
	assert.NilError(t, err, "simulation setup failed")

	newApp := umeeapp.New(
		logger,
		newDB,
		nil,
		true,
		map[int64]bool{},
		newDir,
		simapp.FlagPeriodValue,
		umeeapp.MakeEncodingConfig(),
		umeeapp.EmptyAppOptions{},
		umeeapp.GetWasmEnabledProposals(),
		umeeapp.EmptyWasmOpts,
		fauxMerkleModeOpt,
	)
	assert.Equal(t, appparams.Name, newApp.Name())

	return db, dir, app, logger, exported, stopEarly, newDB, newDir, newApp, config
}
