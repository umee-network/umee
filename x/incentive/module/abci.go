package module

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/umee-network/umee/v4/util"
	"github.com/umee-network/umee/v4/x/incentive/keeper"
)

// EndBlocker implements EndBlock for the x/incentive module.
func EndBlocker(ctx sdk.Context, k keeper.Keeper) []abci.ValidatorUpdate {
	util.Panic(k.EndBlock(ctx))
	return []abci.ValidatorUpdate{}
}
