package simulation

import (
	"encoding/json"
	"fmt"
	"io"
	"io/ioutil"
	"math/rand"
	"time"

	sdkmath "cosmossdk.io/math"
	gravitytypes "github.com/Gravity-Bridge/Gravity-Bridge/module/x/gravity/types"
	"github.com/cosmos/cosmos-sdk/codec"
	"github.com/cosmos/cosmos-sdk/crypto/keys/secp256k1"
	"github.com/cosmos/cosmos-sdk/simapp"
	simappparams "github.com/cosmos/cosmos-sdk/simapp/params"
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	authtypes "github.com/cosmos/cosmos-sdk/x/auth/types"
	banktypes "github.com/cosmos/cosmos-sdk/x/bank/types"
	stakingtypes "github.com/cosmos/cosmos-sdk/x/staking/types"
	tmjson "github.com/tendermint/tendermint/libs/json"
	tmtypes "github.com/tendermint/tendermint/types"

	umeeapp "github.com/umee-network/umee/v2/app"
	"github.com/umee-network/umee/v2/tests/util"
)

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

		gravityStateBz, ok := rawState[gravitytypes.ModuleName]
		if !ok {
			panic("gravity genesis state is missing")
		}

		gravityState := new(gravitytypes.GenesisState)
		if err := cdc.UnmarshalJSON(gravityStateBz, gravityState); err != nil {
			panic(err)
		}

		delegateKeys := make([]gravitytypes.MsgSetOrchestratorAddress, 0, len(stakingState.Validators))
		for _, val := range stakingState.Validators {
			if val.Status != stakingtypes.Bonded {
				_, _, ethAddr, err := util.GenerateRandomEthKeyFromRand(r)
				if err != nil {
					panic(err)
				}

				gravityEthAddr, err := gravitytypes.NewEthAddress(ethAddr.Hex())
				if err != nil {
					panic(err)
				}

				delegateKeys = append(delegateKeys, *gravitytypes.NewMsgSetOrchestratorAddress(
					val.GetOperator(),
					sdk.AccAddress(val.GetOperator()),
					*gravityEthAddr,
				))
			}
		}

		gravityState.DelegateKeys = delegateKeys

		rawState[stakingtypes.ModuleName] = cdc.MustMarshalJSON(stakingState)
		rawState[banktypes.ModuleName] = cdc.MustMarshalJSON(bankState)
		rawState[gravitytypes.ModuleName] = cdc.MustMarshalJSON(gravityState)

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
