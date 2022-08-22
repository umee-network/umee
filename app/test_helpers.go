package app

import (
	"encoding/json"
	"fmt"
	"testing"
	"time"

	"github.com/cosmos/cosmos-sdk/baseapp"
	"github.com/cosmos/cosmos-sdk/codec"
	codectypes "github.com/cosmos/cosmos-sdk/codec/types"
	cryptocodec "github.com/cosmos/cosmos-sdk/crypto/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	pruningtypes "github.com/cosmos/cosmos-sdk/pruning/types"
	servertypes "github.com/cosmos/cosmos-sdk/server/types"
	"github.com/cosmos/cosmos-sdk/testutil/mock"
	"github.com/cosmos/cosmos-sdk/testutil/network"
	sdk "github.com/cosmos/cosmos-sdk/types"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	govtypes "github.com/cosmos/cosmos-sdk/x/gov/types"
	govv1 "github.com/cosmos/cosmos-sdk/x/gov/types/v1"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
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

	privVal := mock.NewPV()
	pubKey, err := privVal.GetPubKey()
	require.NoError(t, err)

	// create validator set with single validator
	validator := tmtypes.NewValidator(pubKey, 1)
	valSet := tmtypes.NewValidatorSet([]*tmtypes.Validator{validator})

	// generate genesis account
	senderPrivKey := secp256k1.GenPrivKey()
	acc := authtypes.NewBaseAccount(senderPrivKey.PubKey().Address().Bytes(), senderPrivKey.PubKey(), 0, 0)
	balance := banktypes.Balance{
		Address: acc.GetAddress().String(),
		Coins:   sdk.NewCoins(sdk.NewCoin(BondDenom, sdk.NewInt(10000000000000000))),
	}

	app := SetupWithGenesisValSet(t, valSet, []authtypes.GenesisAccount{acc}, balance)
	return app
}

// SetupWithGenesisValSet initializes a new app with a validator set and genesis accounts
// that also act as delegators. For simplicity, each validator is bonded with a delegation
// of one consensus engine unit in the default token of the simapp from first genesis
// account. A Nop logger is set in app.
func SetupWithGenesisValSet(
	t *testing.T,
	valSet *tmtypes.ValidatorSet,
	genAccs []authtypes.GenesisAccount,
	balances ...banktypes.Balance,
) *UmeeApp {
	t.Helper()

	app, genesisState := setup(true, 5)
	genesisState, err := GenesisStateWithValSet(app.AppCodec(), genesisState, valSet, genAccs, balances...)
	require.NoError(t, err)

	stateBytes, err := json.MarshalIndent(genesisState, "", " ")
	require.NoError(t, err)

	// init chain will set the validator set and initialize the genesis accounts
	app.InitChain(
		abci.RequestInitChain{
			Validators:      []abci.ValidatorUpdate{},
			ConsensusParams: DefaultConsensusParams,
			AppStateBytes:   stateBytes,
		},
	)

	// commit genesis changes
	app.Commit()
	app.BeginBlock(abci.RequestBeginBlock{Header: tmproto.Header{
		Height:             app.LastBlockHeight() + 1,
		AppHash:            app.LastCommitID().Hash,
		ValidatorsHash:     valSet.Hash(),
		NextValidatorsHash: valSet.Hash(),
	}})

	return app
}

// GenesisStateWithValSet returns a new genesis state with the validator set
func GenesisStateWithValSet(codec codec.Codec, genesisState map[string]json.RawMessage,
	valSet *tmtypes.ValidatorSet, genAccs []authtypes.GenesisAccount,
	balances ...banktypes.Balance,
) (map[string]json.RawMessage, error) {
	// set genesis accounts
	authGenesis := authtypes.NewGenesisState(authtypes.DefaultParams(), genAccs)
	genesisState[authtypes.ModuleName] = codec.MustMarshalJSON(authGenesis)

	validators := make([]stakingtypes.Validator, 0, len(valSet.Validators))
	delegations := make([]stakingtypes.Delegation, 0, len(valSet.Validators))

	bondAmt := sdk.DefaultPowerReduction

	for _, val := range valSet.Validators {
		pk, err := cryptocodec.FromTmPubKeyInterface(val.PubKey)
		if err != nil {
			return nil, fmt.Errorf("failed to convert pubkey: %w", err)
		}

		pkAny, err := codectypes.NewAnyWithValue(pk)
		if err != nil {
			return nil, fmt.Errorf("failed to create new any: %w", err)
		}

		validator := stakingtypes.Validator{
			OperatorAddress:   sdk.ValAddress(val.Address).String(),
			ConsensusPubkey:   pkAny,
			Jailed:            false,
			Status:            stakingtypes.Bonded,
			Tokens:            bondAmt,
			DelegatorShares:   sdk.OneDec(),
			Description:       stakingtypes.Description{},
			UnbondingHeight:   int64(0),
			UnbondingTime:     time.Unix(0, 0).UTC(),
			Commission:        stakingtypes.NewCommission(sdk.ZeroDec(), sdk.ZeroDec(), sdk.ZeroDec()),
			MinSelfDelegation: sdk.ZeroInt(),
		}
		validators = append(validators, validator)
		newDel := stakingtypes.NewDelegation(genAccs[0].GetAddress(), val.Address.Bytes(), sdk.OneDec())
		delegations = append(delegations, newDel)

	}

	defaultStParams := stakingtypes.DefaultParams()
	stParams := stakingtypes.NewParams(
		defaultStParams.UnbondingTime,
		defaultStParams.MaxValidators,
		defaultStParams.MaxEntries,
		defaultStParams.HistoricalEntries,
		BondDenom,
		defaultStParams.MinCommissionRate,
	)
	// set validators and delegations
	stakingGenesis := stakingtypes.NewGenesisState(stParams, validators, delegations)
	genesisState[stakingtypes.ModuleName] = codec.MustMarshalJSON(stakingGenesis)

	totalSupply := sdk.NewCoins()
	for _, b := range balances {
		// add genesis acc tokens to total supply
		totalSupply = totalSupply.Add(b.Coins...)
	}

	for range delegations {
		// add delegated tokens to total supply
		totalSupply = totalSupply.Add(sdk.NewCoin(BondDenom, bondAmt))
	}

	// add bonded amount to bonded pool module account
	balances = append(balances, banktypes.Balance{
		Address: authtypes.NewModuleAddress(stakingtypes.BondedPoolName).String(),
		Coins:   sdk.Coins{sdk.NewCoin(BondDenom, bondAmt)},
	})

	// update total supply
	bankGenesis := banktypes.NewGenesisState(
		banktypes.DefaultGenesisState().Params,
		balances,
		totalSupply,
		[]banktypes.Metadata{},
	)
	genesisState[banktypes.ModuleName] = codec.MustMarshalJSON(bankGenesis)

	return genesisState, nil
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

	// Default genesis doesn't have tokens in Token registry, so here we add one:
	var leverageGenState leveragetypes.GenesisState
	if err := cdc.UnmarshalJSON(appGenState[leveragetypes.ModuleName], &leverageGenState); err != nil {
		panic(err)
	}
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
		MaxSupply:              sdk.NewInt(100000000000),
	})
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
