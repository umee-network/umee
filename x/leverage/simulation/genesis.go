package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/umee-network/umee/v2/x/leverage/types"
)

// Simulation parameter constants
const (
	completeLiquidationThresholdKey = "complete_liquidation_threshold"
	minimumCloseFactorKey           = "minimum_close_factor"
	oracleRewardFactorKey           = "oracle_reward_factor"
	smallLiquidationSizeKey         = "small_liquidation_size"
)

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

// GenSmallLiquidationSize produces a randomized SmallLiquidationSize in the range of [0, 1000]
func GenSmallLiquidationSize(r *rand.Rand) sdk.Dec {
	return sdk.NewDec(int64(r.Intn(1000)))
}

// RandomizedGenState generates a random GenesisState for oracle
func RandomizedGenState(simState *module.SimulationState) {
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

	var smallLiquidationSize sdk.Dec
	simState.AppParams.GetOrGenerate(
		simState.Cdc, smallLiquidationSizeKey, &smallLiquidationSize, simState.Rand,
		func(r *rand.Rand) { smallLiquidationSize = GenSmallLiquidationSize(r) },
	)

	leverageGenesis := types.NewGenesisState(
		types.Params{
			CompleteLiquidationThreshold: completeLiquidationThreshold,
			MinimumCloseFactor:           minimumCloseFactor,
			OracleRewardFactor:           oracleRewardFactor,
			SmallLiquidationSize:         smallLiquidationSize,
		},
		[]types.Token{},
		[]types.AdjustedBorrow{},
		[]types.CollateralSetting{},
		[]types.Collateral{},
		sdk.Coins{},
		0,
		[]types.BadDebt{},
		[]types.InterestScalar{},
		sdk.Coins{},
	)

	bz, err := json.MarshalIndent(&leverageGenesis.Params, "", " ")
	if err != nil {
		panic(err)
	}
	fmt.Printf("Selected randomly generated leverage parameters:\n%s\n", bz)
	simState.GenState[types.ModuleName] = simState.Cdc.MustMarshalJSON(leverageGenesis)
}
