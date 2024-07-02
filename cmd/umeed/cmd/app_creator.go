package cmd

import (
	"errors"
	"fmt"
	"io"
	stdlog "log"
	"os"
	"path/filepath"

	"cosmossdk.io/log"
	"cosmossdk.io/store"
	"cosmossdk.io/store/snapshots"
	snapshottypes "cosmossdk.io/store/snapshots/types"
	storetypes "cosmossdk.io/store/types"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	dbm "github.com/cosmos/cosmos-db"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client/flags"
	"github.com/cosmos/cosmos-sdk/server"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/mempool"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/types/module/testutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/prometheus/client_golang/prometheus"
	"github.com/spf13/cast"

	"github.com/umee-network/umee/v6/ante"
	umeeapp "github.com/umee-network/umee/v6/app"
	appparams "github.com/umee-network/umee/v6/app/params"
)

type appCreator struct {
	encCfg        testutil.TestEncodingConfig
	moduleManager module.BasicManager
}

func (a appCreator) newApp(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	appOpts servertypes.AppOptions,
) servertypes.Application {
	var cache storetypes.MultiStorePersistentCache

	if cast.ToBool(appOpts.Get(server.FlagInterBlockCache)) {
		cache = store.NewCommitKVStoreCacheManager()
	}

	skipUpgradeHeights := make(map[int64]bool)
	for _, h := range cast.ToIntSlice(appOpts.Get(server.FlagUnsafeSkipUpgrades)) {
		skipUpgradeHeights[int64(h)] = true
	}

	pruningOpts, err := server.GetPruningOptionsFromFlags(appOpts)
	if err != nil {
		panic(fmt.Sprintf("failed to parse pruning options from flags: %s", err))
	}

	snapshotDir := filepath.Join(cast.ToString(appOpts.Get(flags.FlagHome)), "data", "snapshots")
	snapshotDB, err := dbm.NewDB("metadata", server.GetAppDBBackend(appOpts), snapshotDir)
	if err != nil {
		panic(fmt.Sprintf("failed to create LevelDB database: %s", err))
	}

	snapshotStore, err := snapshots.NewStore(snapshotDB, snapshotDir)
	if err != nil {
		panic(fmt.Sprintf("failed to create snapshot store: %s", err))
	}
	snapshotOptions := snapshottypes.NewSnapshotOptions(
		cast.ToUint64(appOpts.Get(server.FlagStateSyncSnapshotInterval)),
		cast.ToUint32(appOpts.Get(server.FlagStateSyncSnapshotKeepRecent)),
	)

	minGasPrices := cast.ToString(appOpts.Get(server.FlagMinGasPrices))
	mustMinUmeeGasPrice(minGasPrices)

	var wasmOpts []wasmkeeper.Option
	if cast.ToBool(appOpts.Get("telemetry.enabled")) {
		wasmOpts = append(wasmOpts, wasmkeeper.WithVMCacheMetrics(prometheus.DefaultRegisterer))
	}

	homeDir := cast.ToString(appOpts.Get(flags.FlagHome))
	chainID := cast.ToString(appOpts.Get(flags.FlagChainID))
	if chainID == "" {
		// fallback to genesis chain-id
		reader, err := os.Open(filepath.Join(homeDir, "config", "genesis.json"))
		if err != nil {
			panic(err)
		}
		defer reader.Close()

		chainID, err = genutiltypes.ParseChainIDFromGenesis(reader)
		if err != nil {
			panic(fmt.Errorf("failed to parse chain-id from genesis file: %w", err))
		}
	}

	defaultMempool := baseapp.SetMempool(mempool.NoOpMempool{})
	if maxTxs := cast.ToInt(appOpts.Get(server.FlagMempoolMaxTxs)); maxTxs >= 0 {
		defaultMempool = baseapp.SetMempool(
			mempool.NewSenderNonceMempool(
				mempool.SenderNonceMaxTxOpt(maxTxs),
			),
		)
	}

	return umeeapp.New(
		logger, db, traceStore, true, skipUpgradeHeights,
		cast.ToString(appOpts.Get(flags.FlagHome)),
		cast.ToUint(appOpts.Get(server.FlagInvCheckPeriod)),
		appOpts,
		wasmOpts,
		baseapp.SetPruning(pruningOpts),
		baseapp.SetMinGasPrices(minGasPrices),
		baseapp.SetMinRetainBlocks(cast.ToUint64(appOpts.Get(server.FlagMinRetainBlocks))),
		baseapp.SetHaltHeight(cast.ToUint64(appOpts.Get(server.FlagHaltHeight))),
		baseapp.SetHaltTime(cast.ToUint64(appOpts.Get(server.FlagHaltTime))),
		baseapp.SetInterBlockCache(cache),
		baseapp.SetTrace(cast.ToBool(appOpts.Get(server.FlagTrace))),
		baseapp.SetIndexEvents(cast.ToStringSlice(appOpts.Get(server.FlagIndexEvents))),
		baseapp.SetSnapshot(snapshotStore, snapshotOptions),
		baseapp.SetChainID(chainID),
		baseapp.SetIAVLCacheSize(cast.ToInt(appOpts.Get(server.FlagIAVLCacheSize))),
		baseapp.SetIAVLDisableFastNode(cast.ToBool(appOpts.Get(server.FlagDisableIAVLFastNode))),
		defaultMempool,
		baseapp.SetQueryGasLimit(cast.ToUint64(appOpts.Get(server.FlagQueryGasLimit))),
	)
}

func mustMinUmeeGasPrice(minGasPrices string) {
	gasPrices, err := sdk.ParseDecCoins(minGasPrices)
	if err != nil {
		stdlog.Fatalf("invalid minimum gas prices: %v", err)
	}
	if err := ante.AssertMinProtocolGasPrice(gasPrices); err != nil {
		stdlog.Fatal("minimum-gas-price config in app.toml must be at least ",
			appparams.ProtocolMinGasPrice, " [", err, "]")
	}
}

// appExport creates a new simapp, optionally at a given height.
func (a appCreator) appExport(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	height int64,
	forZeroHeight bool,
	jailAllowedAddrs []string,
	appOpts servertypes.AppOptions,
	modulesToExport []string,
) (servertypes.ExportedApp, error) {
	homePath, ok := appOpts.Get(flags.FlagHome).(string)
	if !ok || homePath == "" {
		return servertypes.ExportedApp{}, errors.New("application home is not set")
	}

	var loadLatest bool
	if height == -1 {
		loadLatest = true
	}

	app := umeeapp.New(
		logger,
		db,
		traceStore,
		loadLatest,
		map[int64]bool{},
		homePath,
		uint(1),
		appOpts,
		umeeapp.EmptyWasmOpts,
	)

	if height != -1 {
		if err := app.LoadHeight(height); err != nil {
			return servertypes.ExportedApp{}, err
		}
	}

	return app.ExportAppStateAndValidators(forZeroHeight, jailAllowedAddrs, modulesToExport)
}
