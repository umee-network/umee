package simulation

import (
	"encoding/json"
	"fmt"
	"math/rand"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/cosmos/cosmos-sdk/types/module"
	"github.com/umee-network/umee/v2/x/leverage/types"
)

const atomIBCDenom = "ibc/CDC4587874B85BEA4FCEC3CEA5A1195139799A1FEE711A07D972537E18FDA39D"

// Simulation parameter constants
const (
	completeLiquidationThresholdKey = "complete_liquidation_threshold"
	minimumCloseFactorKey           = "minimum_close_factor"
	oracleRewardFactorKey           = "oracle_reward_factor"
	smallLiquidationSizeKey         = "small_liquidation_size"
)

// GenCompleteLiquidationThreshold produces a randomized CompleteLiquidationThreshold in the range of [0.050, 0.100]
func GenCompleteLiquidationThreshold(r *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(0o50, 3).Add(sdk.NewDecWithPrec(int64(r.Intn(950)), 3))
}

// GenMinimumCloseFactor produces a randomized MinimumCloseFactor in the range of [0.001, 0.047]
func GenMinimumCloseFactor(r *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(0o01, 3).Add(sdk.NewDecWithPrec(int64(r.Intn(0o46)), 3))
}

// GenOracleRewardFactor produces a randomized OracleRewardFactor in the range of [0.005, 0.100]
func GenOracleRewardFactor(r *rand.Rand) sdk.Dec {
	return sdk.NewDecWithPrec(0o05, 3).Add(sdk.NewDecWithPrec(int64(r.Intn(995)), 3))
}

// GenSmallLiquidationSize produces a randomized SmallLiquidationSize in the range of [0, 1000]
func GenSmallLiquidationSize(r *rand.Rand) sdk.Dec {
	return sdk.NewDec(int64(r.Intn(1000)))
}

// RandomizedGenState generates a random GenesisState for leverage
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
		simTokenRegistry(),
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

func simTokenRegistry() []types.Token {
	return []types.Token{
		{
			BaseDenom:            "uumee",
			ReserveFactor:        sdk.MustNewDecFromStr("0.20"),
			CollateralWeight:     sdk.MustNewDecFromStr("0.25"),
			LiquidationThreshold: sdk.MustNewDecFromStr("0.25"),
			BaseBorrowRate:       sdk.MustNewDecFromStr("0.02"),
			KinkBorrowRate:       sdk.MustNewDecFromStr("0.22"),
			MaxBorrowRate:        sdk.MustNewDecFromStr("1.52"),
			KinkUtilizationRate:  sdk.MustNewDecFromStr("0.8"),
			LiquidationIncentive: sdk.MustNewDecFromStr("0.1"),
			SymbolDenom:          "UMEE",
			Exponent:             6,
		},
		{
			BaseDenom:            atomIBCDenom,
			ReserveFactor:        sdk.MustNewDecFromStr("0.25"),
			CollateralWeight:     sdk.MustNewDecFromStr("0.5"),
			LiquidationThreshold: sdk.MustNewDecFromStr("0.5"),
			BaseBorrowRate:       sdk.MustNewDecFromStr("0.02"),
			KinkBorrowRate:       sdk.MustNewDecFromStr("0.2"),
			MaxBorrowRate:        sdk.MustNewDecFromStr("1.0"),
			KinkUtilizationRate:  sdk.MustNewDecFromStr("0.8"),
			LiquidationIncentive: sdk.MustNewDecFromStr("0.1"),
			SymbolDenom:          "ATOM",
			Exponent:             6,
		},
		{
			BaseDenom:            "uabcd",
			ReserveFactor:        sdk.MustNewDecFromStr("0.10"),
			CollateralWeight:     sdk.MustNewDecFromStr("0.05"),
			LiquidationThreshold: sdk.MustNewDecFromStr("0.25"),
			BaseBorrowRate:       sdk.MustNewDecFromStr("0.02"),
			KinkBorrowRate:       sdk.MustNewDecFromStr("0.22"),
			MaxBorrowRate:        sdk.MustNewDecFromStr("1.52"),
			KinkUtilizationRate:  sdk.MustNewDecFromStr("0.8"),
			LiquidationIncentive: sdk.MustNewDecFromStr("0.2"),
			SymbolDenom:          "ABCD",
			Exponent:             6,
		},
	}
}
