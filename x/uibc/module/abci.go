package uibc

import (
	"time"

	abci "github.com/cometbft/cometbft/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v6/x/uibc/quota/keeper"
)

// BeginBlock implements BeginBlock for the x/uibc module.
func BeginBlock(ctx sdk.Context, k keeper.Keeper) {
	logger := ctx.Logger().With("module", "uibc")
	quotaExpires, err := k.GetExpire()
	if err != nil {
		// TODO, use logger as argument
		logger.Error("can't get quota exipre", "error", err)
		return
	}

	// reset quotas
	if quotaExpires == nil || quotaExpires.Before(ctx.BlockTime()) {
		if err = k.ResetAllQuotas(); err != nil {
			logger.Error("can't get quota exipre", "error", err)
		} else {
			logger.Info("IBC Quota Reset")
			ctx.EventManager().EmitEvent(
				sdk.NewEvent("/umee/uibc/v1/EventQuotaReset",
					sdk.NewAttribute("next_expire", quotaExpires.UTC().Format(time.RFC3339))))
		}
	}
}

// EndBlocker implements EndBlock.
func EndBlocker() []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}
