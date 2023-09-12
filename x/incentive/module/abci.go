package module

import (
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v6/util"
	"github.com/umee-network/umee/v6/x/incentive/keeper"
)

// EndBlocker implements EndBlock for the x/incentive module.
func EndBlocker(ctx sdk.Context, k keeper.Keeper) []abci.ValidatorUpdate {
	_, err := k.EndBlock(ctx)
	util.Panic(err)
	return []abci.ValidatorUpdate{}
}
