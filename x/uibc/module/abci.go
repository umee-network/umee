package uibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/umee-network/umee/v4/x/uibc/quota/keeper"
)

// BeginBlock implements BeginBlock for the x/uibc module.
func BeginBlock(ctx sdk.Context, keeper keeper.Keeper) {
	params := keeper.GetParams(ctx)
	quotaExpires, err := keeper.GetExpire(ctx)
	if err != nil {
		panic(err)
	}

	// reset quotas
	if quotaExpires == nil || quotaExpires.Before(ctx.BlockTime()) {
		newExpires := ctx.BlockTime().Add(params.QuotaDuration)
		if err := keeper.SetExpire(ctx, newExpires); err != nil {
			panic(err)
		}
		if err := keeper.SetTotalOutflowSum(ctx, sdk.NewDec(0)); err != nil {
			panic(err)
		}

		quotaOfIBCDenoms, err := keeper.GetQuotaOfIBCDenoms(ctx)
		if err != nil {
			panic(err)
		}

		for _, quotaOfIBCDenom := range quotaOfIBCDenoms {
			// reset the outflow sum to 0
			quotaOfIBCDenom.OutflowSum = sdk.NewDec(0)
			// storing the rate limits to store
			if err := keeper.SetDenomQuota(ctx, quotaOfIBCDenom); err != nil {
				panic(err)
			}
		}
	}
}

// EndBlocker implements EndBlock for the x/leverage module.
func EndBlocker(_ sdk.Context, _ keeper.Keeper) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}
