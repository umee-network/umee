package leverage

import (
	abci "github.com/cometbft/cometbft/abci/types"
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v6/util"
	"github.com/umee-network/umee/v6/x/leverage/keeper"
)

// EndBlocker implements EndBlock for the x/leverage module.
func EndBlocker(ctx sdk.Context, k keeper.Keeper) []abci.ValidatorUpdate {
	util.Panic(k.SweepBadDebts(ctx))
	util.Panic(k.AccrueAllInterest(ctx))

	return []abci.ValidatorUpdate{}
}
