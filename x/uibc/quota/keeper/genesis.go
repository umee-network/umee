package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v4/x/uibc"
)

// InitGenesis initializes the x/uibc module's state from a provided genesis
// state.
func (k Keeper) InitGenesis(ctx sdk.Context, genState uibc.GenesisState) {
	if err := k.SetParams(ctx, genState.Params); err != nil {
		panic(err)
	}

	if err := k.SetQuotaOfIBCDenoms(ctx, genState.Quotas); err != nil {
		panic(err)
	}

	if err := k.SetTotalOutflowSum(ctx, genState.TotalOutflowSum); err != nil {
		panic(err)
	}

	if err := k.SetExpire(ctx, genState.QuotaExpires); err != nil {
		panic(err)
	}
}

// ExportGenesis returns the x/leverage module's exported genesis state.
func (k Keeper) ExportGenesis(ctx sdk.Context) *uibc.GenesisState {
	quotas, err := k.GetQuotaOfIBCDenoms(ctx)
	if err != nil {
		panic(err)
	}

	quotaExpires, err := k.GetExpire(ctx)
	if err != nil {
		panic(err)
	}

	return &uibc.GenesisState{
		Params:          k.GetParams(ctx),
		Quotas:          quotas,
		TotalOutflowSum: k.GetTotalOutflowSum(ctx),
		QuotaExpires:    *quotaExpires,
	}
}
