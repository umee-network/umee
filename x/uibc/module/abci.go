package uibc

import (
	"time"

	abci "github.com/cometbft/cometbft/abci/types"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v6/x/uibc/quota"
)

// BeginBlock implements BeginBlock for the x/uibc module.
func BeginBlock(ctx sdk.Context, k quota.Keeper) {
	logger := ctx.Logger().With("module", "uibc")
	quotaExpires, err := k.GetExpire()
	if err != nil {
		logger.Error("can't get quota expire", "error", err)
		return
	}

	// reset quotas
	if quotaExpires == nil || quotaExpires.Before(ctx.BlockTime()) {
		if err = k.ResetAllQuotas(); err != nil {
			logger.Error("can't get quota expire", "error", err)
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
