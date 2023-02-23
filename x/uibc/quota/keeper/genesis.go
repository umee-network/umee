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

	k.SetOutflows(ctx, genState.Outflows)
	k.SetTotalOutflowSum(ctx, genState.TotalOutflowSum)

	err = k.SetExpire(ctx, genState.QuotaExpires)
	util.Panic(err)
}

// ExportGenesis returns the x/uibc module's exported genesis state.
func (k Keeper) ExportGenesis(ctx sdk.Context) *uibc.GenesisState {
	outflows, err := k.GetAllOutflows(ctx)
	util.Panic(err)
	quotaExpires, err := k.GetExpire(ctx)
	util.Panic(err)

	return &uibc.GenesisState{
		Params:          k.GetParams(ctx),
		Outflows:        outflows,
		TotalOutflowSum: k.GetTotalOutflow(ctx),
		QuotaExpires:    *quotaExpires,
	}
}
