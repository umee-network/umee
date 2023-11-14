package keeper

import sdk "github.com/cosmos/cosmos-sdk/types"

// GetOldTotalOutflow returns the total outflow of ibc-transfer amount.
// Note: only use for v6.2 migration from v6.1.0
func (k Keeper) GetOldTotalOutflow() sdk.Dec {
	bz := k.store.Get(keyTotalOutflows)
	return sdk.MustNewDecFromStr(string(bz))
}

// MigrateTotalOutflowSum migrate the old total outflow type to new one
// Note: only use for v6.2 migration from v6.1.0
func (k Keeper) MigrateTotalOutflowSum() {
	oldTotalOutflow := k.GetOldTotalOutflow()
	k.SetTotalOutflowSum(oldTotalOutflow)
}
