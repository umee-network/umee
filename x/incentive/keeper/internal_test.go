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

func (tk *TestKeeper) UpdateRewards(ctx sdk.Context, prevTime, blockTime int64) error {
	return tk.Keeper.updateRewards(ctx, prevTime, blockTime)
}

func (tk *TestKeeper) UpdatePrograms(ctx sdk.Context, blockTime int64) error {
	return tk.Keeper.updatePrograms(ctx, blockTime)
}

func (tk *TestKeeper) GetIncentivePrograms(ctx sdk.Context, status incentive.ProgramStatus,
) ([]incentive.IncentiveProgram, error) {
	return tk.Keeper.getAllIncentivePrograms(ctx, status)
}

func (tk *TestKeeper) GetIncentiveProgram(ctx sdk.Context, id uint32,
) (incentive.IncentiveProgram, incentive.ProgramStatus, error) {
	return tk.Keeper.getIncentiveProgram(ctx, id)
}

func (tk *TestKeeper) GetNextProgramID(ctx sdk.Context) uint32 {
	return tk.Keeper.getNextProgramID(ctx)
}

func (tk *TestKeeper) PendingRewards(ctx sdk.Context, addr sdk.AccAddress) sdk.Coins {
	return tk.Keeper.calculateRewards(ctx, addr)
}
