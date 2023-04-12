package incentive

import (
	"encoding/json"

	"github.com/cosmos/cosmos-sdk/codec"
	sdk "github.com/cosmos/cosmos-sdk/types"

	leveragetypes "github.com/umee-network/umee/v4/x/leverage/types"
)

// NewGenesisState creates a new GenesisState object
func NewGenesisState(
	params Params,
	completedPrograms []IncentiveProgram,
	ongoingPrograms []IncentiveProgram,
	upcomingPrograms []IncentiveProgram,
	nextProgramID uint32,
	lastRewardTime int64,
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
		Bonds:              bonds,
		RewardTrackers:     rewardTrackers,
		RewardAccumulators: rewardAccumulators,
		AccountUnbondings:  accountUnbondings,
	}
}

// DefaultGenesis returns the default genesis state
func DefaultGenesis() *GenesisState {
	return &GenesisState{
		Params:          DefaultParams(),
		NextProgramId:   1,
		LastRewardsTime: 0,
	}
}

// Validate checks a genesis state for basic issues
func (gs GenesisState) Validate() error {
	if err := gs.Params.Validate(); err != nil {
		return err
	}
	if gs.NextProgramId == 0 {
		return ErrInvalidProgramID.Wrap("next program ID was zero")
	}
	if gs.LastRewardsTime < 0 {
		return ErrDecreaseLastRewardTime.Wrap("last reward time was negative")
	}
	// TODO: enforce no duplicate (account,denom)
	for _, rt := range gs.RewardTrackers {
		if _, err := sdk.AccAddressFromBech32(rt.Account); err != nil {
			return err
		}
		if !leveragetypes.HasUTokenPrefix(rt.Denom) {
			return leveragetypes.ErrNotUToken.Wrap(rt.Denom)
		}
		if err := rt.Rewards.Validate(); err != nil {
			return err
		}
		for _, r := range rt.Rewards {
			if leveragetypes.HasUTokenPrefix(r.Denom) {
				return leveragetypes.ErrUToken.Wrap(r.Denom)
			}
		}
	}
	// TODO: enforce no duplicate denoms
	for _, ra := range gs.RewardAccumulators {
		if !leveragetypes.HasUTokenPrefix(ra.Denom) {
			return leveragetypes.ErrNotUToken.Wrap(ra.Denom)
		}
		if err := ra.Rewards.Validate(); err != nil {
			return err
		}
		for _, r := range ra.Rewards {
			if leveragetypes.HasUTokenPrefix(r.Denom) {
				return leveragetypes.ErrUToken.Wrap(r.Denom)
			}
		}
	}
	// TODO: enforce no duplicate program IDs
	for _, up := range gs.UpcomingPrograms {
		if err := validatePassedIncentiveProgram(up); err != nil {
			return err
		}
	}
	for _, op := range gs.OngoingPrograms {
		if err := validatePassedIncentiveProgram(op); err != nil {
			return err
		}
	}
	for _, cp := range gs.CompletedPrograms {
		if err := validatePassedIncentiveProgram(cp); err != nil {
			return err
		}
	}

	// TODO: enforce no duplicate (account,denom)
	for _, b := range gs.Bonds {
		if _, err := sdk.AccAddressFromBech32(b.Account); err != nil {
			return err
		}
		if err := b.UToken.Validate(); err != nil {
			return err
		}
		if !leveragetypes.HasUTokenPrefix(b.UToken.Denom) {
			return leveragetypes.ErrNotUToken.Wrap(b.UToken.Denom)
		}
	}

	// TODO: enforce no duplicate (account,denom)
	for _, au := range gs.AccountUnbondings {
		if _, err := sdk.AccAddressFromBech32(au.Account); err != nil {
			return err
		}
		if !leveragetypes.HasUTokenPrefix(au.Denom) {
			return leveragetypes.ErrNotUToken.Wrap(au.Denom)
		}
		for _, u := range au.Unbondings {
			if u.End < u.Start {
				return ErrInvalidUnbonding.Wrap("start time > end time")
			}
			if err := u.UToken.Validate(); err != nil {
				return err
			}
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

// NewIncentiveProgram creates the IncentiveProgram struct used in GenesisState
func NewIncentiveProgram(
	id uint32,
	startTime int64,
	duration int64,
	uDenom string,
	totalRewards, remainingRewards sdk.Coin,
	funded bool,
) IncentiveProgram {
	return IncentiveProgram{
		ID:               id,
		StartTime:        startTime,
		Duration:         duration,
		UDenom:           uDenom,
		TotalRewards:     totalRewards,
		RemainingRewards: remainingRewards,
		Funded:           funded,
	}
}

// NewBond creates the Bond struct used in GenesisState
func NewBond(addr string, coin sdk.Coin) Bond {
	return Bond{
		Account: addr,
		UToken:  coin,
	}
}

// NewRewardTracker creates the RewardTracker struct used in GenesisState
func NewRewardTracker(addr, denom string, coins sdk.DecCoins) RewardTracker {
	return RewardTracker{
		Account: addr,
		Denom:   denom,
		Rewards: coins,
	}
}

// NewRewardAccumulator creates the RewardAccumulator struct used in GenesisState
func NewRewardAccumulator(denom string, exponent uint32, coins sdk.DecCoins) RewardAccumulator {
	return RewardAccumulator{
		Denom:    denom,
		Exponent: exponent,
		Rewards:  coins,
	}
}

// NewUnbonding creates the Unbonding struct used in GenesisState
func NewUnbonding(startTime, endTime int64, coin sdk.Coin) Unbonding {
	return Unbonding{
		Start:  startTime,
		End:    endTime,
		UToken: coin,
	}
}

// NewAccountUnbondings creates the AccountUnbondings struct used in GenesisState
func NewAccountUnbondings(addr, denom string, unbondings []Unbonding) AccountUnbondings {
	return AccountUnbondings{
		Account:    addr,
		Denom:      denom,
		Unbondings: unbondings,
	}
}
