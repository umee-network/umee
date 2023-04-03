package keeper

import (
	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	storetypes "github.com/cosmos/cosmos-sdk/store/types"

	"github.com/umee-network/umee/v4/x/incentive"
)

// TestKeeper is a keeper with some normally
// unexported methods exposed for testing.
type TestKeeper struct {
	*Keeper
}

// NewTestKeeper returns a new incentive keeper, and
// an additional TestKeeper that exposes normally
// unexported methods for testing.
func NewTestKeeper(
	cdc codec.Codec,
	storeKey storetypes.StoreKey,
	bk incentive.BankKeeper,
	lk incentive.LeverageKeeper,
) (Keeper, TestKeeper) {
	k := NewKeeper(
		cdc,
		storeKey,
		bk,
		lk,
	)
	return k, TestKeeper{&k}
}

func (tk *TestKeeper) SetLastRewardsTime(ctx sdk.Context, t int64) error {
	return tk.Keeper.setLastRewardsTime(ctx, t)
}
