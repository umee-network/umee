package module

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/umee-network/umee/v4/util"
	"github.com/umee-network/umee/v4/x/incentive/keeper"
)

// EndBlocker implements EndBlock for the x/incentive module.
func EndBlocker(ctx sdk.Context, k keeper.Keeper) []abci.ValidatorUpdate {
	err, _ := k.EndBlock(ctx)
	util.Panic(err)
	return []abci.ValidatorUpdate{}
}
