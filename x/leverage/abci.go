package leverage

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	govkeeper "github.com/cosmos/cosmos-sdk/x/gov/keeper"

	"github.com/umee-network/umee/v3/x/leverage/keeper"
	"github.com/umee-network/umee/v3/x/leverage/migrations/mv2"
)

// EndBlocker implements EndBlock for the x/leverage module.
func EndBlocker(ctx sdk.Context, k keeper.Keeper, gk govkeeper.Keeper) []abci.ValidatorUpdate {
	if err := k.SweepBadDebts(ctx); err != nil {
		panic(err)
	}

	if err := k.AccrueAllInterest(ctx); err != nil {
		panic(err)
	}

	mv2.Migrator1to2(gk)(ctx)

	return []abci.ValidatorUpdate{}
}
