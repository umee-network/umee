package app

import (
	"encoding/json"
	"fmt"
	"io"
	"os"
	"path/filepath"
	"strings"

	"github.com/spf13/cast"

	dbm "github.com/cometbft/cometbft-db"
	abci "github.com/cometbft/cometbft/abci/types"
	"github.com/cometbft/cometbft/libs/log"
	tmos "github.com/cometbft/cometbft/libs/os"
	tmproto "github.com/cometbft/cometbft/proto/tendermint/types"

	autocliv1 "cosmossdk.io/api/cosmos/autocli/v1"
	reflectionv1 "cosmossdk.io/api/cosmos/reflection/v1"
	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/client"
	nodeservice "github.com/cosmos/cosmos-sdk/client/grpc/node"
	"github.com/cosmos/cosmos-sdk/client/grpc/tmservice"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/codec/types"
	"github.com/cosmos/cosmos-sdk/runtime"
	runtimeservices "github.com/cosmos/cosmos-sdk/runtime/services"
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
	consensusparamskeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	consensusparamstypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
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

	packetforward "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward"
	packetforwardkeeper "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward/keeper"
	packetforwardtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward/types"
	ica "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts"
	icahost "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host"
	icahostkeeper "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/keeper"
	icahosttypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/types"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	ibctransfer "github.com/cosmos/ibc-go/v7/modules/apps/transfer"
	ibctransferkeeper "github.com/cosmos/ibc-go/v7/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v7/modules/core"
	ibcclient "github.com/cosmos/ibc-go/v7/modules/core/02-client"
	ibcclientclient "github.com/cosmos/ibc-go/v7/modules/core/02-client/client"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibcporttypes "github.com/cosmos/ibc-go/v7/modules/core/05-port/types"
	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"

	// cosmwasm
	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"

	customante "github.com/umee-network/umee/v6/ante"
	"github.com/umee-network/umee/v6/app/inflation"
	appparams "github.com/umee-network/umee/v6/app/params"
	uwasm "github.com/umee-network/umee/v6/app/wasm"
	"github.com/umee-network/umee/v6/swagger"
	"github.com/umee-network/umee/v6/util/genmap"
	"github.com/umee-network/umee/v6/x/auction"
	auctionkeeper "github.com/umee-network/umee/v6/x/auction/keeper"
	auctionmodule "github.com/umee-network/umee/v6/x/auction/module"
	"github.com/umee-network/umee/v6/x/incentive"
	incentivekeeper "github.com/umee-network/umee/v6/x/incentive/keeper"
	incentivemodule "github.com/umee-network/umee/v6/x/incentive/module"
	"github.com/umee-network/umee/v6/x/leverage"
	leveragekeeper "github.com/umee-network/umee/v6/x/leverage/keeper"
	leveragetypes "github.com/umee-network/umee/v6/x/leverage/types"
	"github.com/umee-network/umee/v6/x/metoken"
	metokenkeeper "github.com/umee-network/umee/v6/x/metoken/keeper"
	metokenmodule "github.com/umee-network/umee/v6/x/metoken/module"
	"github.com/umee-network/umee/v6/x/oracle"
	oraclekeeper "github.com/umee-network/umee/v6/x/oracle/keeper"
	oracletypes "github.com/umee-network/umee/v6/x/oracle/types"
	"github.com/umee-network/umee/v6/x/ugov"
	ugovkeeper "github.com/umee-network/umee/v6/x/ugov/keeper"
	ugovmodule "github.com/umee-network/umee/v6/x/ugov/module"
	"github.com/umee-network/umee/v6/x/uibc"
	uibcmodule "github.com/umee-network/umee/v6/x/uibc/module"
	uibcoracle "github.com/umee-network/umee/v6/x/uibc/oracle"
	uibcquota "github.com/umee-network/umee/v6/x/uibc/quota"
	"github.com/umee-network/umee/v6/x/uibc/uics20"

	feeabsmodule "github.com/osmosis-labs/fee-abstraction/v7/x/feeabs"
	feeabskeeper "github.com/osmosis-labs/fee-abstraction/v7/x/feeabs/keeper"
	feeabstypes "github.com/osmosis-labs/fee-abstraction/v7/x/feeabs/types"
)

var (
	_ runtime.AppI            = (*UmeeApp)(nil)
	_ servertypes.Application = (*UmeeApp)(nil)

	// DefaultNodeHome defines the default home directory for the application daemon.
	DefaultNodeHome string

	// ModuleBasics defines the module BasicManager is in charge of setting up basic,
	// non-dependant module elements, such as codec registration
	// and genesis verification.
	ModuleBasics module.BasicManager

	// module account permissions
	maccPerms map[string][]string
)

func init() {
	userHomeDir, err := os.UserHomeDir()
	if err != nil {
		panic(fmt.Sprintf("failed to get user home directory: %s", err))
	}
	DefaultNodeHome = filepath.Join(userHomeDir, "."+appparams.Name)

	moduleBasics := []module.AppModuleBasic{
		auth.AppModuleBasic{},
		genutil.NewAppModuleBasic(genutiltypes.DefaultMessageValidator),
		BankModule{},
		capability.AppModuleBasic{},
		StakingModule{},
		MintModule{},
		distr.AppModuleBasic{},
		gov.NewAppModuleBasic(getGovProposalHandlers()),
		params.AppModuleBasic{},
		SlashingModule{},
		feegrantmodule.AppModuleBasic{},
		upgrade.AppModuleBasic{},
		evidence.AppModuleBasic{},
		authzmodule.AppModuleBasic{},
		groupmodule.AppModuleBasic{},
		vesting.AppModuleBasic{},
		nftmodule.AppModuleBasic{},
		ibc.AppModuleBasic{},
		ibctm.AppModuleBasic{},
		ibctransfer.AppModuleBasic{},
		ica.AppModuleBasic{},
		// intertx.AppModuleBasic{},
		// ibcfee.AppModuleBasic{},

		leverage.AppModuleBasic{},
		oracle.AppModuleBasic{},
		uibcmodule.AppModuleBasic{},
		ugovmodule.AppModuleBasic{},
		wasm.AppModuleBasic{},
		auctionmodule.AppModuleBasic{},
		incentivemodule.AppModuleBasic{},
		metokenmodule.AppModuleBasic{},
		packetforward.AppModuleBasic{},
		feeabsmodule.AppModuleBasic{},
	}
	// if Experimental {}

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
		icatypes.ModuleName:         nil,

		leveragetypes.ModuleName: {authtypes.Minter, authtypes.Burner},
		wasmtypes.ModuleName:     {authtypes.Burner},

		auction.ModuleName:     {authtypes.Burner},
		incentive.ModuleName:   nil,
		metoken.ModuleName:     {authtypes.Minter, authtypes.Burner},
		oracletypes.ModuleName: nil,
		ugov.ModuleName:        nil,
		uibc.ModuleName:        nil,
		feeabstypes.ModuleName: nil,
	}
	// if Experimental {}
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
	AccountKeeper         authkeeper.AccountKeeper
	BankKeeper            bankkeeper.BaseKeeper
	CapabilityKeeper      *capabilitykeeper.Keeper
	ConsensusParamsKeeper consensusparamskeeper.Keeper
	StakingKeeper         *stakingkeeper.Keeper
	SlashingKeeper        slashingkeeper.Keeper
	MintKeeper            mintkeeper.Keeper
	DistrKeeper           distrkeeper.Keeper
	GovKeeper             *govkeeper.Keeper
	UpgradeKeeper         *upgradekeeper.Keeper
	ParamsKeeper          paramskeeper.Keeper
	AuthzKeeper           authzkeeper.Keeper
	EvidenceKeeper        evidencekeeper.Keeper
	FeeGrantKeeper        feegrantkeeper.Keeper
	GroupKeeper           groupkeeper.Keeper
	NFTKeeper             nftkeeper.Keeper
	WasmKeeper            wasmkeeper.Keeper

	IBCTransferKeeper   ibctransferkeeper.Keeper
	IBCKeeper           *ibckeeper.Keeper // IBC Keeper must be a pointer in the app, so we can SetRouter on it correctly
	PacketForwardKeeper *packetforwardkeeper.Keeper
	ICAHostKeeper       icahostkeeper.Keeper
	LeverageKeeper      leveragekeeper.Keeper
	FeeabsKeeper        feeabskeeper.Keeper

	AuctionKeeperB   auctionkeeper.Builder
	IncentiveKeeper  incentivekeeper.Keeper
	MetokenKeeperB   metokenkeeper.Builder
	OracleKeeper     oraclekeeper.Keeper
	UGovKeeperB      ugovkeeper.Builder
	UIbcQuotaKeeperB uibcquota.Builder

	// make scoped keepers public for testing purposes
	ScopedIBCKeeper      capabilitykeeper.ScopedKeeper
	ScopedTransferKeeper capabilitykeeper.ScopedKeeper
	ScopedWasmKeeper     capabilitykeeper.ScopedKeeper
	ScopedFeeabsKeeper   capabilitykeeper.ScopedKeeper

	// the module manager
	mm *module.Manager

	// simulation manager
	sm *module.SimulationManager
	// simulation manager to create state
	StateSimulationManager *module.SimulationManager

	// module configurator
	configurator module.Configurator

	// wasm
	wasmCfg wasmtypes.WasmConfig
}

func New(
	logger log.Logger,
	db dbm.DB,
	traceStore io.Writer,
	loadLatest bool,
	skipUpgradeHeights map[int64]bool,
	homePath string,
	invCheckPeriod uint,
	appOpts servertypes.AppOptions,
	wasmOpts []wasmkeeper.Option,
	baseAppOptions ...func(*baseapp.BaseApp),
) *UmeeApp {
	encCfg := MakeEncodingConfig()
	interfaceRegistry := encCfg.InterfaceRegistry
	appCodec := encCfg.Codec
	legacyAmino := encCfg.Amino
	txConfig := encCfg.TxConfig

	bApp := baseapp.NewBaseApp(appparams.Name, logger, db, txConfig.TxDecoder(), baseAppOptions...)
	bApp.SetCommitMultiStoreTracer(traceStore)
	bApp.SetVersion(version.Version)
	bApp.SetInterfaceRegistry(interfaceRegistry)

	storeKeys := []string{
		authtypes.StoreKey, banktypes.StoreKey, stakingtypes.StoreKey,
		minttypes.StoreKey, distrtypes.StoreKey, slashingtypes.StoreKey,
		govtypes.StoreKey, paramstypes.StoreKey, upgradetypes.StoreKey, feegrant.StoreKey,
		evidencetypes.StoreKey, capabilitytypes.StoreKey,
		authzkeeper.StoreKey, nftkeeper.StoreKey, group.StoreKey,
		ibcexported.StoreKey, ibctransfertypes.StoreKey, icahosttypes.StoreKey,
		packetforwardtypes.StoreKey,
		wasmtypes.StoreKey,
		consensusparamstypes.StoreKey,

		auction.StoreKey,
		incentive.StoreKey,
		leveragetypes.StoreKey,
		metoken.StoreKey,
		oracletypes.StoreKey,
		ugov.StoreKey,
		uibc.StoreKey,
		feeabstypes.StoreKey,
	}
	// if Experimental {}

	keys := sdk.NewKVStoreKeys(storeKeys...)
	tkeys := sdk.NewTransientStoreKeys(paramstypes.TStoreKey)
	memKeys := sdk.NewMemoryStoreKeys(capabilitytypes.MemStoreKey)
	govModuleAddr := authtypes.NewModuleAddress(govtypes.ModuleName).String()

	app := &UmeeApp{
		BaseApp:           bApp,
		legacyAmino:       legacyAmino,
		appCodec:          appCodec,
		interfaceRegistry: interfaceRegistry,
		txConfig:          txConfig,
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

	app.ConsensusParamsKeeper = consensusparamskeeper.NewKeeper(
		appCodec,
		keys[consensusparamstypes.StoreKey],
		govModuleAddr,
	)
	// set the BaseApp's parameter store
	bApp.SetParamStore(&app.ConsensusParamsKeeper)

	app.CapabilityKeeper = capabilitykeeper.NewKeeper(
		appCodec,
		keys[capabilitytypes.StoreKey],
		memKeys[capabilitytypes.MemStoreKey],
	)

	// grant capabilities for the ibc and ibc-transfer modules
	app.ScopedIBCKeeper = app.CapabilityKeeper.ScopeToModule(ibcexported.ModuleName)
	app.ScopedTransferKeeper = app.CapabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)
	scopedICAHostKeeper := app.CapabilityKeeper.ScopeToModule(icahosttypes.SubModuleName)
	app.ScopedWasmKeeper = app.CapabilityKeeper.ScopeToModule(wasmtypes.ModuleName)
	app.ScopedFeeabsKeeper = app.CapabilityKeeper.ScopeToModule(feeabstypes.ModuleName)

	// Applications that wish to enforce statically created ScopedKeepers should call `Seal` after creating
	// their scoped modules in `NewApp` with `ScopeToModule`
	app.CapabilityKeeper.Seal()

	app.AccountKeeper = authkeeper.NewAccountKeeper(
		appCodec,
		keys[authtypes.StoreKey],
		authtypes.ProtoBaseAccount,
		maccPerms,
		appparams.AccountAddressPrefix,
		govModuleAddr,
	)
	app.BankKeeper = bankkeeper.NewBaseKeeper(
		appCodec,
		keys[banktypes.StoreKey],
		app.AccountKeeper,
		app.ModuleAccountAddrs(),
		govModuleAddr,
	)
	app.StakingKeeper = stakingkeeper.NewKeeper(
		appCodec,
		keys[stakingtypes.StoreKey],
		app.AccountKeeper,
		app.BankKeeper,
		govModuleAddr,
	)
	app.MintKeeper = mintkeeper.NewKeeper(
		appCodec, keys[minttypes.StoreKey], app.StakingKeeper,
		app.AccountKeeper, app.BankKeeper, authtypes.FeeCollectorName,
		govModuleAddr,
	)
	app.DistrKeeper = distrkeeper.NewKeeper(
		appCodec, keys[distrtypes.StoreKey], app.AccountKeeper, app.BankKeeper,
		app.StakingKeeper, authtypes.FeeCollectorName,
		govModuleAddr,
	)
	app.SlashingKeeper = slashingkeeper.NewKeeper(
		appCodec, app.legacyAmino, keys[slashingtypes.StoreKey],
		app.StakingKeeper, govModuleAddr,
	)
	app.FeeGrantKeeper = feegrantkeeper.NewKeeper(
		appCodec, keys[feegrant.StoreKey], app.AccountKeeper,
	)
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
	app.UpgradeKeeper = upgradekeeper.NewKeeper(
		skipUpgradeHeights,
		keys[upgradetypes.StoreKey],
		appCodec,
		homePath,
		app.BaseApp,
		govModuleAddr,
	)

	app.NFTKeeper = nftkeeper.NewKeeper(keys[nftkeeper.StoreKey], appCodec, app.AccountKeeper, app.BankKeeper)

	app.UGovKeeperB = ugovkeeper.NewBuilder(appCodec, keys[ugov.ModuleName])

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

	rewardsAuctionAccs := auctionmodule.SubAccounts()
	app.AuctionKeeperB = auctionkeeper.NewBuilder(
		appCodec,
		keys[auction.StoreKey],
		rewardsAuctionAccs,
		app.BankKeeper,
		app.UGovKeeperB.EmergencyGroup,
	)

	app.LeverageKeeper = leveragekeeper.NewKeeper(
		appCodec,
		keys[leveragetypes.ModuleName],
		keys[authtypes.StoreKey],
		app.BankKeeper,
		app.AccountKeeper,
		app.OracleKeeper,
		app.UGovKeeperB.EmergencyGroup,
		cast.ToBool(appOpts.Get(leveragetypes.FlagEnableLiquidatorQuery)),
		rewardsAuctionAccs.RewardsCollect,
	)

	app.LeverageKeeper.SetTokenHooks(app.OracleKeeper.Hooks())

	app.IncentiveKeeper = incentivekeeper.NewKeeper(
		appCodec,
		keys[incentive.StoreKey],
		app.BankKeeper,
		app.LeverageKeeper,
	)
	app.LeverageKeeper.SetBondHooks(app.IncentiveKeeper.BondHooks())

	app.MetokenKeeperB = metokenkeeper.NewBuilder(
		appCodec,
		keys[metoken.StoreKey],
		app.BankKeeper,
		app.LeverageKeeper,
		app.OracleKeeper,
		app.UGovKeeperB.EmergencyGroup,
		rewardsAuctionAccs.RewardsCollect,
	)

	// register the staking hooks
	// NOTE: stakingKeeper above is passed by reference, so that it will contain these hooks
	app.StakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(
			app.DistrKeeper.Hooks(), app.SlashingKeeper.Hooks(),
		),
	)

	// Create evidence Keeper before IBC to register the IBC light client misbehavior
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
		keys[ibcexported.StoreKey],
		app.GetSubspace(ibcexported.ModuleName),
		app.StakingKeeper,
		app.UpgradeKeeper,
		app.ScopedIBCKeeper,
	)
	app.ICAHostKeeper = icahostkeeper.NewKeeper(
		appCodec, keys[icahosttypes.StoreKey], app.GetSubspace(icahosttypes.SubModuleName),
		app.IBCKeeper.ChannelKeeper, // app.IBCFeeKeeper, // use ics29 fee as ics4Wrapper in middleware stack
		app.IBCKeeper.ChannelKeeper, &app.IBCKeeper.PortKeeper,
		app.AccountKeeper, scopedICAHostKeeper, app.MsgServiceRouter(),
	)

	app.ICAHostKeeper.WithQueryRouter(app.GRPCQueryRouter())

	// UIbcQuotaKeeper implements ibcporttypes.ICS4Wrapper
	app.UIbcQuotaKeeperB = uibcquota.NewBuilder(
		appCodec, keys[uibc.StoreKey],
		app.LeverageKeeper, uibcoracle.FromUmeeAvgPriceOracle(app.OracleKeeper), app.UGovKeeperB.EmergencyGroup,
	)

	/**********
	 * ICS20 (Transfer) Middleware Stacks
	 * SendPacket, originates from the application to an IBC channel:
	   transferKeeper.SendPacket -> uibcquota.SendPacket -> channel.SendPacket
	 * RecvPacket, message that originates from an IBC channel and goes down to app, the flow is the other way
	   channel.RecvPacket -> uibcquota.OnRecvPacket -> forward.OnRecvPacket -> transfer.OnRecvPacket

	* Note that the forward middleware is only integrated on the "receive" direction.
	  It can be safely skipped when sending.

	* transfer stack contains (from top to bottom):
	  - Umee IBC Transfer
	  - IBC Rate Limit Middleware
	  - Packet Forward Middleware
	 **********/

	quotaICS4 := uics20.NewICS4(app.IBCKeeper.ChannelKeeper, app.UIbcQuotaKeeperB)

	// Create Transfer Keeper and pass IBCFeeKeeper as expected Channel and PortKeeper
	// since fee middleware will wrap the IBCKeeper for underlying application.
	app.IBCTransferKeeper = ibctransferkeeper.NewKeeper(
		appCodec, keys[ibctransfertypes.StoreKey], app.GetSubspace(ibctransfertypes.ModuleName),
		quotaICS4, // ISC4 Wrapper: fee IBC middleware
		app.IBCKeeper.ChannelKeeper, &app.IBCKeeper.PortKeeper,
		app.AccountKeeper, app.BankKeeper, app.ScopedTransferKeeper,
	)

	app.FeeabsKeeper = feeabskeeper.NewKeeper(
		appCodec,
		app.keys[feeabstypes.StoreKey],
		app.GetSubspace(feeabstypes.ModuleName),
		app.StakingKeeper,
		app.AccountKeeper,
		app.BankKeeper,
		app.IBCTransferKeeper,
		app.IBCKeeper.ChannelKeeper,
		&app.IBCKeeper.PortKeeper,
		app.ScopedFeeabsKeeper,
	)

	// Packet Forward Middleware
	// Initialize packet forward middleware router
	app.PacketForwardKeeper = packetforwardkeeper.NewKeeper(
		appCodec,
		keys[packetforwardtypes.StoreKey],
		app.IBCTransferKeeper,
		app.IBCKeeper.ChannelKeeper,
		app.DistrKeeper,
		app.BankKeeper,
		quotaICS4, // ISC4 Wrapper: fee IBC middleware
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// create IBC module from bottom to top of stack
	var transferStack ibcporttypes.IBCModule
	transferStack = ibctransfer.NewIBCModule(app.IBCTransferKeeper)
	// transferStack = ibcfee.NewIBCMiddleware(transferStack, app.IBCFeeKeeper)
	transferStack = packetforward.NewIBCMiddleware(
		transferStack,
		app.PacketForwardKeeper,
		0, // retries on timeout
		packetforwardkeeper.DefaultForwardTransferPacketTimeoutTimestamp, // forward timeout
		packetforwardkeeper.DefaultRefundTransferPacketTimeoutTimestamp,  // refund timeout
	)
	// NOTE: uics20 module must be the last middleware. We need to be sure there is no other code
	// that will manipulate packet between the UICS20 middleware (which executes transfer hooks)
	// and the transfer app.
	transferStack = uics20.NewICS20Module(transferStack, appCodec,
		app.UIbcQuotaKeeperB,
		leveragekeeper.NewMsgServerImpl(app.LeverageKeeper))

	// Create Interchain Accounts Controller Stack
	// SendPacket, since it is originating from the application to core IBC:
	// icaAuthModuleKeeper.SendTx -> icaController.SendPacket -> fee.SendPacket -> channel.SendPacket

	// RecvPacket, message that originates from core IBC and goes down to app, the flow is:
	// channel.RecvPacket -> fee.OnRecvPacket -> icaHost.OnRecvPacket
	var icaHostStack ibcporttypes.IBCModule = icahost.NewIBCModule(app.ICAHostKeeper)
	// icaHostStack = ibcfee.NewIBCMiddleware(icaHostStack, app.IBCFeeKeeper)

	/*
		Create fee enabled wasm ibc Stack
		var wasmStack ibcporttypes.IBCModule
		wasmStack = wasm.NewIBCHandler(app.WasmKeeper, app.IBCKeeper.ChannelKeeper, app.IBCFeeKeeper)
		wasmStack = ibcfee.NewIBCMiddleware(wasmStack, app.IBCFeeKeeper)
	*/

	// Create static IBC router, add app routes, then set and seal it
	ibcRouter := ibcporttypes.NewRouter().
		AddRoute(ibctransfertypes.ModuleName, transferStack).
		AddRoute(icahosttypes.SubModuleName, icaHostStack).
		AddRoute(feeabstypes.ModuleName, feeabsmodule.NewIBCModule(appCodec, app.FeeabsKeeper))
	/*
		// we will add cosmwasm IBC routing later
		AddRoute(wasm.ModuleName, wasmStack).
		// we don't integrate the controller now
		AddRoute(intertxtypes.ModuleName, icaControllerStack).
		AddRoute(icacontrollertypes.SubModuleName, icaControllerStack).
	*/
	app.IBCKeeper.SetRouter(ibcRouter)

	// Register the proposal types
	// Deprecated: Avoid adding new handlers, instead use the new proposal flow
	// by granting the governance module the right to execute the message.
	// See: https://github.com/cosmos/cosmos-sdk/blob/release/v0.46.x/x/gov/spec/01_concepts.md#proposal-messages
	govRouter := govv1beta1.NewRouter()
	govRouter.
		AddRoute(govtypes.RouterKey, govv1beta1.ProposalHandler).
		AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(app.ParamsKeeper)).
		AddRoute(upgradetypes.RouterKey, upgrade.NewSoftwareUpgradeProposalHandler(app.UpgradeKeeper)).
		AddRoute(ibcclienttypes.RouterKey, ibcclient.NewClientProposalHandler(app.IBCKeeper.ClientKeeper)).
		AddRoute(feeabstypes.RouterKey, feeabsmodule.NewHostZoneProposal(app.FeeabsKeeper))

	govConfig := govtypes.DefaultConfig()
	govConfig.MaxMetadataLen = 800
	app.GovKeeper = govkeeper.NewKeeper(
		appCodec, keys[govtypes.StoreKey], app.AccountKeeper, app.BankKeeper,
		app.StakingKeeper, app.MsgServiceRouter(), govConfig,
		govModuleAddr,
	)
	app.GovKeeper.SetLegacyRouter(govRouter)

	var err error
	wasmDir := filepath.Join(homePath, "wasm")
	app.wasmCfg, err = wasm.ReadWasmConfig(appOpts)
	if err != nil {
		panic(fmt.Sprintf("error while reading wasm config: %s", err))
	}

	// Register stargate queries
	wasmOpts = append(wasmOpts, uwasm.RegisterStargateQueries(*bApp.GRPCQueryRouter(), appCodec)...)
	availableCapabilities := strings.Join(AllCapabilities(), ",")
	app.WasmKeeper = wasmkeeper.NewKeeper(
		appCodec,
		keys[wasmtypes.StoreKey],
		app.AccountKeeper,
		app.BankKeeper,
		app.StakingKeeper,
		distrkeeper.NewQuerier(app.DistrKeeper),
		app.IBCKeeper.ChannelKeeper,
		app.IBCKeeper.ChannelKeeper,
		&app.IBCKeeper.PortKeeper,
		app.ScopedWasmKeeper,   // capabilities
		&app.IBCTransferKeeper, // ICS20TransferPortSource
		app.MsgServiceRouter(),
		nil,
		wasmDir,
		app.wasmCfg,
		availableCapabilities,
		govModuleAddr,
		wasmOpts...,
	)

	/****************
	 Module Options
	/****************/

	inflationCalculator := inflation.Calculator{
		UgovKeeperB: app.UGovKeeperB.Params,
		MintKeeper:  &app.MintKeeper,
	}

	// NOTE: Any module instantiated in the module manager that is later modified
	// must be passed by reference here.
	appModules := []module.AppModule{
		genutil.NewAppModule(
			app.AccountKeeper,
			app.StakingKeeper,
			app.DeliverTx,
			txConfig,
		),
		auth.NewAppModule(appCodec, app.AccountKeeper, authsims.RandomGenesisAccounts, app.GetSubspace(authtypes.ModuleName)),
		vesting.NewAppModule(app.AccountKeeper, app.BankKeeper),
		bank.NewAppModule(appCodec, app.BankKeeper, app.AccountKeeper, app.GetSubspace(banktypes.ModuleName)),
		capability.NewAppModule(appCodec, *app.CapabilityKeeper, true),
		feegrantmodule.NewAppModule(appCodec, app.AccountKeeper, app.BankKeeper, app.FeeGrantKeeper, app.interfaceRegistry),
		gov.NewAppModule(appCodec, app.GovKeeper, app.AccountKeeper, app.BankKeeper, app.GetSubspace(govtypes.ModuleName)),
		mint.NewAppModule(appCodec, app.MintKeeper, app.AccountKeeper, inflationCalculator.InflationRate, app.GetSubspace(minttypes.ModuleName)),             //nolint: lll
		slashing.NewAppModule(appCodec, app.SlashingKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, app.GetSubspace(slashingtypes.ModuleName)), //nolint: lll
		distr.NewAppModule(appCodec, app.DistrKeeper, app.AccountKeeper, app.BankKeeper, app.StakingKeeper, app.GetSubspace(distrtypes.ModuleName)),          //nolint: lll
		staking.NewAppModule(appCodec, app.StakingKeeper, app.AccountKeeper, app.BankKeeper, app.GetSubspace(stakingtypes.ModuleName)),                       //nolint: lll
		upgrade.NewAppModule(app.UpgradeKeeper),
		evidence.NewAppModule(app.EvidenceKeeper),
		ibc.NewAppModule(app.IBCKeeper),
		params.NewAppModule(app.ParamsKeeper),
		authzmodule.NewAppModule(appCodec, app.AuthzKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		groupmodule.NewAppModule(appCodec, app.GroupKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		nftmodule.NewAppModule(appCodec, app.NFTKeeper, app.AccountKeeper, app.BankKeeper, app.interfaceRegistry),
		ibctransfer.NewAppModule(app.IBCTransferKeeper),
		ica.NewAppModule(nil, &app.ICAHostKeeper),
		leverage.NewAppModule(appCodec, app.LeverageKeeper, app.AccountKeeper, app.BankKeeper),
		oracle.NewAppModule(appCodec, app.OracleKeeper, app.AccountKeeper, app.BankKeeper),
		uibcmodule.NewAppModule(appCodec, app.UIbcQuotaKeeperB),
		packetforward.NewAppModule(app.PacketForwardKeeper, app.GetSubspace(packetforwardtypes.ModuleName)),
		ugovmodule.NewAppModule(appCodec, app.UGovKeeperB),
		wasm.NewAppModule(app.appCodec, &app.WasmKeeper, app.StakingKeeper, app.AccountKeeper, app.BankKeeper, app.MsgServiceRouter(), app.GetSubspace(wasmtypes.ModuleName)), //nolint: lll
		incentivemodule.NewAppModule(appCodec, app.IncentiveKeeper, app.BankKeeper, app.LeverageKeeper),
		metokenmodule.NewAppModule(appCodec, app.MetokenKeeperB),
		auctionmodule.NewAppModule(appCodec, app.AuctionKeeperB, app.BankKeeper),
		feeabsmodule.NewAppModule(appCodec, app.FeeabsKeeper),
	}
	// if Experimental {}

	app.mm = module.NewManager(appModules...)

	// During begin block slashing happens after distr.BeginBlocker so that
	// there is nothing left over in the validator fee pool, so as to keep the
	// CanWithdrawInvariant invariant.
	// NOTE: staking module is required if HistoricalEntries param > 0
	// NOTE: capability module's beginblocker must come before any modules using capabilities (e.g. IBC)
	beginBlockers := []string{
		upgradetypes.ModuleName,
		capabilitytypes.ModuleName,
		minttypes.ModuleName,
		distrtypes.ModuleName,
		slashingtypes.ModuleName,
		evidencetypes.ModuleName,
		stakingtypes.ModuleName,
		ibcexported.ModuleName,
		ibctransfertypes.ModuleName,
		authtypes.ModuleName,
		banktypes.ModuleName,
		govtypes.ModuleName,
		genutiltypes.ModuleName,
		authz.ModuleName,
		feegrant.ModuleName,
		nft.ModuleName,
		group.ModuleName,
		paramstypes.ModuleName,
		vestingtypes.ModuleName,
		icatypes.ModuleName, //  ibcfeetypes.ModuleName,
		leveragetypes.ModuleName,
		metoken.ModuleName,
		oracletypes.ModuleName,
		uibc.ModuleName,
		packetforwardtypes.ModuleName,
		ugov.ModuleName,
		wasmtypes.ModuleName,
		incentive.ModuleName,
		auction.ModuleName,
		feeabstypes.ModuleName,
	}
	endBlockers := []string{
		metoken.ModuleName,     // must be before oracle
		oracletypes.ModuleName, // must be before gov and staking
		govtypes.ModuleName, stakingtypes.ModuleName,
		ibcexported.ModuleName, ibctransfertypes.ModuleName,
		capabilitytypes.ModuleName, authtypes.ModuleName, banktypes.ModuleName, distrtypes.ModuleName,
		slashingtypes.ModuleName, minttypes.ModuleName,
		genutiltypes.ModuleName, evidencetypes.ModuleName, authz.ModuleName,
		feegrant.ModuleName, nft.ModuleName, group.ModuleName,
		paramstypes.ModuleName, upgradetypes.ModuleName, vestingtypes.ModuleName,
		icatypes.ModuleName, //  ibcfeetypes.ModuleName,
		leveragetypes.ModuleName,
		uibc.ModuleName,
		packetforwardtypes.ModuleName,
		ugov.ModuleName,
		wasmtypes.ModuleName,
		incentive.ModuleName,
		auction.ModuleName,
		feeabstypes.ModuleName,
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
		ibcexported.ModuleName, genutiltypes.ModuleName, evidencetypes.ModuleName,
		authz.ModuleName,
		ibctransfertypes.ModuleName, icatypes.ModuleName, // ibcfeetypes.ModuleName
		feegrant.ModuleName, nft.ModuleName, group.ModuleName,
		paramstypes.ModuleName, upgradetypes.ModuleName, vestingtypes.ModuleName,

		oracletypes.ModuleName,
		leveragetypes.ModuleName,
		uibc.ModuleName,
		packetforwardtypes.ModuleName,
		ugov.ModuleName,
		wasmtypes.ModuleName,
		incentive.ModuleName,
		metoken.ModuleName,
		auction.ModuleName,
		feeabstypes.ModuleName,
	}
	orderMigrations := []string{
		capabilitytypes.ModuleName, authtypes.ModuleName, banktypes.ModuleName, distrtypes.ModuleName,
		stakingtypes.ModuleName, slashingtypes.ModuleName, govtypes.ModuleName, minttypes.ModuleName,
		ibcexported.ModuleName, genutiltypes.ModuleName, evidencetypes.ModuleName,
		authz.ModuleName, ibctransfertypes.ModuleName, icatypes.ModuleName, // ibcfeetypes.ModuleName
		feegrant.ModuleName, nft.ModuleName, group.ModuleName,
		paramstypes.ModuleName, upgradetypes.ModuleName, vestingtypes.ModuleName,

		oracletypes.ModuleName,
		leveragetypes.ModuleName,
		uibc.ModuleName,
		packetforwardtypes.ModuleName,
		ugov.ModuleName,
		wasmtypes.ModuleName,
		incentive.ModuleName,
		metoken.ModuleName,
		auction.ModuleName,
		feeabstypes.ModuleName,
	}
	// if Experimental {}

	app.mm.SetOrderBeginBlockers(beginBlockers...)
	app.mm.SetOrderEndBlockers(endBlockers...)
	app.mm.SetOrderInitGenesis(initGenesis...)
	app.mm.SetOrderMigrations(orderMigrations...)

	app.configurator = module.NewConfigurator(app.appCodec, app.MsgServiceRouter(), app.GRPCQueryRouter())
	app.mm.RegisterServices(app.configurator)

	// RegisterUpgradeHandlers is used for registering any on-chain upgrades.
	// Make sure it's called after `app.mm` and `app.configurator` are set.
	app.RegisterUpgradeHandlers()

	autocliv1.RegisterQueryServer(app.GRPCQueryRouter(), runtimeservices.NewAutoCLIQueryService(app.mm.Modules))

	reflectionSvc, err := runtimeservices.NewReflectionService()
	if err != nil {
		panic(err)
	}
	reflectionv1.RegisterReflectionServiceServer(app.GRPCQueryRouter(), reflectionSvc)

	// add test gRPC service for testing gRPC queries in isolation
	testdata.RegisterQueryServer(app.GRPCQueryRouter(), testdata.QueryImpl{})

	//
	// create the simulation manager and define the order of the modules for deterministic simulations
	//
	overrideModules := map[string]module.AppModuleSimulation{
		authtypes.ModuleName: auth.NewAppModule(
			app.appCodec, app.AccountKeeper, authsims.RandomGenesisAccounts,
			app.GetSubspace(authtypes.ModuleName),
		),
	}
	simStateModules := genmap.Pick(
		app.mm.Modules,
		[]string{
			stakingtypes.ModuleName, authtypes.ModuleName, oracletypes.ModuleName,
			ibcexported.ModuleName,
		},
	)
	// TODO: Ensure x/leverage, x/incentive implement simulator and add it here:
	simTestModules := genmap.Pick(
		simStateModules,
		[]string{oracletypes.ModuleName, ibcexported.ModuleName},
	)

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
	app.setAnteHandler(txConfig, keys[wasmtypes.StoreKey])
	app.setPostHandler()

	if manager := app.SnapshotManager(); manager != nil {
		err := manager.RegisterExtensions(
			wasmkeeper.NewWasmSnapshotter(app.CommitMultiStore(), &app.WasmKeeper),
		)
		if err != nil {
			panic(fmt.Errorf("failed to register snapshot extension: %s", err))
		}
	}

	if loadLatest {
		if err := app.LoadLatestVersion(); err != nil {
			tmos.Exit(fmt.Sprintf("failed to load latest version: %s", err))
		}

		// Initialize pinned codes in wasmvm as they are not persisted there
		ctx := app.NewUncachedContext(true, tmproto.Header{})
		if err := app.WasmKeeper.InitializePinnedCodes(ctx); err != nil {
			tmos.Exit(fmt.Sprintf("failed initialize pinned codes %s", err))
		}
	}

	return app
}

func (app *UmeeApp) setAnteHandler(txConfig client.TxConfig, wasmStoreKey *storetypes.KVStoreKey) {
	anteHandler, err := customante.NewAnteHandler(
		customante.HandlerOptions{
			AccountKeeper:     app.AccountKeeper,
			BankKeeper:        app.BankKeeper,
			OracleKeeper:      app.OracleKeeper,
			IBCKeeper:         app.IBCKeeper,
			SignModeHandler:   txConfig.SignModeHandler(),
			FeegrantKeeper:    app.FeeGrantKeeper,
			SigGasConsumer:    ante.DefaultSigVerificationGasConsumer,
			WasmConfig:        &app.wasmCfg,
			TXCounterStoreKey: wasmStoreKey,
		},
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
		// Skip the feeabstypes.ModuleName to allow the feeabs module to receive tokens
		if acc == feeabstypes.ModuleName {
			continue
		}
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
	authtx.RegisterTxService(app.GRPCQueryRouter(), clientCtx, app.Simulate, app.interfaceRegistry)
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
	paramsKeeper.Subspace(govtypes.ModuleName).WithKeyTable(govv1.ParamKeyTable()) //nolint: staticcheck
	paramsKeeper.Subspace(ibctransfertypes.ModuleName)
	paramsKeeper.Subspace(ibcexported.ModuleName)
	paramsKeeper.Subspace(icahosttypes.SubModuleName)
	paramsKeeper.Subspace(packetforwardtypes.ModuleName).WithKeyTable(packetforwardtypes.ParamKeyTable())
	paramsKeeper.Subspace(oracletypes.ModuleName)
	paramsKeeper.Subspace(wasmtypes.ModuleName)
	paramsKeeper.Subspace(feeabstypes.ModuleName)

	return paramsKeeper
}

func getGovProposalHandlers() []govclient.ProposalHandler {
	handlers := []govclient.ProposalHandler{
		paramsclient.ProposalHandler,
		upgradeclient.LegacyProposalHandler,
		upgradeclient.LegacyCancelProposalHandler,
		ibcclientclient.UpdateClientProposalHandler,
		ibcclientclient.UpgradeProposalHandler,
	}

	return handlers
}
