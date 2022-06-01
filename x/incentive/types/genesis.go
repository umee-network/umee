package types

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"
)

// NewGenesisState creates a new GenesisState object
func NewGenesisState(
	params Params,
	programs []IncentiveProgram,
	programRewards []ProgramReward,
	nextID uint32,
	totalLocked sdk.Coins,
	lockAmounts []LockAmount,
	pendingRewards []PendingReward,
	rewardBases []RewardBasis,
	rewardAccumulators []RewardAccumulator,
	unlockings []Unlocking,
) *GenesisState {
	return &GenesisState{
		Params:             params,
		Programs:           programs,
		ProgramRewards:     programRewards,
		NextId:             nextID,
		TotalLocked:        totalLocked,
		LockAmounts:        lockAmounts,
		PendingRewards:     pendingRewards,
		RewardBases:        rewardBases,
		RewardAccumulators: rewardAccumulators,
		Unlockings:         unlockings,
	}
}

// DefaultGenesis returns the default genesis state of the x/incentive module.
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:      DefaultParams(),
		NextId:      1,
		TotalLocked: sdk.NewCoins(),
	}
}

// Validate performs basic genesis state validation returning an error upon any
// failure.
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}

	// TODO: Either group programs and rewards by ID map and ensure matching
	// denoms / ids or combine the rewards field back into program, maybe
	// eliminating ID, or other.

	// TODO: Finish validation logic

	for _, p := range gs.Programs {
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

// NewProgramReward creates the ProgramReward struct used in GenesisState
func NewProgramReward(id uint32, amount sdk.Coin) ProgramReward {
	return ProgramReward{
		Id:     id,
		Amount: amount,
	}
}

// NewLockAmount creates the LockAmount struct used in GenesisState
func NewLockAmount(addr string, tier uint32, amount sdk.Coin) LockAmount {
	return LockAmount{
		Address: addr,
		Tier:    tier,
		Amount:  amount,
	}
}

// NewPendingReward creates the PendingReward struct used in GenesisState
func NewPendingReward(addr string, amount sdk.Coins) PendingReward {
	return PendingReward{
		Address:       addr,
		PendingReward: amount,
	}
}

// NewRewardBasis creates the RewardBasis struct used in GenesisState
func NewRewardBasis(addr, lockDenom string, tier uint32, basis sdk.DecCoins) RewardBasis {
	return RewardBasis{
		Address:     addr,
		LockDenom:   lockDenom,
		Tier:        tier,
		RewardBasis: basis,
	}
}

// NewRewardAccumulator creates the RewardAccumulator struct used in GenesisState
func NewRewardAccumulator(addr, lockDenom string, tier uint32, basis sdk.DecCoins) RewardAccumulator {
	return RewardAccumulator{
		LockDenom:   lockDenom,
		Tier:        tier,
		RewardBasis: basis,
	}
}

// NewUnlocking creates the Unlocking struct used in GenesisState
func NewUnlocking(addr string, tier uint32, unlockTime uint64, amount sdk.Coin) Unlocking {
	return Unlocking{
		Address: addr,
		Tier:    tier,
		End:     unlockTime,
		Amount:  amount,
	}
}
