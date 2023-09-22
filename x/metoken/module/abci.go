package module

import (
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/umee-network/umee/v6/util"
	"github.com/umee-network/umee/v6/x/metoken/keeper"
)

// EndBlocker implements EndBlock for the x/metoken module.
func EndBlocker(k keeper.Keeper) []abci.ValidatorUpdate {
	util.Panic(k.ClaimLeverageInterest())
	util.Panic(k.RebalanceReserves())
	util.Panic(k.Bond())
	return []abci.ValidatorUpdate{}
}
