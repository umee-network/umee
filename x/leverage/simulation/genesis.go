package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/umee-network/umee/x/leverage/types"
)

// Simulation parameter constants
const (
	interestEpochKey                = "interest_epoch"
	completeLiquidationThresholdKey = "complete_liquidation_threshold"
	minimumCloseFactorKey           = "minimum_close_factor"
	oracleRewardFactorKey           = "oracle_reward_factor"
)

// GenInterestEpoch produces a randomized InterestEpoch in the range of [10, 100]
func GenInterestEpoch(r *rand.Rand) int64 {
	return int64(10 + r.Intn(90))
}

// GenCompleteLiquidationThreshold produces a randomized CompleteLiquidationThreshold in the range of [0.050, 0.100]
func GenCompleteLiquidationThreshold(r *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(050, 3).Add(sdk.NewDecWithPrec(int64(r.Intn(950)), 3))
}

// GenMinimumCloseFactor produces a randomized MinimumCloseFactor in the range of [0.001, 0.047]
func GenMinimumCloseFactor(r *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(001, 3).Add(sdk.NewDecWithPrec(int64(r.Intn(046)), 3))
}

// GenOracleRewardFactor produces a randomized OracleRewardFactor in the range of [0.005, 0.100]
func GenOracleRewardFactor(r *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(005, 3).Add(sdk.NewDecWithPrec(int64(r.Intn(995)), 3))
}

// RandomizedGenState generates a random GenesisState for oracle
func RandomizedGenState(simState *module.SimulationState) {
	var interestEpoch int64
	simState.AppParams.GetOrGenerate(
		simState.Cdc, interestEpochKey, &interestEpoch, simState.Rand,
		func(r *rand.Rand) { interestEpoch = GenInterestEpoch(r) },
	)

	var completeLiquidationThreshold sdk.Dec
	simState.AppParams.GetOrGenerate(
		simState.Cdc, completeLiquidationThresholdKey, &completeLiquidationThreshold, simState.Rand,
		func(r *rand.Rand) { completeLiquidationThreshold = GenCompleteLiquidationThreshold(r) },
	)

	var minimumCloseFactor sdk.Dec
	simState.AppParams.GetOrGenerate(
		simState.Cdc, minimumCloseFactorKey, &minimumCloseFactor, simState.Rand,
		func(r *rand.Rand) { minimumCloseFactor = GenMinimumCloseFactor(r) },
	)

	var oracleRewardFactor sdk.Dec
	simState.AppParams.GetOrGenerate(
		simState.Cdc, oracleRewardFactorKey, &oracleRewardFactor, simState.Rand,
		func(r *rand.Rand) { oracleRewardFactor = GenOracleRewardFactor(r) },
	)

	leverageGenesis := types.NewGenesisState(
		types.Params{
			InterestEpoch:                interestEpoch,
			CompleteLiquidationThreshold: completeLiquidationThreshold,
			MinimumCloseFactor:           minimumCloseFactor,
			OracleRewardFactor:           oracleRewardFactor,
		},
		[]types.Token{},
		[]types.Borrow{},
		[]types.CollateralSetting{},
		[]types.Collateral{},
		sdk.Coins{},
		0,
		[]types.ExchangeRate{},
		[]types.BadDebt{},
		[]types.APY{},
		[]types.APY{},
	)

	bz, err := json.MarshalIndent(&leverageGenesis.Params, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated leverage parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(leverageGenesis)
}
