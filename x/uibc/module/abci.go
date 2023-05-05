package uibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/umee-network/umee/v4/x/uibc/quota/keeper"
)

// BeginBlock implements BeginBlock for the x/uibc module.
func BeginBlock(ctx sdk.Context, k keeper.Keeper) {
	quotaExpires, err := k.GetExpire()
	if err != nil {
		// TODO, use logger as argument
		ctx.Logger().Error("can't get quota exipre", "error", err)
		return
	}

	// reset quotas
	if quotaExpires == nil || quotaExpires.Before(ctx.BlockTime()) {
		if err = k.ResetAllQuotas(); err != nil {
			ctx.Logger().Error("can't get quota exipre", "error", err)
		}
	}
}

// EndBlocker implements EndBlock.
func EndBlocker() []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}
