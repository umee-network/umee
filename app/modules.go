package app

import (
	"encoding/json"
	"fmt"
	"time"

	"github.com/CosmWasm/wasmd/x/wasm"
	wasmtypes "github.com/CosmWasm/wasmd/x/wasm/types"
	"github.com/cosmos/cosmos-sdk/client"
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/cosmos/cosmos-sdk/x/auth"
	authsims "github.com/cosmos/cosmos-sdk/x/auth/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	"github.com/cosmos/cosmos-sdk/x/auth/vesting"
	"github.com/cosmos/cosmos-sdk/x/authz"
	authzmodule "github.com/cosmos/cosmos-sdk/x/authz/module"
	"github.com/cosmos/cosmos-sdk/x/bank"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	"github.com/cosmos/cosmos-sdk/x/capability"
	"github.com/cosmos/cosmos-sdk/x/crisis"
	crisistypes "github.com/cosmos/cosmos-sdk/x/crisis/types"
	distr "github.com/cosmos/cosmos-sdk/x/distribution"
	distrtypes "github.com/cosmos/cosmos-sdk/x/distribution/types"
	"github.com/cosmos/cosmos-sdk/x/evidence"
	"github.com/cosmos/cosmos-sdk/x/feegrant"
	feegrantmodule "github.com/cosmos/cosmos-sdk/x/feegrant/module"
	"github.com/cosmos/cosmos-sdk/x/genutil"
	genutiltypes "github.com/cosmos/cosmos-sdk/x/genutil/types"
	"github.com/cosmos/cosmos-sdk/x/gov"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/cosmos/cosmos-sdk/x/group"
	groupmodule "github.com/cosmos/cosmos-sdk/x/group/module"
	"github.com/cosmos/cosmos-sdk/x/mint"
	minttypes "github.com/cosmos/cosmos-sdk/x/mint/types"
	"github.com/cosmos/cosmos-sdk/x/nft"
	nftmodule "github.com/cosmos/cosmos-sdk/x/nft/module"
	"github.com/cosmos/cosmos-sdk/x/params"
	"github.com/cosmos/cosmos-sdk/x/slashing"
	slashingtypes "github.com/cosmos/cosmos-sdk/x/slashing/types"
	"github.com/cosmos/cosmos-sdk/x/staking"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	"github.com/cosmos/cosmos-sdk/x/upgrade"
	packetforward "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward"
	packetforwardtypes "github.com/cosmos/ibc-apps/middleware/packet-forward-middleware/v7/packetforward/types"
	ica "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts"
	icatypes "github.com/cosmos/ibc-go/v7/modules/apps/27-interchain-accounts/types"
	ibctransfer "github.com/cosmos/ibc-go/v7/modules/apps/transfer"
	ibctransfertypes "github.com/cosmos/ibc-go/v7/modules/apps/transfer/types"
	ibc "github.com/cosmos/ibc-go/v7/modules/core"
	ibcexported "github.com/cosmos/ibc-go/v7/modules/core/exported"
	ibctm "github.com/cosmos/ibc-go/v7/modules/light-clients/07-tendermint"

	vestingtypes "github.com/cosmos/cosmos-sdk/x/auth/vesting/types"
	capabilitytypes "github.com/cosmos/cosmos-sdk/x/capability/types"
	evidencetypes "github.com/cosmos/cosmos-sdk/x/evidence/types"
	paramstypes "github.com/cosmos/cosmos-sdk/x/params/types"
	upgradetypes "github.com/cosmos/cosmos-sdk/x/upgrade/types"
	"github.com/umee-network/umee/v6/app/inflation"
	appparams "github.com/umee-network/umee/v6/app/params"
	"github.com/umee-network/umee/v6/x/incentive"
	incentivemodule "github.com/umee-network/umee/v6/x/incentive/module"
	"github.com/umee-network/umee/v6/x/leverage"
	leveragetypes "github.com/umee-network/umee/v6/x/leverage/types"
	"github.com/umee-network/umee/v6/x/metoken"
	metokenmodule "github.com/umee-network/umee/v6/x/metoken/module"
	"github.com/umee-network/umee/v6/x/oracle"
	oracletypes "github.com/umee-network/umee/v6/x/oracle/types"
	"github.com/umee-network/umee/v6/x/ugov"
	ugovmodule "github.com/umee-network/umee/v6/x/ugov/module"
	"github.com/umee-network/umee/v6/x/uibc"
	uibcmodule "github.com/umee-network/umee/v6/x/uibc/module"
)

// module account permissions
var maccPerms = map[string][]string{
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

	incentive.ModuleName:   nil,
	oracletypes.ModuleName: nil,
	uibc.ModuleName:        nil,
	ugov.ModuleName:        nil,
	metoken.ModuleName:     {authtypes.Minter, authtypes.Burner},
}

// ModuleBasics defines the module BasicManager is in charge of setting up basic,
// non-dependant module elements, such as codec registration
// and genesis verification.
var ModuleBasics = module.NewBasicManager(
	auth.AppModuleBasic{},
	genutil.NewAppModuleBasic(genutiltypes.DefaultMessageValidator),
	BankModule{},
	capability.AppModuleBasic{},
	StakingModule{},
	MintModule{},
	distr.AppModuleBasic{},
	gov.NewAppModuleBasic(getGovProposalHandlers()),
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
	incentivemodule.AppModuleBasic{},
	metokenmodule.AppModuleBasic{},
	packetforward.AppModuleBasic{},
)

// NOTE: Any module instantiated in the module manager that is later modified
// must be passed by reference here.
func appModules(
	app *UmeeApp,
	appCodec codec.Codec,
	txConfig client.TxConfig,
	skipGenesisInvariants bool,
	inflationCalculator inflation.Calculator,
) []module.AppModule {
	// if Experimental {}

	return []module.AppModule{
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
		crisis.NewAppModule(app.CrisisKeeper, skipGenesisInvariants, app.GetSubspace(crisistypes.ModuleName)),
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
	}
}

// orderBeginBlockers tell the app's module manager how to set the order of
// BeginBlockers, which are run at the beginning of every block.
//
// During begin block slashing happens after distr.BeginBlocker so that
// there is nothing left over in the validator fee pool, so as to keep the
// CanWithdrawInvariant invariant.
// NOTE: staking module is required if HistoricalEntries param > 0
// NOTE: capability module's beginblocker must come before any modules using capabilities (e.g. IBC)
func orderBeginBlockers() []string {
	return []string{
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
		crisistypes.ModuleName,
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
	}
}

func orderEndBlockers() []string {
	return []string{
		crisistypes.ModuleName,
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
	}
}

// NOTE: The genutils module must occur after staking so that pools are
// properly initialized with tokens from genesis accounts.
// NOTE: The genutils module must also occur after auth so that it can access the params from auth.
// NOTE: Capability module must occur first so that it can initialize any capabilities
// so that other modules that want to create or claim capabilities afterwards in InitChain
// can do so safely.
func orderInitBlockers() []string {
	return []string{
		capabilitytypes.ModuleName, authtypes.ModuleName, banktypes.ModuleName, distrtypes.ModuleName,
		stakingtypes.ModuleName, slashingtypes.ModuleName, govtypes.ModuleName, minttypes.ModuleName,
		crisistypes.ModuleName, ibcexported.ModuleName, genutiltypes.ModuleName, evidencetypes.ModuleName,
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
	}
}

func orderMigrations() []string {
	return []string{
		capabilitytypes.ModuleName, authtypes.ModuleName, banktypes.ModuleName, distrtypes.ModuleName,
		stakingtypes.ModuleName, slashingtypes.ModuleName, govtypes.ModuleName, minttypes.ModuleName,
		crisistypes.ModuleName, ibcexported.ModuleName, genutiltypes.ModuleName, evidencetypes.ModuleName,
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
	}
}

// Cosmos SDK module wrappers

// BankModule defines a custom wrapper around the x/bank module's AppModuleBasic
// implementation to provide custom default genesis state.
type BankModule struct {
	bank.AppModuleBasic
}

// DefaultGenesis returns custom Umee x/bank module genesis state.
func (BankModule) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	umeeMetadata := banktypes.Metadata{
		Description: "The native staking token of the Umee network.",
		Base:        appparams.BondDenom,
		Name:        appparams.DisplayDenom,
		Display:     appparams.DisplayDenom,
		Symbol:      appparams.DisplayDenom,
		DenomUnits: []*banktypes.DenomUnit{
			{
				Denom:    appparams.BondDenom,
				Exponent: 0,
				Aliases: []string{
					"microumee",
				},
			},
			{
				Denom:    appparams.DisplayDenom,
				Exponent: 6,
				Aliases:  []string{},
			},
		},
	}

	genState := banktypes.DefaultGenesisState()
	genState.DenomMetadata = append(genState.DenomMetadata, umeeMetadata)

	return cdc.MustMarshalJSON(genState)
}

// StakingModule defines a custom wrapper around the x/staking module's
// AppModuleBasic implementation to provide custom default genesis state.
type StakingModule struct {
	staking.AppModuleBasic
}

// DefaultGenesis returns custom Umee x/staking module genesis state.
func (StakingModule) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	p := stakingtypes.DefaultParams()
	p.BondDenom = appparams.BondDenom
	return cdc.MustMarshalJSON(&stakingtypes.GenesisState{
		Params: p,
	})
}

// CrisisModule defines a custom wrapper around the x/crisis module's
// AppModuleBasic implementation to provide custom default genesis state.
type CrisisModule struct {
	crisis.AppModuleBasic
}

// DefaultGenesis returns custom Umee x/crisis module genesis state.
func (CrisisModule) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	return cdc.MustMarshalJSON(&crisistypes.GenesisState{
		ConstantFee: sdk.NewCoin(appparams.BondDenom, sdk.NewInt(1000)),
	})
}

// MintModule defines a custom wrapper around the x/mint module's
// AppModuleBasic implementation to provide custom default genesis state.
type MintModule struct {
	mint.AppModuleBasic
}

// DefaultGenesis returns custom Umee x/mint module genesis state.
func (MintModule) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	genState := minttypes.DefaultGenesisState()
	genState.Params.MintDenom = appparams.BondDenom

	return cdc.MustMarshalJSON(genState)
}

// GovModule defines a custom wrapper around the x/gov module's
// AppModuleBasic implementation to provide custom default genesis state.
type GovModule struct {
	gov.AppModuleBasic
}

// DefaultGenesis returns custom Umee x/gov module genesis state.
func (GovModule) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	minDeposit := sdk.NewCoins(sdk.NewCoin(appparams.BondDenom, govv1.DefaultMinDepositTokens))
	genState := govv1.DefaultGenesisState()
	genState.Params.MinDeposit = minDeposit

	return cdc.MustMarshalJSON(genState)
}

// SlashingModule defines a custom wrapper around the x/slashing module's
// AppModuleBasic implementation to provide custom default genesis state.
type SlashingModule struct {
	slashing.AppModuleBasic
}

// DefaultGenesis returns custom Umee x/slashing module genesis state.
func (SlashingModule) DefaultGenesis(cdc codec.JSONCodec) json.RawMessage {
	genState := slashingtypes.DefaultGenesisState()
	genState.Params.SignedBlocksWindow = 10000
	genState.Params.DowntimeJailDuration = 24 * time.Hour

	return cdc.MustMarshalJSON(genState)
}

func GenTxValidator(msgs []sdk.Msg) error {
	if n := len(msgs); n != 1 {
		return fmt.Errorf(
			"contains invalid number of messages; expected: 2 or 1; got: %d", n)
	}

	if err := assertMsgType[*stakingtypes.MsgCreateValidator](msgs[0], 0); err != nil {
		return err
	}

	for i := range msgs {
		if err := msgs[i].ValidateBasic(); err != nil {
			return fmt.Errorf("invalid GenTx msg[%d] '%s': %s", i, msgs[i], err)
		}
	}
	return nil
}

func assertMsgType[T sdk.Msg](m sdk.Msg, idx int) error {
	if _, ok := m.(T); !ok {
		var t T
		return fmt.Errorf(
			"contains invalid message at index %d; expected: %T; got: %T",
			idx, t, m)
	}
	return nil
}
