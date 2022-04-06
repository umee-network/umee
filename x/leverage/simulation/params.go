package simulation

import (
	"fmt"
	"math/rand"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
	"github.com/cosmos/cosmos-sdk/x/simulation"

	"github.com/umee-network/umee/v2/x/leverage/types"
)

// ParamChanges defines the parameters that can be modified by param change proposals
// on the simulation
func ParamChanges(r *rand.Rand) []simtypes.ParamChange {
	return []simtypes.ParamChange{
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyCompleteLiquidationThreshold),
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenCompleteLiquidationThreshold(r))
			},
		),
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyMinimumCloseFactor),
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenMinimumCloseFactor(r))
			},
		),
		simulation.NewSimParamChange(types.ModuleName, string(types.KeyOracleRewardFactor),
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenOracleRewardFactor(r))
			},
		),
		simulation.NewSimParamChange(types.ModuleName, string(types.KeySmallLiquidationSize),
			func(r *rand.Rand) string {
				return fmt.Sprintf("\"%s\"", GenSmallLiquidationSize(r))
			},
		),
	}
}
