package simulation

import (
	"math/rand"

	simtypes "github.com/cosmos/cosmos-sdk/types/simulation"
)

// ParamChanges defines the parameters that can be modified by param change proposals
// on the simulation
func ParamChanges(*rand.Rand) []simtypes.LegacyParamChange {
	return []simtypes.LegacyParamChange{
		// empty: leverage params are in regular state
	}
}
