package app

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"

	"github.com/spf13/cast"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmos "github.com/tendermint/tendermint/libs/os"
	dbm "github.com/tendermint/tm-db"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/server/api"
	"github.com/cosmos/cosmos-sdk/server/config"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	"github.com/cosmos/cosmos-sdk/testutil/testdata"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/version"
	"github.com/cosmos/cosmos-sdk/x/auth"
	"github.com/cosmos/cosmos-sdk/x/auth/ante"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	"github.com/cosmos/cosmos-sdk/x/auth/posthandler"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtx "github.com/cosmos/cosmos-sdk/x/auth/tx"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/capability"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrclient "github.com/cosmos/cosmos-sdk/x/distribution/client"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	evidencekeeper "github.com/cosmos/cosmos-sdk/x/evidence/keeper"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	feegrantkeeper "github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	feegrantmodule "github.com/cosmos/cosmos-sdk/x/feegrant/module"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govclient "github.com/cosmos/cosmos-sdk/x/gov/client"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	"github.com/cosmos/cosmos-sdk/x/group"
	groupkeeper "github.com/cosmos/cosmos-sdk/x/group/keeper"
	groupmodule "github.com/cosmos/cosmos-sdk/x/group/module"
	"github.com/cosmos/cosmos-sdk/x/mint"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/nft"
	nftkeeper "github.com/cosmos/cosmos-sdk/x/nft/keeper"
	nftmodule "github.com/cosmos/cosmos-sdk/x/nft/module"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramsclient "github.com/cosmos/cosmos-sdk/x/params/client"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradeclient "github.com/cosmos/cosmos-sdk/x/upgrade/client"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"

	ibctransfer "github.com/cosmos/ibc-go/v5/modules/apps/transfer"
	ibctransferkeeper "github.com/cosmos/ibc-go/v5/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v5/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v5/modules/core"
	ibcclient "github.com/cosmos/ibc-go/v5/modules/core/02-client"
	ibcclientclient "github.com/cosmos/ibc-go/v5/modules/core/02-client/client"
	ibcclienttypes "github.com/cosmos/ibc-go/v5/modules/core/02-client/types"
	ibcporttypes "github.com/cosmos/ibc-go/v5/modules/core/05-port/types"
	ibchost "github.com/cosmos/ibc-go/v5/modules/core/24-host"
	ibckeeper "github.com/cosmos/ibc-go/v5/modules/core/keeper"
	ibctesting "github.com/cosmos/ibc-go/v5/testing/types"
	"github.com/osmosis-labs/bech32-ibc/x/bech32ibc"
	bech32ibckeeper "github.com/osmosis-labs/bech32-ibc/x/bech32ibc/keeper"
	bech32ibctypes "github.com/osmosis-labs/bech32-ibc/x/bech32ibc/types"

	// cosmwasm
	"github.com/CosmWasm/wasmd/x/wasm"
	wasmclient "github.com/CosmWasm/wasmd/x/wasm/client"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	customante "github.com/umee-network/umee/v4/ante"
	appparams "github.com/umee-network/umee/v4/app/params"
	"github.com/umee-network/umee/v4/swagger"
	"github.com/umee-network/umee/v4/util/genmap"
	"github.com/umee-network/umee/v4/x/leverage"
	leveragekeeper "github.com/umee-network/umee/v4/x/leverage/keeper"
	leveragetypes "github.com/umee-network/umee/v4/x/leverage/types"
	"github.com/umee-network/umee/v4/x/oracle"
	oraclekeeper "github.com/umee-network/umee/v4/x/oracle/keeper"
	oracletypes "github.com/umee-network/umee/v4/x/oracle/types"

	// umee ibc-transfer and quota for ibc-transfer
	"github.com/umee-network/umee/v4/x/uibc"
	uics20transfer "github.com/umee-network/umee/v4/x/uibc/ics20"
	uibctransferkeeper "github.com/umee-network/umee/v4/x/uibc/ics20/keeper"
	uibcmodule "github.com/umee-network/umee/v4/x/uibc/module"
	uibcquota "github.com/umee-network/umee/v4/x/uibc/quota"
	uibcquotakeeper "github.com/umee-network/umee/v4/x/uibc/quota/keeper"
)

var (
	_ CosmosApp               = (*UmeeApp)(nil)
	_ servertypes.Application = (*UmeeApp)(nil)

	// DefaultNodeHome defines the default home directory for the application
	// daemon.
	DefaultNodeHome string

	// ModuleBasics defines the module BasicManager is in charge of setting up basic,
	// non-dependant module elements, such as codec registration
	// and genesis verification.
	ModuleBasics module.BasicManager

	// module account permissions
	maccPerms map[string][]string
)

func init() {
	moduleBasics := []module.AppModuleBasic{
		auth.AppModuleBasic{},
		genutil.NewAppModuleBasic(genutiltypes.DefaultMessageValidator),
		BankModule{},
		capability.AppModuleBasic{},
		StakingModule{},
		MintModule{},
		distr.AppModuleBasic{},
		GovModule{AppModuleBasic: gov.NewAppModuleBasic(getGovProposalHandlers())},
		params.AppModuleBasic{},
		CrisisModule{},
		SlashingModule{},
		feegrantmodule.AppModuleBasic{},
		upgrade.AppModuleBasic{},
		evidence.AppModuleBasic{},
		authzmodule.AppModuleBasic{},
		groupmodule.AppModuleBasic{},
		vesting.AppModuleBasic{},
		nftmodule.AppModuleBasic{},
		ibc.AppModuleBasic{},
		ibctransfer.AppModuleBasic{},
		leverage.AppModuleBasic{},
		oracle.AppModuleBasic{},
		bech32ibc.AppModuleBasic{},
		uibcmodule.AppModuleBasic{},
	}

	if Experimental {
		moduleBasics = append(moduleBasics, wasm.AppModuleBasic{})
	}

	ModuleBasics = module.NewBasicManager(moduleBasics...)

	maccPerms = map[string][]string{
		authtypes.FeeCollectorName:     nil,
		distrtypes.ModuleName:          nil,
		minttypes.ModuleName:           {authtypes.Minter},
		stakingtypes.BondedPoolName:    {authtypes.Burner, authtypes.Staking},
		stakingtypes.NotBondedPoolName: {authtypes.Burner, authtypes.Staking},
		govtypes.ModuleName:            {authtypes.Burner},
		nft.ModuleName:                 nil,

		ibctransfertypes.ModuleName: {authtypes.Minter, authtypes.Burner},
		leveragetypes.ModuleName:    {authtypes.Minter, authtypes.Burner},
		oracletypes.ModuleName:      nil,
		uibc.ModuleName:             nil,
	}

	if Experimental {
		maccPerms[wasm.ModuleName] = []string{authtypes.Burner}
	}
}

// UmeeApp defines the ABCI application for the Umee network as an extension of
// the Cosmos SDK's BaseApp.
type UmeeApp struct {
	*baseapp.BaseApp

	legacyAmino       *codec.LegacyAmino
	appCodec          codec.Codec
	interfaceRegistry types.InterfaceRegistry
	txConfig          client.TxConfig

	invCheckPeriod uint

	// keys to access the substores
	keys    map[string]*storetypes.KVStoreKey
	tkeys   map[string]*storetypes.TransientStoreKey
	memKeys map[string]*storetypes.MemoryStoreKey

	// keepers
	AccountKeeper    authkeeper.AccountKeeper
	BankKeeper       bankkeeper.BaseKeeper
	CapabilityKeeper *capabilitykeeper.Keeper
	StakingKeeper    *stakingkeeper.Keeper
	SlashingKeeper   slashingkeeper.Keeper
	MintKeeper       mintkeeper.Keeper
	DistrKeeper      distrkeeper.Keeper
	GovKeeper        govkeeper.Keeper
	CrisisKeeper     crisiskeeper.Keeper
	UpgradeKeeper    upgradekeeper.Keeper
	ParamsKeeper     paramskeeper.Keeper
	AuthzKeeper      authzkeeper.Keeper
	EvidenceKeeper   evidencekeeper.Keeper
	FeeGrantKeeper   feegrantkeeper.Keeper
	GroupKeeper      groupkeeper.Keeper
	NFTKeeper        nftkeeper.Keeper
	WasmKeeper       wasm.Keeper

	UIBCTransferKeeper uibctransferkeeper.Keeper
	IBCKeeper          *ibckeeper.Keeper // IBC Keeper must be a pointer in the app, so we can SetRouter on it correctly
	LeverageKeeper     leveragekeeper.Keeper
	OracleKeeper       oraclekeeper.Keeper
	bech32IbcKeeper    bech32ibckeeper.Keeper
	UIbcQuotaKeeper    uibcquotakeeper.Keeper

	// make scoped keepers public for testing purposes
	ScopedIBCKeeper      capabilitykeeper.ScopedKeeper
	ScopedTransferKeeper capabilitykeeper.ScopedKeeper
	ScopedWasmKeeper     capabilitykeeper.ScopedKeeper

	// the module manager
	mm *module.Manager

	// simulation manager
	sm *module.SimulationManager
	// simulation manager to create state
	StateSimulationManager *module.SimulationManager

	// module configurator
	configurator module.Configurator

	// wasm
	wasmCfg wasm.Config
}

func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("failed to get user home directory: %s", err))
	}

	DefaultNodeHome = filepath.Join(userHomeDir, fmt.Sprintf(".%s", appparams.Name))
}

func New(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	skipUpgradeHeights map[int64]bool,
	homePath string,
	invCheckPeriod uint,
	encodingConfig appparams.EncodingConfig,
	appOpts servertypes.AppOptions,
	wasmEnabledProposals []wasm.ProposalType,
	wasmOpts []wasm.Option,
	baseAppOptions ...func(*baseapp.BaseApp),
) *UmeeApp {
	appCodec := encodingConfig.Codec
	legacyAmino := encodingConfig.Amino
	interfaceRegistry := encodingConfig.InterfaceRegistry

	bApp := baseapp.NewBaseApp(appparams.Name, logger, db, encodingConfig.TxConfig.TxDecoder(), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(interfaceRegistry)

	storeKeys := []string{
		authtypes.StoreKey, banktypes.StoreKey, stakingtypes.StoreKey,
		minttypes.StoreKey, distrtypes.StoreKey, slashingtypes.StoreKey,
		govtypes.StoreKey, paramstypes.StoreKey, upgradetypes.StoreKey, feegrant.StoreKey,
		evidencetypes.StoreKey, capabilitytypes.StoreKey,
		authzkeeper.StoreKey, nftkeeper.StoreKey, group.StoreKey,
		ibchost.StoreKey, ibctransfertypes.StoreKey,
		leveragetypes.StoreKey, oracletypes.StoreKey, bech32ibctypes.StoreKey,
		uibc.StoreKey,
	}
	if Experimental {
		storeKeys = append(storeKeys, wasm.StoreKey)
	}

	keys := sdk.NewKVStoreKeys(storeKeys...)
	tkeys := sdk.NewTransientStoreKeys(paramstypes.TStoreKey)
	memKeys := sdk.NewMemoryStoreKeys(capabilitytypes.MemStoreKey)
	govModuleAddr := authtypes.NewModuleAddress(govtypes.ModuleName).String()

	// configure state listening capabilities using AppOptions
	// we are doing nothing with the returned streamingServices and waitGroup in this case
	// if _, _, err := streaming.LoadStreamingServices(bApp, appOpts, appCodec, keys); err != nil {
	// 	tmos.Exit(err.Error())
	// }

	app := &UmeeApp{
		BaseApp:           bApp,
		legacyAmino:       legacyAmino,
		appCodec:          appCodec,
		interfaceRegistry: interfaceRegistry,
		txConfig:          encodingConfig.TxConfig,
		invCheckPeriod:    invCheckPeriod,
		keys:              keys,
		tkeys:             tkeys,
		memKeys:           memKeys,
	}

	app.ParamsKeeper = initParamsKeeper(
		appCodec,
		legacyAmino,
		keys[paramstypes.StoreKey],
		tkeys[paramstypes.TStoreKey],
	)

	// set the BaseApp's parameter store
	bApp.SetParamStore(app.ParamsKeeper.Subspace(baseapp.Paramspace).WithKeyTable(paramstypes.ConsensusParamsKeyTable()))

	app.CapabilityKeeper = capabilitykeeper.NewKeeper(
		appCodec,
		keys[capabilitytypes.StoreKey],
		memKeys[capabilitytypes.MemStoreKey],
	)

	// grant capabilities for the ibc and ibc-transfer modules
	app.ScopedIBCKeeper = app.CapabilityKeeper.ScopeToModule(ibchost.ModuleName)
	app.ScopedTransferKeeper = app.CapabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)

	app.initializeCustomScopedKeepers()

	// Applications that wish to enforce statically created ScopedKeepers should call `Seal` after creating
	// their scoped modules in `NewApp` with `ScopeToModule`
	app.CapabilityKeeper.Seal()

	app.AccountKeeper = authkeeper.NewAccountKeeper(
		appCodec,
		keys[authtypes.StoreKey],
		app.GetSubspace(authtypes.ModuleName),
		authtypes.ProtoBaseAccount,
		maccPerms,
		appparams.AccountAddressPrefix,
	)
	app.BankKeeper = bankkeeper.NewBaseKeeper(
		appCodec,
		keys[banktypes.StoreKey],
		app.AccountKeeper,
		app.GetSubspace(banktypes.ModuleName),
		app.ModuleAccountAddrs(),
	)
	_stakingKeeper := stakingkeeper.NewKeeper(
		appCodec, keys[stakingtypes.StoreKey], app.AccountKeeper, app.BankKeeper, app.GetSubspace(stakingtypes.ModuleName),
	)
	app.StakingKeeper = &_stakingKeeper
	app.MintKeeper = mintkeeper.NewKeeper(
		appCodec, keys[minttypes.StoreKey], app.GetSubspace(minttypes.ModuleName), app.StakingKeeper,
		app.AccountKeeper, app.BankKeeper, authtypes.FeeCollectorName,
	)
	app.DistrKeeper = distrkeeper.NewKeeper(
		appCodec, keys[distrtypes.StoreKey], app.GetSubspace(distrtypes.ModuleName), app.AccountKeeper, app.BankKeeper,
		app.StakingKeeper, authtypes.FeeCollectorName,
	)
	app.SlashingKeeper = slashingkeeper.NewKeeper(
		appCodec, keys[slashingtypes.StoreKey], app.StakingKeeper, app.GetSubspace(slashingtypes.ModuleName),
	)
	app.CrisisKeeper = crisiskeeper.NewKeeper(
		app.GetSubspace(crisistypes.ModuleName), invCheckPeriod, app.BankKeeper, authtypes.FeeCollectorName,
	)

	app.FeeGrantKeeper = feegrantkeeper.NewKeeper(appCodec, keys[feegrant.StoreKey], app.AccountKeeper)

	app.AuthzKeeper = authzkeeper.NewKeeper(
		keys[authzkeeper.StoreKey],
		appCodec,
		app.MsgServiceRouter(),
		app.AccountKeeper,
	)

	groupConfig := group.DefaultConfig()
	groupConfig.MaxMetadataLen = 600
	app.GroupKeeper = groupkeeper.NewKeeper(
		keys[group.StoreKey],
		appCodec,
		app.MsgServiceRouter(),
		app.AccountKeeper,
		groupConfig,
	)

	// set the governance module account as the authority for conducting upgrades
	app.UpgradeKeeper = upgradekeeper.NewKeeper(
		skipUpgradeHeights,
		keys[upgradetypes.StoreKey],
		appCodec,
		homePath,
		app.BaseApp,
		govModuleAddr,
	)

	app.NFTKeeper = nftkeeper.NewKeeper(keys[nftkeeper.StoreKey], appCodec, app.AccountKeeper, app.BankKeeper)

	app.OracleKeeper = oraclekeeper.NewKeeper(
		appCodec,
		keys[oracletypes.ModuleName],
		app.GetSubspace(oracletypes.ModuleName),
		app.AccountKeeper,
		app.BankKeeper,
		app.DistrKeeper,
		app.StakingKeeper,
		distrtypes.ModuleName,
	)
	app.LeverageKeeper = leveragekeeper.NewKeeper(
		appCodec,
		keys[leveragetypes.ModuleName],
		app.GetSubspace(leveragetypes.ModuleName),
		app.BankKeeper,
		app.OracleKeeper,
		cast.ToBool(appOpts.Get(leveragetypes.FlagEnableLiquidatorQuery)),
	)
	app.LeverageKeeper = *app.LeverageKeeper.SetHooks(
		leveragetypes.NewMultiHooks(
			app.OracleKeeper.Hooks(),
		),
	)

	// register the staking hooks
	// NOTE: stakingKeeper above is passed by reference, so that it will contain these hooks
	app.StakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(
			app.DistrKeeper.Hooks(),
			app.SlashingKeeper.Hooks(),
		),
	)

	// Create evidence Keeper so we can register the IBC light client misbehavior
	// evidence route.
	evidenceKeeper := evidencekeeper.NewKeeper(
		appCodec,
		keys[evidencetypes.StoreKey],
		app.StakingKeeper,
		app.SlashingKeeper,
	)
	// If evidence needs to be handled for the app, set routes in router here and seal
	app.EvidenceKeeper = *evidenceKeeper

	app.IBCKeeper = ibckeeper.NewKeeper(
		appCodec,
		keys[ibchost.StoreKey],
		app.GetSubspace(ibchost.ModuleName),
		*app.StakingKeeper,
		app.UpgradeKeeper,
		app.ScopedIBCKeeper,
	)

	var ics4Wrapper ibcporttypes.ICS4Wrapper
	app.UIbcQuotaKeeper = uibcquotakeeper.NewKeeper(
		appCodec,
		keys[uibc.StoreKey],
		app.IBCKeeper.ChannelKeeper, app.LeverageKeeper, app.OracleKeeper,
	)
	ics4Wrapper = app.UIbcQuotaKeeper

	// Middleware Stacks
	// Create an original ICS-20 transfer keeper and AppModule and then use it to
	// created an Umee wrapped ICS-20 transfer keeper and AppModule.
	ibcTransferKeeper := ibctransferkeeper.NewKeeper(
		appCodec,
		keys[ibctransfertypes.StoreKey],
		app.GetSubspace(ibctransfertypes.ModuleName),
		ics4Wrapper, // ISC4 Wrapper: IBC Rate Limit middleware
		app.IBCKeeper.ChannelKeeper,
		&app.IBCKeeper.PortKeeper,
		app.AccountKeeper,
		app.BankKeeper,
		app.ScopedTransferKeeper,
	)

	app.UIBCTransferKeeper = uibctransferkeeper.New(ibcTransferKeeper, app.BankKeeper)

	// Create Transfer Stack
	// SendPacket, since it is originating from the application to core IBC:
	// transferKeeper.SendPacket -> uibcquota.SendPacket -> channel.SendPacket

	// RecvPacket, message that originates from core IBC and goes down to app, the flow is the other way
	// channel.RecvPacket -> uibcquota.OnRecvPacket -> transfer.OnRecvPacket

	// transfer stack contains (from top to bottom):
	// - Umee IBC Transfer
	// - IBC Rate Limit Middleware

	// create IBC module from bottom to top of stack
	var transferStack ibcporttypes.IBCModule
	transferStack = uics20transfer.NewIBCModule(
		ibctransfer.NewIBCModule(ibcTransferKeeper),
		app.UIBCTransferKeeper,
	)

	transferStack = uibcquota.NewIBCMiddleware(transferStack, app.UIbcQuotaKeeper, appCodec)

	// Create IBC Router
	// create static IBC router, add transfer route, then set and seal it
	ibcRouter := ibcporttypes.NewRouter()
	// Add transfer stack to IBC Router
	ibcRouter.AddRoute(ibctransfertypes.ModuleName, transferStack)
	ibcRouter.AddRoute(wasm.ModuleName, wasm.NewIBCHandler(app.WasmKeeper, app.IBCKeeper.ChannelKeeper))
	app.IBCKeeper.SetRouter(ibcRouter)

	app.bech32IbcKeeper = *bech32ibckeeper.NewKeeper(
		app.IBCKeeper.ChannelKeeper, appCodec, keys[bech32ibctypes.StoreKey],
		app.UIBCTransferKeeper,
	)

	// Register the proposal types
	// Deprecated: Avoid adding new handlers, instead use the new proposal flow
	// by granting the governance module the right to execute the message.
	// See: https://github.com/cosmos/cosmos-sdk/blob/release/v0.46.x/x/gov/spec/01_concepts.md#proposal-messages
	govRouter := govv1beta1.NewRouter()
	govRouter.
		AddRoute(govtypes.RouterKey, govv1beta1.ProposalHandler).
		AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(app.ParamsKeeper)).
		AddRoute(distrtypes.RouterKey, distr.NewCommunityPoolSpendProposalHandler(app.DistrKeeper)).
		AddRoute(upgradetypes.RouterKey, upgrade.NewSoftwareUpgradeProposalHandler(app.UpgradeKeeper)).
		AddRoute(ibcclienttypes.RouterKey, ibcclient.NewClientProposalHandler(app.IBCKeeper.ClientKeeper)).
		AddRoute(bech32ibctypes.RouterKey, bech32ibc.NewBech32IBCProposalHandler(app.bech32IbcKeeper))

	// The wasm gov proposal types can be individually enabled
	if Experimental && len(wasmEnabledProposals) != 0 {
		govRouter.AddRoute(wasm.RouterKey, wasm.NewWasmProposalHandler(app.WasmKeeper, wasmEnabledProposals))
	}

	govConfig := govtypes.DefaultConfig()
	govConfig.MaxMetadataLen = 800
	govKeeper := govkeeper.NewKeeper(
		appCodec, keys[govtypes.StoreKey], app.GetSubspace(govtypes.ModuleName), app.AccountKeeper, app.BankKeeper,
		app.StakingKeeper, govRouter, app.MsgServiceRouter(), govConfig,
	)
	app.GovKeeper = *govKeeper.SetHooks(
		govtypes.NewMultiGovHooks(
		// register the governance hooks
		),
	)

	app.customKeepers(bApp, keys, appCodec, govRouter, homePath, appOpts, wasmOpts)

	/****  Module Options ****/

	// NOTE: we may consider parsing `appOpts` inside module constructors. For the moment
	// we prefer to be more strict in what arguments the modules expect.
	skipGenesisInvariants := cast.ToBool(appOpts.Get(crisis.FlagSkipGenesisInvariants))

	// NOTE: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.
	appModules := []module.AppModule{
		genutil.NewAppModule(
			app.AccountKeeper, app.StakingKeeper, app.BaseApp.DeliverTx,
			encodingConfig.TxConfig,
		),
		auth.NewAppModule(appCodec, app.AccountKeeper, nil),
		vesting.NewAppModule(app.AccountKeeper, app.BankKeeper),
		bank.NewAppModule(appCodec, app.BankKeeper, app.AccountKeeper),
		capability.NewAppModule(appCodec, *app.CapabilityKeeper),
		crisis.NewAppModule(&app.CrisisKeeper, skipGenesisInvariants),
		feegrantmodule.NewAppModule(appCodec, app.AccountKeeper, app.BankKeeper, app.FeeGrantKeeper, app.interfaceRegistry),
		gov.NewAppModule(appCodec, app.GovKeeper, app.AccountKeeper, app.BankKeeper),
		mint.NewAppModule(appCodec, app.MintKeeper, app.AccountKeeper, nil),
		// need to dereference StakingKeeper because x/distribution uses interface casting :(
		// TODO: in the next SDK version we can remove the dereference
		slashing.NewAppModule(appCodec, app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, *app.StakingKeeper),
		distr.NewAppModule(appCodec, app.DistrKeeper, app.AccountKeeper, app.BankKeeper, *app.StakingKeeper),
		staking.NewAppModule(appCodec, *app.StakingKeeper, app.AccountKeeper, app.BankKeeper),
		upgrade.NewAppModule(app.UpgradeKeeper),
		evidence.NewAppModule(app.EvidenceKeeper),
		ibc.NewAppModule(app.IBCKeeper),
		params.NewAppModule(app.ParamsKeeper),
		authzmodule.NewAppModule(appCodec, app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		groupmodule.NewAppModule(appCodec, app.GroupKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		nftmodule.NewAppModule(appCodec, app.NFTKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),

		ibctransfer.NewAppModule(ibcTransferKeeper),
		leverage.NewAppModule(appCodec, app.LeverageKeeper, app.AccountKeeper, app.BankKeeper),
		oracle.NewAppModule(appCodec, app.OracleKeeper, app.AccountKeeper, app.BankKeeper),
		bech32ibc.NewAppModule(appCodec, app.bech32IbcKeeper),
		uibcmodule.NewAppModule(appCodec, app.UIbcQuotaKeeper),
	}
	if Experimental {
		appModules = append(
			appModules,
			wasm.NewAppModule(app.appCodec, &app.WasmKeeper, app.StakingKeeper, app.AccountKeeper, app.BankKeeper),
		)
	}

	app.mm = module.NewManager(appModules...)

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	// NOTE: staking module is required if HistoricalEntries param > 0
	// NOTE: capability module's beginblocker must come before any modules using capabilities (e.g. IBC)
	beginBlockers := []string{
		upgradetypes.ModuleName, capabilitytypes.ModuleName, minttypes.ModuleName, distrtypes.ModuleName,
		slashingtypes.ModuleName, evidencetypes.ModuleName, stakingtypes.ModuleName,
		ibchost.ModuleName, ibctransfertypes.ModuleName,
		authtypes.ModuleName, banktypes.ModuleName, govtypes.ModuleName, crisistypes.ModuleName, genutiltypes.ModuleName,
		authz.ModuleName, feegrant.ModuleName,
		nft.ModuleName,
		group.ModuleName,
		paramstypes.ModuleName, vestingtypes.ModuleName,
		// icatypes.ModuleName,  ibcfeetypes.ModuleName,
		leveragetypes.ModuleName,
		oracletypes.ModuleName,
		bech32ibctypes.ModuleName,
		uibc.ModuleName,
	}

	endBlockers := []string{
		crisistypes.ModuleName,
		oracletypes.ModuleName, // must be before gov and staking
		govtypes.ModuleName, stakingtypes.ModuleName,
		ibchost.ModuleName, ibctransfertypes.ModuleName,
		capabilitytypes.ModuleName, authtypes.ModuleName, banktypes.ModuleName, distrtypes.ModuleName,
		slashingtypes.ModuleName, minttypes.ModuleName,
		genutiltypes.ModuleName, evidencetypes.ModuleName, authz.ModuleName,
		feegrant.ModuleName, nft.ModuleName, group.ModuleName,
		paramstypes.ModuleName, upgradetypes.ModuleName, vestingtypes.ModuleName,
		// icatypes.ModuleName,
		leveragetypes.ModuleName,
		bech32ibctypes.ModuleName,
		uibc.ModuleName,
	}

	// NOTE: The genutils module must occur after staking so that pools are
	// properly initialized with tokens from genesis accounts.
	// NOTE: The genutils module must also occur after auth so that it can access the params from auth.
	// NOTE: Capability module must occur first so that it can initialize any capabilities
	// so that other modules that want to create or claim capabilities afterwards in InitChain
	// can do so safely.
	initGenesis := []string{
		capabilitytypes.ModuleName, authtypes.ModuleName, banktypes.ModuleName, distrtypes.ModuleName,
		stakingtypes.ModuleName, slashingtypes.ModuleName, govtypes.ModuleName, minttypes.ModuleName,
		crisistypes.ModuleName, ibchost.ModuleName, genutiltypes.ModuleName, evidencetypes.ModuleName,
		authz.ModuleName, ibctransfertypes.ModuleName, // icatypes.ModuleName,
		feegrant.ModuleName, nft.ModuleName, group.ModuleName,
		paramstypes.ModuleName, upgradetypes.ModuleName, vestingtypes.ModuleName,

		oracletypes.ModuleName,
		leveragetypes.ModuleName,
		bech32ibctypes.ModuleName,
		uibc.ModuleName,
	}

	orderMigrations := []string{
		capabilitytypes.ModuleName, authtypes.ModuleName, banktypes.ModuleName, distrtypes.ModuleName,
		stakingtypes.ModuleName, slashingtypes.ModuleName, govtypes.ModuleName, minttypes.ModuleName,
		crisistypes.ModuleName, ibchost.ModuleName, genutiltypes.ModuleName, evidencetypes.ModuleName,
		authz.ModuleName, ibctransfertypes.ModuleName, // icatypes.ModuleName,
		feegrant.ModuleName, nft.ModuleName, group.ModuleName,
		paramstypes.ModuleName, upgradetypes.ModuleName, vestingtypes.ModuleName,

		oracletypes.ModuleName,
		leveragetypes.ModuleName,
		bech32ibctypes.ModuleName,
		uibc.ModuleName,
	}

	if Experimental {
		beginBlockers = append(beginBlockers, wasm.ModuleName)
		endBlockers = append(endBlockers, wasm.ModuleName)
		initGenesis = append(initGenesis, wasm.ModuleName)
		orderMigrations = append(orderMigrations, wasm.ModuleName)
	}

	app.mm.SetOrderBeginBlockers(beginBlockers...)
	app.mm.SetOrderEndBlockers(endBlockers...)
	app.mm.SetOrderInitGenesis(initGenesis...)
	app.mm.SetOrderMigrations(orderMigrations...)

	app.mm.RegisterInvariants(&app.CrisisKeeper)
	app.mm.RegisterRoutes(app.Router(), app.QueryRouter(), encodingConfig.Amino)
	app.configurator = module.NewConfigurator(app.appCodec, app.MsgServiceRouter(), app.GRPCQueryRouter())
	app.mm.RegisterServices(app.configurator)

	// RegisterUpgradeHandlers is used for registering any on-chain upgrades.
	// Make sure it's called after `app.mm` and `app.configurator` are set.
	app.RegisterUpgradeHandlers(Experimental)

	// add test gRPC service for testing gRPC queries in isolation
	testdata.RegisterQueryServer(app.GRPCQueryRouter(), testdata.QueryImpl{})

	// create the simulation manager and define the order of the modules for deterministic simulations
	overrideModules := map[string]module.AppModuleSimulation{
		authtypes.ModuleName: auth.NewAppModule(app.appCodec, app.AccountKeeper, authsims.RandomGenesisAccounts),
	}

	simStateModules := genmap.Pick(
		app.mm.Modules,
		[]string{stakingtypes.ModuleName, authtypes.ModuleName, oracletypes.ModuleName},
	)
	// TODO: Ensure x/leverage implements simulator and add it here:
	simTestModules := genmap.Pick(simStateModules, []string{oracletypes.ModuleName})

	app.StateSimulationManager = module.NewSimulationManagerFromAppModules(simStateModules, overrideModules)
	app.sm = module.NewSimulationManagerFromAppModules(simTestModules, nil)

	app.sm.RegisterStoreDecoders()

	// initialize stores
	app.MountKVStores(keys)
	app.MountTransientStores(tkeys)
	app.MountMemoryStores(memKeys)

	// initialize BaseApp
	app.SetInitChainer(app.InitChainer)
	app.SetBeginBlocker(app.BeginBlocker)
	app.SetEndBlocker(app.EndBlocker)
	app.setAnteHandler(encodingConfig.TxConfig, &app.wasmCfg, keys[wasm.StoreKey])
	// In v0.46, the SDK introduces _postHandlers_. PostHandlers are like
	// antehandlers, but are run _after_ the `runMsgs` execution. They are also
	// defined as a chain, and have the same signature as antehandlers.
	//
	// In baseapp, postHandlers are run in the same store branch as `runMsgs`,
	// meaning that both `runMsgs` and `postHandler` state will be committed if
	// both are successful, and both will be reverted if any of the two fails.
	//
	// The SDK exposes a default empty postHandlers chain.
	//
	// Please note that changing any of the anteHandler or postHandler chain is
	// likely to be a state-machine breaking change, which needs a coordinated
	// upgrade.
	app.setPostHandler()

	app.registerCustomExtensions()

	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			tmos.Exit(fmt.Sprintf("failed to load latest version: %s", err))
		}
	}

	return app
}

func (app *UmeeApp) setAnteHandler(txConfig client.TxConfig,
	wasmConfig *wasmtypes.WasmConfig, wasmStoreKey *storetypes.KVStoreKey,
) {
	anteHandler, err := customante.NewAnteHandler(
		customante.HandlerOptions{
			AccountKeeper:     app.AccountKeeper,
			BankKeeper:        app.BankKeeper,
			OracleKeeper:      app.OracleKeeper,
			IBCKeeper:         app.IBCKeeper,
			SignModeHandler:   txConfig.SignModeHandler(),
			FeegrantKeeper:    app.FeeGrantKeeper,
			SigGasConsumer:    ante.DefaultSigVerificationGasConsumer,
			WasmConfig:        wasmConfig,
			TXCounterStoreKey: wasmStoreKey,
		},
		Experimental,
	)
	if err != nil {
		panic(err)
	}

	app.SetAnteHandler(anteHandler)
}

func (app *UmeeApp) setPostHandler() {
	postHandler, err := posthandler.NewPostHandler(
		posthandler.HandlerOptions{},
	)
	if err != nil {
		panic(err)
	}

	app.SetPostHandler(postHandler)
}

// Name returns the name of the App
func (app *UmeeApp) Name() string { return app.BaseApp.Name() }

// BeginBlocker implements Umee's BeginBlock ABCI method.
func (app *UmeeApp) BeginBlocker(ctx sdk.Context, req abci.RequestBeginBlock) abci.ResponseBeginBlock {
	return app.mm.BeginBlock(ctx, req)
}

// EndBlocker implements Umee's EndBlock ABCI method.
func (app *UmeeApp) EndBlocker(ctx sdk.Context, req abci.RequestEndBlock) abci.ResponseEndBlock {
	return app.mm.EndBlock(ctx, req)
}

// InitChainer implements Umee's InitChain ABCI method.
func (app *UmeeApp) InitChainer(ctx sdk.Context, req abci.RequestInitChain) abci.ResponseInitChain {
	var genesisState GenesisState
	if err := json.Unmarshal(req.AppStateBytes, &genesisState); err != nil {
		panic(fmt.Sprintf("failed to unmarshal genesis state: %v", err))
	}
	app.UpgradeKeeper.SetModuleVersionMap(ctx, app.mm.GetVersionMap())
	return app.mm.InitGenesis(ctx, app.appCodec, genesisState)
}

// LoadHeight loads a particular height via Umee's BaseApp.
func (app *UmeeApp) LoadHeight(height int64) error {
	return app.LoadVersion(height)
}

// ModuleAccountAddrs returns all of Umee's module account addresses.
func (app *UmeeApp) ModuleAccountAddrs() map[string]bool {
	modAccAddrs := make(map[string]bool)
	for acc := range maccPerms {
		modAccAddrs[authtypes.NewModuleAddress(acc).String()] = true
	}

	return modAccAddrs
}

// LegacyAmino returns Umee's amino codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *UmeeApp) LegacyAmino() *codec.LegacyAmino {
	return app.legacyAmino
}

// AppCodec returns Umee's app codec.
//
// NOTE: This is solely to be used for testing purposes as it may be desirable
// for modules to register their own custom testing types.
func (app *UmeeApp) AppCodec() codec.Codec {
	return app.appCodec
}

// InterfaceRegistry returns Umee's InterfaceRegistry.
func (app *UmeeApp) InterfaceRegistry() types.InterfaceRegistry {
	return app.interfaceRegistry
}

// GetKey returns the KVStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *UmeeApp) GetKey(storeKey string) *storetypes.KVStoreKey {
	return app.keys[storeKey]
}

// GetTKey returns the TransientStoreKey for the provided store key.
//
// NOTE: This is solely to be used for testing purposes.
func (app *UmeeApp) GetTKey(storeKey string) *storetypes.TransientStoreKey {
	return app.tkeys[storeKey]
}

// GetMemKey returns the MemStoreKey for the provided mem key.
//
// NOTE: This is solely used for testing purposes.
func (app *UmeeApp) GetMemKey(storeKey string) *storetypes.MemoryStoreKey {
	return app.memKeys[storeKey]
}

// GetSubspace returns a param subspace for a given module name.
//
// NOTE: This is solely to be used for testing purposes.
func (app *UmeeApp) GetSubspace(moduleName string) paramstypes.Subspace {
	subspace, _ := app.ParamsKeeper.GetSubspace(moduleName)
	return subspace
}

// SimulationManager returns the application's SimulationManager.
func (app *UmeeApp) SimulationManager() *module.SimulationManager {
	return app.sm
}

// RegisterAPIRoutes registers all application module routes with the provided
//
// API server.
func (app *UmeeApp) RegisterAPIRoutes(apiSvr *api.Server, apiConfig config.APIConfig) {
	clientCtx := apiSvr.ClientCtx
	// Register new tx routes from grpc-gateway.
	authtx.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)
	// Register new tendermint queries routes from grpc-gateway.
	tmservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register node gRPC service for grpc-gateway.
	nodeservice.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// Register grpc-gateway routes for all modules.
	ModuleBasics.RegisterGRPCGatewayRoutes(clientCtx, apiSvr.GRPCGatewayRouter)

	// register swagger API from root so that other applications can override easily
	if apiConfig.Swagger {
		swagger.RegisterSwaggerAPI(apiSvr.Router)
	}
}

// RegisterTxService implements the Application.RegisterTxService method.
func (app *UmeeApp) RegisterTxService(clientCtx client.Context) {
	authtx.RegisterTxService(app.BaseApp.GRPCQueryRouter(), clientCtx, app.BaseApp.Simulate, app.interfaceRegistry)
}

// RegisterTendermintService implements the Application.RegisterTendermintService method.
func (app *UmeeApp) RegisterTendermintService(clientCtx client.Context) {
	tmservice.RegisterTendermintService(
		clientCtx,
		app.BaseApp.GRPCQueryRouter(),
		app.interfaceRegistry,
		app.Query,
	)
}

func (app *UmeeApp) RegisterNodeService(clientCtx client.Context) {
	nodeservice.RegisterNodeService(clientCtx, app.GRPCQueryRouter())
}

// GetBaseApp is used solely for testing purposes.
func (app *UmeeApp) GetBaseApp() *baseapp.BaseApp {
	return app.BaseApp
}

// GetStakingKeeper is used solely for testing purposes.
func (app *UmeeApp) GetStakingKeeper() ibctesting.StakingKeeper {
	return *app.StakingKeeper
}

// GetIBCKeeper is used solely for testing purposes.
func (app *UmeeApp) GetIBCKeeper() *ibckeeper.Keeper {
	return app.IBCKeeper
}

// GetScopedIBCKeeper is used solely for testing purposes.
func (app *UmeeApp) GetScopedIBCKeeper() capabilitykeeper.ScopedKeeper {
	return app.ScopedIBCKeeper
}

// GetTxConfig is used solely for testing purposes.
func (app *UmeeApp) GetTxConfig() client.TxConfig {
	return app.txConfig
}

// GetMaccPerms returns a deep copy of the module account permissions.
func GetMaccPerms() map[string][]string {
	dupMaccPerms := make(map[string][]string)
	for k, v := range maccPerms {
		dupMaccPerms[k] = v
	}
	return dupMaccPerms
}

// initParamsKeeper init params keeper and its subspaces
func initParamsKeeper(
	appCodec codec.BinaryCodec,
	legacyAmino *codec.LegacyAmino,
	key,
	tkey storetypes.StoreKey,
) paramskeeper.Keeper {
	paramsKeeper := paramskeeper.NewKeeper(appCodec, legacyAmino, key, tkey)

	paramsKeeper.Subspace(authtypes.ModuleName)
	paramsKeeper.Subspace(banktypes.ModuleName)
	paramsKeeper.Subspace(stakingtypes.ModuleName)
	paramsKeeper.Subspace(minttypes.ModuleName)
	paramsKeeper.Subspace(distrtypes.ModuleName)
	paramsKeeper.Subspace(slashingtypes.ModuleName)
	paramsKeeper.Subspace(govtypes.ModuleName).WithKeyTable(govv1.ParamKeyTable())
	paramsKeeper.Subspace(crisistypes.ModuleName)
	paramsKeeper.Subspace(ibctransfertypes.ModuleName)
	paramsKeeper.Subspace(ibchost.ModuleName)
	// paramsKeeper.Subspace(icacontrollertypes.SubModuleName)
	// paramsKeeper.Subspace(icahosttypes.SubModuleName)
	paramsKeeper.Subspace(leveragetypes.ModuleName)
	paramsKeeper.Subspace(oracletypes.ModuleName)
	if Experimental {
		paramsKeeper.Subspace(wasm.ModuleName)
	}

	return paramsKeeper
}

func getGovProposalHandlers() []govclient.ProposalHandler {
	handlers := []govclient.ProposalHandler{
		paramsclient.ProposalHandler,
		distrclient.ProposalHandler,
		upgradeclient.LegacyProposalHandler,
		upgradeclient.LegacyCancelProposalHandler,
		ibcclientclient.UpdateClientProposalHandler,
		ibcclientclient.UpgradeProposalHandler,
	}
	if Experimental {
		handlers = append(handlers, wasmclient.ProposalHandlers...)
	}

	return handlers
}
