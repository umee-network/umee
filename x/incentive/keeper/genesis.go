package keeper

import (
	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v5/x/incentive"
)

// InitGenesis initializes the x/incentive module state from a provided genesis state.
func (k Keeper) InitGenesis(ctx sdk.Context, gs incentive.GenesisState) {
	if err := gs.Validate(); err != nil {
		panic(err)
	}

	if err := k.setParams(ctx, gs.Params); err != nil {
		panic(err)
	}
	if err := k.setNextProgramID(ctx, gs.NextProgramId); err != nil {
		panic(err)
	}
	if err := k.setLastRewardsTime(ctx, gs.LastRewardsTime); err != nil {
		panic(err)
	}

	for _, ip := range gs.UpcomingPrograms {
		if err := k.setIncentiveProgram(ctx, ip, incentive.ProgramStatusUpcoming); err != nil {
			panic(err)
		}
	}
	for _, ip := range gs.OngoingPrograms {
		if err := k.setIncentiveProgram(ctx, ip, incentive.ProgramStatusOngoing); err != nil {
			panic(err)
		}
	}
	for _, ip := range gs.CompletedPrograms {
		if err := k.setIncentiveProgram(ctx, ip, incentive.ProgramStatusCompleted); err != nil {
			panic(err)
		}
	}

	for _, b := range gs.Bonds {
		if err := b.Validate(); err != nil {
			panic(err)
		}
		if err := k.setBonded(ctx, sdk.MustAccAddressFromBech32(b.Account), b.UToken); err != nil {
			panic(err)
		}
	}

	for _, au := range gs.AccountUnbondings {
		if err := k.setUnbondings(ctx, au); err != nil {
			panic(err)
		}
	}

	for _, ra := range gs.RewardAccumulators {
		if err := k.setRewardAccumulator(ctx, ra); err != nil {
			panic(err)
		}
	}

	for _, rt := range gs.RewardTrackers {
		if err := k.setRewardTracker(ctx, rt); err != nil {
			panic(err)
		}
	}
}

// ExportGenesis returns the x/incentive module's exported genesis state.
func (k Keeper) ExportGenesis(ctx sdk.Context) *incentive.GenesisState {
	completedPrograms, err := k.getAllIncentivePrograms(ctx, incentive.ProgramStatusCompleted)
	if err != nil {
		panic(err)
	}
	ongoingPrograms, err := k.getAllIncentivePrograms(ctx, incentive.ProgramStatusOngoing)
	if err != nil {
		panic(err)
	}
	upcomingPrograms, err := k.getAllIncentivePrograms(ctx, incentive.ProgramStatusUpcoming)
	if err != nil {
		panic(err)
	}
	bonds, err := k.getAllBonds(ctx)
	if err != nil {
		panic(err)
	}
	trackers, err := k.getAllRewardTrackers(ctx)
	if err != nil {
		panic(err)
	}
	accumulators, err := k.getAllRewardAccumulators(ctx)
	if err != nil {
		panic(err)
	}
	unbondings, err := k.getAllAccountUnbondings(ctx)
	if err != nil {
		panic(err)
	}
	return incentive.NewGenesisState(
		k.GetParams(ctx),
		completedPrograms,
		ongoingPrograms,
		upcomingPrograms,
		k.getNextProgramID(ctx),
		k.GetLastRewardsTime(ctx),
		bonds,
		trackers,
		accumulators,
		unbondings,
	)
}
