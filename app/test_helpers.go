package app

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	pruningtypes "github.com/cosmos/cosmos-sdk/pruning/types"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	"github.com/stretchr/testify/require"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/tendermint/tendermint/libs/log"
	tmproto "github.com/tendermint/tendermint/proto/tendermint/types"
	tmtypes "github.com/tendermint/tendermint/types"
	dbm "github.com/tendermint/tm-db"

	leveragetypes "github.com/umee-network/umee/v2/x/leverage/types"
	oracletypes "github.com/umee-network/umee/v2/x/oracle/types"
)

// DefaultConsensusParams defines the default Tendermint consensus params used
// in UmeeApp testing.
var DefaultConsensusParams = &abci.ConsensusParams{
	Block: &abci.BlockParams{
		MaxBytes: 200000,
		MaxGas:   2000000,
	},
	Evidence: &tmproto.EvidenceParams{
		MaxAgeNumBlocks: 302400,
		MaxAgeDuration:  504 * time.Hour, // 3 weeks is the max duration
		MaxBytes:        10000,
	},
	Validator: &tmproto.ValidatorParams{
		PubKeyTypes: []string{
			tmtypes.ABCIPubKeyTypeEd25519,
		},
	},
}

type EmptyAppOptions struct{}

func (EmptyAppOptions) Get(o string) interface{} { return nil }

func Setup(t *testing.T, isCheckTx bool, invCheckPeriod uint) *UmeeApp {
	t.Helper()

	app, genesisState := setup(!isCheckTx, invCheckPeriod)
	if !isCheckTx {
		// InitChain must be called to stop deliverState from being nil
		stateBytes, err := json.MarshalIndent(genesisState, "", " ")
		require.NoError(t, err)

		// Initialize the chain
		app.InitChain(
			abci.RequestInitChain{
				Validators:      []abci.ValidatorUpdate{},
				ConsensusParams: DefaultConsensusParams,
				AppStateBytes:   stateBytes,
			},
		)
	}

	return app
}

func setup(withGenesis bool, invCheckPeriod uint) (*UmeeApp, GenesisState) {
	db := dbm.NewMemDB()
	encCdc := MakeEncodingConfig()
	app := New(
		log.NewNopLogger(),
		db,
		nil,
		true,
		map[int64]bool{},
		DefaultNodeHome,
		invCheckPeriod,
		encCdc,
		EmptyAppOptions{},
	)
	if withGenesis {
		return app, NewDefaultGenesisState(encCdc.Codec)
	}

	return app, GenesisState{}
}

// IntegrationTestNetworkConfig returns a networking configuration used for
// integration tests using the SDK's in-process network test suite.
func IntegrationTestNetworkConfig() network.Config {
	cfg := network.DefaultConfig()
	encCfg := MakeEncodingConfig()
	cdc := encCfg.Codec

	// Start with the default genesis state
	appGenState := ModuleBasics.DefaultGenesis(encCfg.Codec)

	// Extract the x/leverage part of the genesis state to be modified
	var leverageGenState leveragetypes.GenesisState
	if err := cdc.UnmarshalJSON(appGenState[leveragetypes.ModuleName], &leverageGenState); err != nil {
		panic(err)
	}

	// Modify the x/leverage genesis state
	leverageGenState.Registry = append(leverageGenState.Registry, leveragetypes.Token{
		BaseDenom:              BondDenom,
		SymbolDenom:            DisplayDenom,
		Exponent:               6,
		ReserveFactor:          sdk.MustNewDecFromStr("0.1"),
		CollateralWeight:       sdk.MustNewDecFromStr("0.05"),
		LiquidationThreshold:   sdk.MustNewDecFromStr("0.05"),
		BaseBorrowRate:         sdk.MustNewDecFromStr("0.02"),
		KinkBorrowRate:         sdk.MustNewDecFromStr("0.2"),
		MaxBorrowRate:          sdk.MustNewDecFromStr("1.5"),
		KinkUtilization:        sdk.MustNewDecFromStr("0.2"),
		LiquidationIncentive:   sdk.MustNewDecFromStr("0.18"),
		EnableMsgSupply:        true,
		EnableMsgBorrow:        true,
		Blacklist:              false,
		MaxCollateralShare:     sdk.MustNewDecFromStr("1"),
		MaxSupplyUtilization:   sdk.MustNewDecFromStr("1"),
		MinCollateralLiquidity: sdk.MustNewDecFromStr("0"),
	})

	// Marshal the modified state and add it back into appGenState
	bz, err := cdc.MarshalJSON(&leverageGenState)
	if err != nil {
		panic(err)
	}
	appGenState[leveragetypes.ModuleName] = bz

	var oracleGenState oracletypes.GenesisState
	if err := cdc.UnmarshalJSON(appGenState[oracletypes.ModuleName], &oracleGenState); err != nil {
		panic(err)
	}

	// Set mock exchange rates and a large enough vote period such that we won't
	// execute ballot voting and thus clear out previous exchange rates, since we
	// are not running a price-feeder.
	oracleGenState.Params.VotePeriod = 1000
	oracleGenState.ExchangeRates = append(oracleGenState.ExchangeRates, oracletypes.NewExchangeRateTuple(
		DisplayDenom, sdk.MustNewDecFromStr("34.21"),
	))

	bz, err = cdc.MarshalJSON(&oracleGenState)
	if err != nil {
		panic(err)
	}
	appGenState[oracletypes.ModuleName] = bz

	var govGenState govv1.GenesisState
	if err := cdc.UnmarshalJSON(appGenState[govtypes.ModuleName], &govGenState); err != nil {
		panic(err)
	}

	votingPeriod := time.Minute
	govGenState.VotingParams.VotingPeriod = &votingPeriod

	bz, err = cdc.MarshalJSON(&govGenState)
	if err != nil {
		panic(err)
	}
	appGenState[govtypes.ModuleName] = bz

	cfg.Codec = encCfg.Codec
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
			EmptyAppOptions{},
			baseapp.SetPruning(pruningtypes.NewPruningOptionsFromString(val.AppConfig.Pruning)),
			baseapp.SetMinGasPrices(val.AppConfig.MinGasPrices),
		)
	}

	return cfg
}
