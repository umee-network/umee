package uibc

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	abci "github.com/tendermint/tendermint/abci/types"
	"github.com/umee-network/umee/v3/x/uibc/quota/keeper"
)

// BeginBlock implements BeginBlock for the x/uibc module.
func BeginBlock(ctx sdk.Context, keeper keeper.Keeper) {
	quotaOfIBCDenoms, err := keeper.GetQuotaOfIBCDenoms(ctx)
	if err != nil {
		panic(err)
	}

	params := keeper.GetParams(ctx)

	quotaExpires, err := keeper.GetQuotaExpires(ctx)
	if err != nil {
		panic(err)
	}

	if quotaExpires == nil {
		if err := keeper.SetQuotaExpires(ctx, ctx.BlockTime().Add(params.QuotaDuration)); err != nil {
			panic(err)
		}
		if err := keeper.SetTotalOutflowSum(ctx, sdk.NewDec(0).String()); err != nil {
			panic(err)
		}
	} else if quotaExpires.Before(ctx.BlockTime()) {
		if err := keeper.SetQuotaExpires(ctx, ctx.BlockTime().Add(params.QuotaDuration)); err != nil {
			panic(err)
		}
		if err := keeper.SetTotalOutflowSum(ctx, sdk.NewDec(0).String()); err != nil {
			panic(err)
		}
	}

	for _, quotaOfIBCDenom := range quotaOfIBCDenoms {
		if quotaOfIBCDenom.Expires == nil {
			expires := ctx.BlockTime().Add(params.QuotaDuration)
			quotaOfIBCDenom.Expires = &expires
		} else if quotaOfIBCDenom.Expires.Before(ctx.BlockTime()) {
			// reset the expire time
			expires := ctx.BlockTime().Add(params.QuotaDuration)
			quotaOfIBCDenom.Expires = &expires
			// reset the outflow sum to 0
			quotaOfIBCDenom.OutflowSum = sdk.NewDec(0)
		}
		// storing the rate limits to store
		if err := keeper.SetQuotaOfIBCDenom(ctx, quotaOfIBCDenom); err != nil {
			panic(err)
		}
	}
}

// EndBlocker implements EndBlock for the x/leverage module.
func EndBlocker(_ sdk.Context, _ keeper.Keeper) []abci.ValidatorUpdate {
	return []abci.ValidatorUpdate{}
}
