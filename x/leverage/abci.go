package leverage

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"

	"github.com/umee-network/umee/x/leverage/keeper"
)

// EndBlocker implements EndBlock for the x/leverage module.
func EndBlocker(ctx sdk.Context, k keeper.Keeper) []abci.ValidatorUpdate {
	height := ctx.BlockHeight()
	epoch := k.GetParams(ctx).InterestEpoch

	if height%epoch == 0 {
		if err := k.SweepBadDebts(ctx); err != nil {
			panic(err)
		}

		if err := k.AccrueAllInterest(ctx); err != nil {
			panic(err)
		}

		if err := k.UpdateExchangeRates(ctx); err != nil {
			panic(err)
		}
	}

	return []abci.ValidatorUpdate{}
}
