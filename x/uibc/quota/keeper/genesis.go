package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v6/util"
	"github.com/umee-network/umee/v6/x/uibc"
)

// InitGenesis initializes the x/uibc module's state from a provided genesis
// state.
func (kb Builder) InitGenesis(ctx sdk.Context, genState uibc.GenesisState) {
	k := kb.Keeper(&ctx)
	err := k.SetParams(genState.Params)
	util.Panic(err)

	k.SetTokenOutflows(genState.Outflows)
	k.SetTokenInflows(genState.Inflows)
	k.SetOutflowSum(genState.OutflowSum)
	k.SetInflowSum(genState.InflowSum)

	err = k.SetExpire(genState.QuotaExpires)
	util.Panic(err)
}

// ExportGenesis returns the x/uibc module's exported genesis state.
func (kb Builder) ExportGenesis(ctx sdk.Context) *uibc.GenesisState {
	k := kb.Keeper(&ctx)
	outflows, err := k.GetAllOutflows()
	util.Panic(err)
	inflows, err := k.GetAllInflows()
	util.Panic(err)
	quotaExpires, err := k.GetExpire()
	util.Panic(err)

	return &uibc.GenesisState{
		Params:       k.GetParams(),
		Outflows:     outflows,
		OutflowSum:   k.GetOutflowSum(),
		QuotaExpires: *quotaExpires,
		Inflows:      inflows,
		InflowSum:    k.GetInflowSum(),
	}
}
