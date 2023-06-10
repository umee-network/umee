package refileverage

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/umee-network/umee/v5/util"
	"github.com/umee-network/umee/v5/x/refileverage/keeper"
)

// EndBlocker implements EndBlock for the x/refileverage module.
func EndBlocker(ctx sdk.Context, k keeper.Keeper) []abci.ValidatorUpdate {
	util.Panic(k.SweepBadDebts(ctx))
	util.Panic(k.AccrueAllInterest(ctx))

	return []abci.ValidatorUpdate{}
}
