package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"
	"os"
	"testing"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/server"
	"github.com/cosmos/cosmos-sdk/simapp"
	"github.com/cosmos/cosmos-sdk/store"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"
	"github.com/rs/zerolog"
	"github.com/stretchr/testify/require"
	"github.com/tendermint/tendermint/libs/log"
	tmrand "github.com/tendermint/tendermint/libs/rand"
	dbm "github.com/tendermint/tm-db"

	umeeapp "github.com/umee-network/umee/app"
)

func init() {
	simapp.GetSimulatorFlags()
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

// TestFullAppSimulation tests application fuzzing given a random seed as input.
func TestFullAppSimulation(t *testing.T) {
	config, db, dir, _, skip, err := simapp.SetupSimulation("leveldb-app-sim", "Simulation")
	if skip {
		t.Skip("skipping application simulation")
	}

	require.NoError(t, err, "simulation setup failed")

	defer func() {
		db.Close()
		require.NoError(t, os.RemoveAll(dir))
	}()

	var logger log.Logger
	if simapp.FlagVerboseValue {
		logger = server.ZeroLogWrapper{
			Logger: zerolog.New(os.Stderr).Level(zerolog.InfoLevel).With().Timestamp().Logger(),
		}
	} else {
		logger = server.ZeroLogWrapper{
			Logger: zerolog.Nop(),
		}
	}

	app := umeeapp.New(
		logger,
		db,
		nil,
		true,
		map[int64]bool{},
		umeeapp.DefaultNodeHome,
		simapp.FlagPeriodValue,
		umeeapp.MakeEncodingConfig(),
		umeeapp.EmptyAppOptions{},
		fauxMerkleModeOpt,
	)
	require.Equal(t, umeeapp.Name, app.Name())

	// run randomized simulation
	_, simParams, simErr := simulation.SimulateFromSeed(
		t,
		os.Stdout,
		app.BaseApp,
		appStateFn(app.AppCodec(), app.SimulationManager()),
		simtypes.RandomAccounts,
		simapp.SimulationOperations(app, app.AppCodec(), config),
		app.ModuleAccountAddrs(),
		config,
		app.AppCodec(),
	)

	// export state and simParams before the simulation error is checked
	err = CheckExportSimulation(app, config, simParams)
	require.NoError(t, err)
	require.NoError(t, simErr)

	if config.Commit {
		PrintStats(db)
	}
}

// TestAppStateDeterminism tests for application non-determinism using a PRNG
// as an input for the simulator's seed.
func TestAppStateDeterminism(t *testing.T) {
	if !simapp.FlagEnabledValue {
		t.Skip("skipping application simulation")
	}

	config := simapp.NewConfigFromFlags()
	config.InitialBlockHeight = 1
	config.ExportParamsPath = ""
	config.OnOperation = false
	config.AllInvariants = false
	config.ChainID = fmt.Sprintf("simulation-chain-%s", tmrand.NewRand().Str(6))

	numSeeds := 3
	numTimesToRunPerSeed := 5
	appHashList := make([]json.RawMessage, numTimesToRunPerSeed)

	for i := 0; i < numSeeds; i++ {
		config.Seed = rand.Int63()

		for j := 0; j < numTimesToRunPerSeed; j++ {
			var logger log.Logger
			if simapp.FlagVerboseValue {
				logger = server.ZeroLogWrapper{
					Logger: zerolog.New(os.Stderr).Level(zerolog.InfoLevel).With().Timestamp().Logger(),
				}
			} else {
				logger = server.ZeroLogWrapper{
					Logger: zerolog.Nop(),
				}
			}

			db := dbm.NewMemDB()
			app := umeeapp.New(
				logger,
				db,
				nil,
				true,
				map[int64]bool{},
				umeeapp.DefaultNodeHome,
				simapp.FlagPeriodValue,
				umeeapp.MakeEncodingConfig(),
				umeeapp.EmptyAppOptions{},
				interBlockCacheOpt(),
			)

			fmt.Printf(
				"running non-determinism simulation; seed %d; run: %d/%d; attempt: %d/%d\n",
				config.Seed, i+1, numSeeds, j+1, numTimesToRunPerSeed,
			)

			_, _, err := simulation.SimulateFromSeed(
				t,
				os.Stdout,
				app.BaseApp,
				appStateFn(app.AppCodec(), app.SimulationManager()),
				simtypes.RandomAccounts,
				simapp.SimulationOperations(app, app.AppCodec(), config),
				app.ModuleAccountAddrs(),
				config,
				app.AppCodec(),
			)
			require.NoError(t, err)

			if config.Commit {
				simapp.PrintStats(db)
			}

			appHash := app.LastCommitID().Hash
			appHashList[j] = appHash

			if j != 0 {
				require.Equal(
					t,
					string(appHashList[0]),
					string(appHashList[j]),
					"non-determinism in seed %d; run: %d/%d; attempt: %d/%d\n",
					config.Seed, i+1, numSeeds, j+1, numTimesToRunPerSeed,
				)
			}
		}
	}
}
