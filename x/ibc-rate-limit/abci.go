package ibc_rate_limit

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/umee-network/umee/v3/x/ibc-rate-limit/keeper"
)

// BeginBlock implements BeginBlock for the x/ibc-rate-limit module.
func BeginBlock(ctx sdk.Context, keeper keeper.Keeper) {
	rateLimitsOfIBCDenoms, err := keeper.GetRateLimitsOfIBCDenoms(ctx)
	if err != nil {
		panic(err)
	}

	for _, rateLimitOfIBCDenom := range rateLimitsOfIBCDenoms {
		if rateLimitOfIBCDenom.ExpiredTime == nil {
			expiredTime := ctx.BlockTime().Add(rateLimitOfIBCDenom.TimeWindow)
			rateLimitOfIBCDenom.ExpiredTime = &expiredTime
		} else {
			if rateLimitOfIBCDenom.ExpiredTime.Before(ctx.BlockTime()) {
				// reset the expire time
				expiredTime := ctx.BlockTime().Add(rateLimitOfIBCDenom.TimeWindow)
				rateLimitOfIBCDenom.ExpiredTime = &expiredTime
				// reset the inflow limit to 0
				rateLimitOfIBCDenom.InflowSum = 0
				// reset the outflow limit to 0
				rateLimitOfIBCDenom.OutflowSum = 0
			}
		}
		// storing the rate limits to store
		keeper.SetRateLimitsOfIBCDenom(ctx, rateLimitOfIBCDenom)
	}
}

// EndBlocker implements EndBlock for the x/leverage module.
func EndBlocker(ctx sdk.Context, k keeper.Keeper) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}
