package ibctransfer

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/umee-network/umee/v3/x/ibctransfer/ratelimits/keeper"
)

// BeginBlock implements BeginBlock for the x/ibc-rate-limit module.
func BeginBlock(ctx sdk.Context, keeper keeper.Keeper) {
	rateLimitsOfIBCDenoms, err := keeper.GetRateLimitsOfIBCDenoms(ctx)
	if err != nil {
		panic(err)
	}

	params := keeper.GetParams(ctx)

	for _, rateLimitOfIBCDenom := range rateLimitsOfIBCDenoms {
		if rateLimitOfIBCDenom.ExpiredTime == nil {
			expiredTime := ctx.BlockTime().Add(params.QuotaDuration)
			rateLimitOfIBCDenom.ExpiredTime = &expiredTime
		} else {
			if rateLimitOfIBCDenom.ExpiredTime.Before(ctx.BlockTime()) {
				// reset the expire time
				expiredTime := ctx.BlockTime().Add(params.QuotaDuration)
				rateLimitOfIBCDenom.ExpiredTime = &expiredTime
				// reset the outflow limit to 0
				rateLimitOfIBCDenom.OutflowSum = sdk.NewDec(0)
			}
		}
		// storing the rate limits to store
		if err := keeper.SetRateLimitsOfIBCDenom(ctx, &rateLimitOfIBCDenom); err != nil {
			panic(err)
		}
	}
}

// EndBlocker implements EndBlock for the x/leverage module.
func EndBlocker(_ sdk.Context, _ keeper.Keeper) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}
