package uibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/umee-network/umee/v4/x/uibc/quota/keeper"
)

// BeginBlock implements BeginBlock for the x/uibc module.
func BeginBlock(ctx sdk.Context, keeper keeper.Keeper) {
	quotaExpires, err := keeper.GetExpire(ctx)
	if err != nil {
		panic(err)
	}

	// reset quotas
	if quotaExpires == nil || quotaExpires.Before(ctx.BlockTime()) {
		if err := keeper.ResetQuota(ctx); err != nil {
			panic(err)
		}
	}
}

// EndBlocker implements EndBlock for the x/leverage module.
func EndBlocker(_ sdk.Context, _ keeper.Keeper) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}
