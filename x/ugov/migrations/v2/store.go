package v2

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v6/x/ugov"
)

// MigrateStore performs in-place store migrations from 1 to 2
func MigrateStore(ctx sdk.Context, k ugov.Keeper) error {
	ip := ugov.DefaultInflationParams()
	if err := k.SetInflationParams(ip); err != nil {
		return err
	}

	cycleEnd := ctx.BlockTime().Add(ip.InflationCycle)
	return k.SetInflationCycleEnd(cycleEnd)
}
