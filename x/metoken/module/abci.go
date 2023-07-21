package module

import (
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/umee-network/umee/v5/x/metoken/keeper"
)

// EndBlocker implements EndBlock for the x/metoken module.
func EndBlocker(_ keeper.Keeper) []abci.ValidatorUpdate {
	// todo: add reserves re-balancing
	// todo: add interest claiming
	return []abci.ValidatorUpdate{}
}
