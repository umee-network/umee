package keepers

import (
	"fmt"
	"path/filepath"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	storetypes "github.com/cosmos/cosmos-sdk/store/types"
	authkeeper "github.com/cosmos/cosmos-sdk/x/auth/keeper"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	authzkeeper "github.com/cosmos/cosmos-sdk/x/authz/keeper"
	bankkeeper "github.com/cosmos/cosmos-sdk/x/bank/keeper"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	capabilitykeeper "github.com/cosmos/cosmos-sdk/x/capability/keeper"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	consensusparamskeeper "github.com/cosmos/cosmos-sdk/x/consensus/keeper"
	consensusparamstypes "github.com/cosmos/cosmos-sdk/x/consensus/types"
	crisiskeeper "github.com/cosmos/cosmos-sdk/x/crisis/keeper"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distrkeeper "github.com/cosmos/cosmos-sdk/x/distribution/keeper"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	evidencekeeper "github.com/cosmos/cosmos-sdk/x/evidence/keeper"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	feegrantkeeper "github.com/cosmos/cosmos-sdk/x/feegrant/keeper"
	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/group"
	groupkeeper "github.com/cosmos/cosmos-sdk/x/group/keeper"
	mintkeeper "github.com/cosmos/cosmos-sdk/x/mint/keeper"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	nftkeeper "github.com/cosmos/cosmos-sdk/x/nft/keeper"
	"github.com/cosmos/cosmos-sdk/x/params"
	paramskeeper "github.com/cosmos/cosmos-sdk/x/params/keeper"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	paramproposal "github.com/cosmos/cosmos-sdk/x/params/types/proposal"
	slashingkeeper "github.com/cosmos/cosmos-sdk/x/slashing/keeper"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	stakingkeeper "github.com/cosmos/cosmos-sdk/x/staking/keeper"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	upgradekeeper "github.com/cosmos/cosmos-sdk/x/upgrade/keeper"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/spf13/cast"

	govv1beta1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1beta1"
	packetforwardkeeper "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward/keeper"
	packetforwardtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward/types"
	icahostkeeper "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/keeper"
	icahosttypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/host/types"
	ibctransferkeeper "github.com/cosmos/ibc-go/v7/modules/apps/transfer/keeper"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibcclient "github.com/cosmos/ibc-go/v7/modules/core/02-client"
	ibcclienttypes "github.com/cosmos/ibc-go/v7/modules/core/02-client/types"
	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibckeeper "github.com/cosmos/ibc-go/v7/modules/core/keeper"
	ibctesting "github.com/cosmos/ibc-go/v7/testing/types"

	// cosmwasm
	"github.com/CosmWasm/wasmd/x/wasm"
	wasmkeeper "github.com/CosmWasm/wasmd/x/wasm/keeper"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	uwasm "github.com/umee-network/umee/v6/app/wasm"

	"github.com/umee-network/umee/v6/x/incentive"
	incentivekeeper "github.com/umee-network/umee/v6/x/incentive/keeper"
	leveragekeeper "github.com/umee-network/umee/v6/x/leverage/keeper"
	leveragetypes "github.com/umee-network/umee/v6/x/leverage/types"
	"github.com/umee-network/umee/v6/x/metoken"
	oraclekeeper "github.com/umee-network/umee/v6/x/oracle/keeper"
	oracletypes "github.com/umee-network/umee/v6/x/oracle/types"
	"github.com/umee-network/umee/v6/x/ugov"
	ugovkeeper "github.com/umee-network/umee/v6/x/ugov/keeper"
	"github.com/umee-network/umee/v6/x/uibc"
	uibcoracle "github.com/umee-network/umee/v6/x/uibc/oracle"

	// umee ibc-transfer and quota for ibc-transfer

	uibcquota "github.com/umee-network/umee/v6/x/uibc/quota"
	"github.com/umee-network/umee/v6/x/uibc/uics20"

	appparams "github.com/umee-network/umee/v6/app/params"
	metokenkeeper "github.com/umee-network/umee/v6/x/metoken/keeper"
)

type AppKeepers struct {
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
	CrisisKeeper          *crisiskeeper.Keeper
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
	IncentiveKeeper     incentivekeeper.Keeper
	OracleKeeper        oraclekeeper.Keeper
	UIbcQuotaKeeperB    uibcquota.KeeperBuilder
	UGovKeeperB         ugovkeeper.Builder
	MetokenKeeperB      metokenkeeper.Builder

	// make scoped keepers public for testing purposes
	ScopedIBCKeeper      capabilitykeeper.ScopedKeeper
	ScopedTransferKeeper capabilitykeeper.ScopedKeeper
	ScopedWasmKeeper     capabilitykeeper.ScopedKeeper
}

func NewAppKeepers(
	appCodec codec.Codec,
	bApp *baseapp.BaseApp,
	legacyAmino *codec.LegacyAmino,
	maccPerms map[string][]string,
	moduleAccountAddrs map[string]bool,
	skipUpgradeHeights map[int64]bool,
	homePath string,
	invCheckPeriod uint,
	wasmCapabilities string,
	appOpts servertypes.AppOptions,
	wasmOpts []wasmkeeper.Option,
) AppKeepers {
	appKeepers := AppKeepers{}

	// Set keys KVStoreKey, TransientStoreKey, MemoryStoreKey
	appKeepers.GenerateKeys()

	govModuleAddr := authtypes.NewModuleAddress(govtypes.ModuleName).String()

	appKeepers.ParamsKeeper = initParamsKeeper(
		appCodec,
		legacyAmino,
		appKeepers.keys[paramstypes.StoreKey],
		appKeepers.tkeys[paramstypes.TStoreKey],
	)

	// set the BaseApp's parameter store
	appKeepers.ConsensusParamsKeeper = consensusparamskeeper.NewKeeper(
		appCodec,
		appKeepers.keys[consensusparamstypes.StoreKey],
		govModuleAddr,
	)
	bApp.SetParamStore(&appKeepers.ConsensusParamsKeeper)

	appKeepers.CapabilityKeeper = capabilitykeeper.NewKeeper(
		appCodec,
		appKeepers.keys[capabilitytypes.StoreKey],
		appKeepers.memKeys[capabilitytypes.MemStoreKey],
	)

	// grant capabilities for the ibc and ibc-transfer modules
	appKeepers.ScopedIBCKeeper = appKeepers.CapabilityKeeper.ScopeToModule(ibcexported.ModuleName)
	appKeepers.ScopedTransferKeeper = appKeepers.CapabilityKeeper.ScopeToModule(ibctransfertypes.ModuleName)
	scopedICAHostKeeper := appKeepers.CapabilityKeeper.ScopeToModule(icahosttypes.SubModuleName)
	appKeepers.ScopedWasmKeeper = appKeepers.CapabilityKeeper.ScopeToModule(wasmtypes.ModuleName)

	// Applications that wish to enforce statically created ScopedKeepers should call `Seal` after creating
	// their scoped modules in `NewApp` with `ScopeToModule`
	appKeepers.CapabilityKeeper.Seal()

	appKeepers.AccountKeeper = authkeeper.NewAccountKeeper(
		appCodec,
		appKeepers.keys[authtypes.StoreKey],
		authtypes.ProtoBaseAccount,
		maccPerms,
		appparams.AccountAddressPrefix,
		govModuleAddr,
	)

	appKeepers.BankKeeper = bankkeeper.NewBaseKeeper(
		appCodec,
		appKeepers.keys[banktypes.StoreKey],
		appKeepers.AccountKeeper,
		moduleAccountAddrs,
		govModuleAddr,
	)

	appKeepers.StakingKeeper = stakingkeeper.NewKeeper(
		appCodec,
		appKeepers.keys[stakingtypes.StoreKey],
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		govModuleAddr,
	)

	appKeepers.MintKeeper = mintkeeper.NewKeeper(
		appCodec, appKeepers.keys[minttypes.StoreKey], appKeepers.StakingKeeper,
		appKeepers.AccountKeeper, appKeepers.BankKeeper, authtypes.FeeCollectorName,
		govModuleAddr,
	)
	appKeepers.DistrKeeper = distrkeeper.NewKeeper(
		appCodec, appKeepers.keys[distrtypes.StoreKey], appKeepers.AccountKeeper, appKeepers.BankKeeper,
		appKeepers.StakingKeeper, authtypes.FeeCollectorName,
		govModuleAddr,
	)
	appKeepers.SlashingKeeper = slashingkeeper.NewKeeper(
		appCodec, legacyAmino, appKeepers.keys[slashingtypes.StoreKey],
		appKeepers.StakingKeeper, govModuleAddr,
	)

	appKeepers.CrisisKeeper = crisiskeeper.NewKeeper(
		appCodec, appKeepers.keys[crisistypes.StoreKey], invCheckPeriod, appKeepers.BankKeeper,
		authtypes.FeeCollectorName, govModuleAddr,
	)

	appKeepers.FeeGrantKeeper = feegrantkeeper.NewKeeper(
		appCodec, appKeepers.keys[feegrant.StoreKey], appKeepers.AccountKeeper,
	)

	appKeepers.AuthzKeeper = authzkeeper.NewKeeper(
		appKeepers.keys[authzkeeper.StoreKey],
		appCodec,
		bApp.MsgServiceRouter(),
		appKeepers.AccountKeeper,
	)

	groupConfig := group.DefaultConfig()
	groupConfig.MaxMetadataLen = 600
	appKeepers.GroupKeeper = groupkeeper.NewKeeper(
		appKeepers.keys[group.StoreKey],
		appCodec,
		bApp.MsgServiceRouter(),
		appKeepers.AccountKeeper,
		groupConfig,
	)

	// set the governance module account as the authority for conducting upgrades
	appKeepers.UpgradeKeeper = upgradekeeper.NewKeeper(
		skipUpgradeHeights,
		appKeepers.keys[upgradetypes.StoreKey],
		appCodec,
		homePath,
		bApp,
		govModuleAddr,
	)

	appKeepers.NFTKeeper = nftkeeper.NewKeeper(
		appKeepers.keys[nftkeeper.StoreKey], appCodec, appKeepers.AccountKeeper, appKeepers.BankKeeper,
	)

	appKeepers.UGovKeeperB = ugovkeeper.NewKeeperBuilder(appCodec, appKeepers.keys[ugov.ModuleName])

	appKeepers.OracleKeeper = oraclekeeper.NewKeeper(
		appCodec,
		appKeepers.keys[oracletypes.ModuleName],
		appKeepers.GetSubspace(oracletypes.ModuleName),
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		appKeepers.DistrKeeper,
		appKeepers.StakingKeeper,
		distrtypes.ModuleName,
	)

	appKeepers.LeverageKeeper = leveragekeeper.NewKeeper(
		appCodec,
		appKeepers.keys[leveragetypes.ModuleName],
		appKeepers.BankKeeper,
		appKeepers.OracleKeeper,
		appKeepers.UGovKeeperB.EmergencyGroup,
		cast.ToBool(appOpts.Get(leveragetypes.FlagEnableLiquidatorQuery)),
		authtypes.NewModuleAddress(metoken.ModuleName),
	)

	appKeepers.LeverageKeeper.SetTokenHooks(appKeepers.OracleKeeper.Hooks())

	appKeepers.IncentiveKeeper = incentivekeeper.NewKeeper(
		appCodec,
		appKeepers.keys[incentive.StoreKey],
		appKeepers.BankKeeper,
		appKeepers.LeverageKeeper,
	)
	appKeepers.LeverageKeeper.SetBondHooks(appKeepers.IncentiveKeeper.BondHooks())

	appKeepers.MetokenKeeperB = metokenkeeper.NewKeeperBuilder(
		appCodec,
		appKeepers.keys[metoken.StoreKey],
		appKeepers.BankKeeper,
		appKeepers.LeverageKeeper,
		appKeepers.OracleKeeper,
		appKeepers.UGovKeeperB.EmergencyGroup,
	)

	// register the staking hooks
	// NOTE: stakingKeeper above is passed by reference, so that it will contain these hooks
	appKeepers.StakingKeeper.SetHooks(
		stakingtypes.NewMultiStakingHooks(
			appKeepers.DistrKeeper.Hooks(),
			appKeepers.SlashingKeeper.Hooks(),
		),
	)

	// Create evidence Keeper so we can register the IBC light client misbehavior
	// evidence route.
	evidenceKeeper := evidencekeeper.NewKeeper(
		appCodec,
		appKeepers.keys[evidencetypes.StoreKey],
		appKeepers.StakingKeeper,
		appKeepers.SlashingKeeper,
	)
	// If evidence needs to be handled for the app, set routes in router here and seal
	appKeepers.EvidenceKeeper = *evidenceKeeper

	appKeepers.IBCKeeper = ibckeeper.NewKeeper(
		appCodec,
		appKeepers.keys[ibcexported.StoreKey],
		appKeepers.GetSubspace(ibcexported.ModuleName),
		appKeepers.StakingKeeper,
		appKeepers.UpgradeKeeper,
		appKeepers.ScopedIBCKeeper,
	)
	appKeepers.ICAHostKeeper = icahostkeeper.NewKeeper(
		appCodec, appKeepers.keys[icahosttypes.StoreKey], appKeepers.GetSubspace(icahosttypes.SubModuleName),
		appKeepers.IBCKeeper.ChannelKeeper, // appKeepers.IBCFeeKeeper, // use ics29 fee as ics4Wrapper in middleware stack
		appKeepers.IBCKeeper.ChannelKeeper, &appKeepers.IBCKeeper.PortKeeper,
		appKeepers.AccountKeeper, scopedICAHostKeeper, bApp.MsgServiceRouter(),
	)

	// UIbcQuotaKeeper implements ibcporttypes.ICS4Wrapper
	appKeepers.UIbcQuotaKeeperB = uibcquota.NewKeeperBuilder(
		appCodec, appKeepers.keys[uibc.StoreKey], appKeepers.LeverageKeeper,
		uibcoracle.FromUmeeAvgPriceOracle(appKeepers.OracleKeeper), appKeepers.UGovKeeperB.EmergencyGroup,
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

	quotaICS4 := uics20.NewICS4(appKeepers.IBCKeeper.ChannelKeeper, appKeepers.UIbcQuotaKeeperB)

	// Create Transfer Keeper and pass IBCFeeKeeper as expected Channel and PortKeeper
	// since fee middleware will wrap the IBCKeeper for underlying application.
	appKeepers.IBCTransferKeeper = ibctransferkeeper.NewKeeper(
		appCodec, appKeepers.keys[ibctransfertypes.StoreKey], appKeepers.GetSubspace(ibctransfertypes.ModuleName),
		quotaICS4, // ISC4 Wrapper: fee IBC middleware
		appKeepers.IBCKeeper.ChannelKeeper, &appKeepers.IBCKeeper.PortKeeper,
		appKeepers.AccountKeeper, appKeepers.BankKeeper, appKeepers.ScopedTransferKeeper,
	)

	// Packet Forward Middleware
	// Initialize packet forward middleware router
	appKeepers.PacketForwardKeeper = packetforwardkeeper.NewKeeper(
		appCodec,
		appKeepers.keys[packetforwardtypes.StoreKey],
		appKeepers.IBCTransferKeeper,
		appKeepers.IBCKeeper.ChannelKeeper,
		appKeepers.DistrKeeper,
		appKeepers.BankKeeper,
		quotaICS4, // ISC4 Wrapper: fee IBC middleware
		authtypes.NewModuleAddress(govtypes.ModuleName).String(),
	)

	// Register the proposal types
	// Deprecated: Avoid adding new handlers, instead use the new proposal flow
	// by granting the governance module the right to execute the message.
	// See: https://github.com/cosmos/cosmos-sdk/blob/release/v0.46.x/x/gov/spec/01_concepts.md#proposal-messages
	govRouter := govv1beta1.NewRouter()
	govRouter.
		AddRoute(govtypes.RouterKey, govv1beta1.ProposalHandler).
		AddRoute(paramproposal.RouterKey, params.NewParamChangeProposalHandler(appKeepers.ParamsKeeper)).
		AddRoute(upgradetypes.RouterKey, upgrade.NewSoftwareUpgradeProposalHandler(appKeepers.UpgradeKeeper)).
		AddRoute(ibcclienttypes.RouterKey, ibcclient.NewClientProposalHandler(appKeepers.IBCKeeper.ClientKeeper))

	govConfig := govtypes.DefaultConfig()
	govConfig.MaxMetadataLen = 800
	appKeepers.GovKeeper = govkeeper.NewKeeper(
		appCodec, appKeepers.keys[govtypes.StoreKey], appKeepers.AccountKeeper, appKeepers.BankKeeper,
		appKeepers.StakingKeeper, bApp.MsgServiceRouter(), govConfig,
		govModuleAddr,
	)

	appKeepers.GovKeeper.SetLegacyRouter(govRouter)

	var err error
	wasmDir := filepath.Join(homePath, "wasm")
	wasmCfg, err := wasm.ReadWasmConfig(appOpts)
	if err != nil {
		panic(fmt.Sprintf("error while reading wasm config: %s", err))
	}

	// Register umee custom plugin to wasm
	wasmOpts = append(
		uwasm.RegisterCustomPlugins(
			appKeepers.LeverageKeeper, appKeepers.OracleKeeper, appKeepers.IncentiveKeeper,
			appKeepers.MetokenKeeperB,
		),
		wasmOpts...,
	)
	// Register stargate queries
	wasmOpts = append(wasmOpts, uwasm.RegisterStargateQueries(*bApp.GRPCQueryRouter(), appCodec)...)
	appKeepers.WasmKeeper = wasmkeeper.NewKeeper(
		appCodec,
		appKeepers.keys[wasmtypes.StoreKey],
		appKeepers.AccountKeeper,
		appKeepers.BankKeeper,
		appKeepers.StakingKeeper,
		distrkeeper.NewQuerier(appKeepers.DistrKeeper),
		appKeepers.IBCKeeper.ChannelKeeper,
		appKeepers.IBCKeeper.ChannelKeeper,
		&appKeepers.IBCKeeper.PortKeeper,
		appKeepers.ScopedWasmKeeper,   // capabilities
		&appKeepers.IBCTransferKeeper, // ICS20TransferPortSource
		bApp.MsgServiceRouter(),
		nil,
		wasmDir,
		wasmCfg,
		wasmCapabilities,
		govModuleAddr,
		wasmOpts...,
	)

	return appKeepers
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
	paramsKeeper.Subspace(crisistypes.ModuleName)
	paramsKeeper.Subspace(ibctransfertypes.ModuleName)
	paramsKeeper.Subspace(ibcexported.ModuleName)
	paramsKeeper.Subspace(icahosttypes.SubModuleName)
	paramsKeeper.Subspace(packetforwardtypes.ModuleName).WithKeyTable(packetforwardtypes.ParamKeyTable())
	paramsKeeper.Subspace(leveragetypes.ModuleName)
	paramsKeeper.Subspace(oracletypes.ModuleName)
	paramsKeeper.Subspace(wasmtypes.ModuleName)

	return paramsKeeper
}

// GetSubspace returns a param subspace for a given module name.
//
// NOTE: This is solely to be used for testing purposes.
func (appKeepers *AppKeepers) GetSubspace(moduleName string) paramstypes.Subspace {
	subspace, _ := appKeepers.ParamsKeeper.GetSubspace(moduleName)
	return subspace
}

// GetStakingKeeper is used solely for testing purposes.
func (appKeepers *AppKeepers) GetStakingKeeper() ibctesting.StakingKeeper {
	return appKeepers.StakingKeeper
}

// GetIBCKeeper is used solely for testing purposes.
func (appKeepers *AppKeepers) GetIBCKeeper() *ibckeeper.Keeper {
	return appKeepers.IBCKeeper
}

// GetScopedIBCKeeper is used solely for testing purposes.
func (appKeepers *AppKeepers) GetScopedIBCKeeper() capabilitykeeper.ScopedKeeper {
	return appKeepers.ScopedIBCKeeper
}
