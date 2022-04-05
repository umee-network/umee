package leverage

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/umee-network/umee/v2/x/leverage/keeper"
)

// EndBlocker implements EndBlock for the x/leverage module.
func EndBlocker(ctx sdk.Context, k keeper.Keeper) []abci.ValidatorUpdate {
	if err := k.SweepBadDebts(ctx); err != nil {
		panic(err)
	}

	if err := k.AccrueAllInterest(ctx); err != nil {
		panic(err)
	}

	return []abci.ValidatorUpdate{}
}
