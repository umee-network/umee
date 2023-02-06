package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v4/util"
	"github.com/umee-network/umee/v4/x/uibc"
)

// InitGenesis initializes the x/uibc module's state from a provided genesis
// state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState uibc.GenesisState) {
	err := k.SetParams(ctx, genState.Params)
	util.Panic(err)

	k.SetDenomQuotas(ctx, genState.Quotas)
	k.SetTotalOutflowSum(ctx, genState.TotalOutflowSum)

	err = k.SetExpire(ctx, genState.QuotaExpires)
	util.Panic(err)
}

// ExportGenesis returns the x/uibc module's exported genesis state.
func (k Keeper) ExportGenesis(ctx sdk.Context) *uibc.GenesisState {
	quotas, err := k.GetQuotaOfIBCDenoms(ctx)
	util.Panic(err)
	quotaExpires, err := k.GetExpire(ctx)
	util.Panic(err)

	return &uibc.GenesisState{
		Params:          k.GetParams(ctx),
		Quotas:          quotas,
		TotalOutflowSum: k.GetTotalOutflowSum(ctx),
		QuotaExpires:    *quotaExpires,
	}
}
