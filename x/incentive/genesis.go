package incentive

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewGenesisState creates a new GenesisState object
func NewGenesisState(
	params Params,
	completedPrograms []IncentiveProgram,
	ongoingPrograms []IncentiveProgram,
	upcomingPrograms []IncentiveProgram,
	nextProgramID uint32,
	lastRewardTime uint64,
	totalBonded []TotalBond,
	bonds []Bond,
	rewardTrackers []RewardTracker,
	rewardAccumulators []RewardAccumulator,
	accountUnbondings []AccountUnbondings,
) *GenesisState {
	return &GenesisState{
		Params:             params,
		CompletedPrograms:  completedPrograms,
		OngoingPrograms:    ongoingPrograms,
		UpcomingPrograms:   upcomingPrograms,
		NextProgramId:      nextProgramID,
		LastRewardsTime:    lastRewardTime,
		TotalBonded:        totalBonded,
		Bonds:              bonds,
		RewardTrackers:     rewardTrackers,
		RewardAccumulators: rewardAccumulators,
		AccountUnbondings:  accountUnbondings,
	}
}

// DefaultGenesisState returns the default genesis state
func DefaultGenesisState() *GenesisState {
	return &GenesisState{
		Params:          DefaultParams(),
		NextProgramId:   1,
		LastRewardsTime: 0,
	}
}

// ValidateGenesis checks a genesis state for basic issues
func ValidateGenesis(_ GenesisState) error {
	// TODO #1749
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

// NewIncentiveProgram creates the IncentiveProgram struct used in GenesisState
func NewIncentiveProgram(
	id uint32,
	startTime uint64,
	duration uint64,
	bondDenom string,
	totalRewards, fundedRewards, remainingRewards sdk.Coin,
) IncentiveProgram {
	return IncentiveProgram{
		Id:               id,
		StartTime:        startTime,
		Duration:         duration,
		Denom:            bondDenom,
		TotalRewards:     totalRewards,
		FundedRewards:    fundedRewards,
		RemainingRewards: remainingRewards,
	}
}

// NewBond creates the Bond struct used in GenesisState
func NewBond(addr string, tier uint32, coin sdk.Coin) Bond {
	return Bond{
		Account: addr,
		Tier:    tier,
		Amount:  coin,
	}
}

// NewTotalBond creates the TotalBond struct used in GenesisState
func NewTotalBond(tier uint32, coin sdk.Coin) Bond {
	return Bond{
		Tier:   tier,
		Amount: coin,
	}
}

// NewRewardTracker creates the RewardTracker struct used in GenesisState
func NewRewardTracker(addr, denom string, tier uint32, coins sdk.DecCoins) RewardTracker {
	return RewardTracker{
		Account:       addr,
		Denom:         denom,
		Tier:          tier,
		RewardTracker: coins,
	}
}

// NewRewardAccumulator creates the RewardAccumulator struct used in GenesisState
func NewRewardAccumulator(denom string, tier uint32, coins sdk.DecCoins) RewardAccumulator {
	return RewardAccumulator{
		Denom:         denom,
		Tier:          tier,
		RewardTracker: coins,
	}
}

// NewUnbonding creates the Unbonding struct used in GenesisState
func NewUnbonding(tier uint32, endTime uint64, coin sdk.Coin) Unbonding {
	return Unbonding{
		Tier:   tier,
		End:    endTime,
		Amount: coin,
	}
}

// NewAccountUnbondings creates the AccountUnbondings struct used in GenesisState
func NewAccountUnbondings(addr string, unbondings []Unbonding) AccountUnbondings {
	return AccountUnbondings{
		Account:    addr,
		Unbondings: unbondings,
	}
}
