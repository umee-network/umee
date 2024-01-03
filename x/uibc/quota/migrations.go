package quota

import sdk "github.com/cosmos/cosmos-sdk/types"

// getOldTotalOutflow returns the total outflow of ibc-transfer amount.
// Note: only use for v6.2 migration from v6.1.0
func (k Keeper) getOldTotalOutflow() sdk.Dec {
	bz := k.store.Get(keyOutflowSum)
	return sdk.MustNewDecFromStr(string(bz))
}

// MigrateOutflowSum migrate the old total outflow type to new one
// Note: only use for v6.2 migration from v6.1.0
func (k Keeper) MigrateOutflowSum() {
	oldTotalOutflow := k.getOldTotalOutflow()
	k.SetOutflowSum(oldTotalOutflow)
}
