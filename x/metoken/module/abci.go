package module

import (
	abci "github.com/cometbft/cometbft/abci/types"

	"github.com/umee-network/umee/v5/util"
	"github.com/umee-network/umee/v5/x/metoken/keeper"
)

// EndBlocker implements EndBlock for the x/metoken module.
func EndBlocker(k keeper.Keeper) []abci.ValidatorUpdate {
	util.Panic(k.ClaimLeverageInterest())
	util.Panic(k.RebalanceReserves())
	return []abci.ValidatorUpdate{}
}
