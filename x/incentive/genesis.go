package incentive

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewGenesisState creates a new GenesisState object
func NewGenesisState(
	params Params,
	upcomingPrograms []IncentiveProgram,
	ongoingPrograms []IncentiveProgram,
	completedPrograms []IncentiveProgram,
	nextID uint32,
	lastRewardsTime uint64,
	totalBonded sdk.Coins,
	bondAmounts []BondAmount,
	pendingRewards []PendingReward,
	rewardBases []RewardBasis,
	rewardAccumulators []RewardAccumulator,
	unbondings []Unbonding,
) *GenesisState {
	return &GenesisState{
		Params:             params,
		UpcomingPrograms:   upcomingPrograms,
		OngoingPrograms:    ongoingPrograms,
		CompletedPrograms:  completedPrograms,
		LastRewardsTime:    lastRewardsTime,
		NextProgramId:      nextID,
		TotalBonded:        totalBonded,
		BondAmounts:        bondAmounts,
		PendingRewards:     pendingRewards,
		RewardBases:        rewardBases,
		RewardAccumulators: rewardAccumulators,
		Unbondings:         unbondings,
	}
}

// DefaultGenesis returns the default genesis state of the x/incentive module.
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:          DefaultParams(),
		NextProgramId:   1,
		LastRewardsTime: 0,
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	// TODO: Finish validation logic - includes in-progress programs
	// all ID < NextID (and NextID = len(programs) + 1)

	for _, p := range gs.UpcomingPrograms {
		if err := p.Validate(); err != nil {
			return err
		}
	}
	for _, p := range gs.OngoingPrograms {
		if err := p.Validate(); err != nil {
			return err
		}
	}
	for _, p := range gs.CompletedPrograms {
		if err := p.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// GetGenesisStateFromAppState returns x/incentive GenesisState given raw application
// genesis state.
func GetGenesisStateFromAppState(cdc codec.JSONCodec, appState map[string]json.RawMessage) *GenesisState {
	var genesisState GenesisState

	if appState[ModuleName] != nil {
		cdc.MustUnmarshalJSON(appState[ModuleName], &genesisState)
	}

	return &genesisState
}

// NewBondAmount creates the BondAmount struct used in GenesisState
func NewBondAmount(addr string, tier uint32, amount sdk.Coin) BondAmount {
	return BondAmount{
		Account: addr,
		Tier:    tier,
		Amount:  amount,
	}
}

// NewPendingReward creates the PendingReward struct used in GenesisState
func NewPendingReward(addr string, rewards sdk.Coins) PendingReward {
	return PendingReward{
		Account:       addr,
		PendingReward: rewards,
	}
}

// NewRewardBasis creates the RewardBasis struct used in GenesisState
func NewRewardBasis(addr, denom string, tier uint32, basis sdk.DecCoins) RewardBasis {
	return RewardBasis{
		Account:     addr,
		Denom:       denom,
		Tier:        tier,
		RewardBasis: basis,
	}
}

// NewRewardAccumulator creates the RewardAccumulator struct used in GenesisState
func NewRewardAccumulator(addr, denom string, tier uint32, basis sdk.DecCoins) RewardAccumulator {
	return RewardAccumulator{
		Denom:       denom,
		Tier:        tier,
		RewardBasis: basis,
	}
}

// NewUnbonding creates the Unbonding struct used in GenesisState
func NewUnbonding(addr string, tier uint32, unbondHeight uint64, amount sdk.Coin) Unbonding {
	return Unbonding{
		Account: addr,
		Tier:    tier,
		End:     unbondHeight,
		Amount:  amount,
	}
}
