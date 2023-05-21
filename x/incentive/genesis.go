package incentive

import (
	"encoding/json"
	"fmt"

	"cosmossdk.io/errors"
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
		return ErrInvalidProgramID.Wrap("next program ID must not be zero")
	}
	if gs.LastRewardsTime < 0 {
		return ErrDecreaseLastRewardTime.Wrap("last reward time must not be negative")
	}

	m := map[string]bool{}
	for _, rt := range gs.RewardTrackers {
		// enforce no duplicate (account,denom)
		s := rt.Account + rt.UToken
		if err := noDuplicateString(m, s, "reward trackers"); err != nil {
			return err
		}
		if err := rt.Validate(); err != nil {
			return err
		}
	}

	m = map[string]bool{}
	for _, ra := range gs.RewardAccumulators {
		// enforce no duplicate denoms
		s := ra.UToken
		if err := noDuplicateString(m, s, "reward accumulators"); err != nil {
			return err
		}
		if err := ra.Validate(); err != nil {
			return err
		}
	}

	m = map[string]bool{}
	// enforce no duplicate program IDs
	for _, up := range gs.UpcomingPrograms {
		s := fmt.Sprintf("%d", up.ID)
		if err := noDuplicateString(m, s, "upcoming program ID"); err != nil {
			return err
		}
		if err := up.ValidatePassed(); err != nil {
			return err
		}
	}
	for _, op := range gs.OngoingPrograms {
		s := fmt.Sprintf("%d", op.ID)
		if err := noDuplicateString(m, s, "ongoing program ID"); err != nil {
			return err
		}
		if err := op.ValidatePassed(); err != nil {
			return err
		}
	}
	for _, cp := range gs.CompletedPrograms {
		s := fmt.Sprintf("%d", cp.ID)
		if err := noDuplicateString(m, s, "completed program ID"); err != nil {
			return err
		}
		if err := cp.ValidatePassed(); err != nil {
			return err
		}
	}

	m = map[string]bool{}
	for _, b := range gs.Bonds {
		// enforce no duplicate (account,denom)
		s := b.Account + b.UToken.Denom
		if err := noDuplicateString(m, s, "bonds"); err != nil {
			return err
		}
		if err := b.Validate(); err != nil {
			return err
		}
	}

	m = map[string]bool{}
	for _, au := range gs.AccountUnbondings {
		// enforce no duplicate (account,denom)
		s := au.Account + au.UToken
		if err := noDuplicateString(m, s, "account unbondings"); err != nil {
			return err
		}
		if err := au.Validate(); err != nil {
			return err
		}
	}

	return nil
}

// noDuplicateString checks to see if a string is already present in a map
// and then adds it to the map. If it was already present, an error is returned.
// used to check all uniqueness requirements in genesis state.
func noDuplicateString(m map[string]bool, s, errMsg string) error {
	if _, ok := m[s]; !ok {
		return fmt.Errorf("duplicaate %s: %s", errMsg, s)
	}
	m[s] = true
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
		UToken:           uDenom,
		TotalRewards:     totalRewards,
		RemainingRewards: remainingRewards,
		Funded:           funded,
	}
}

// Validate performs validation on an IncentiveProgram type returning an error
// if the program is invalid.
func (ip IncentiveProgram) Validate() error {
	if err := sdk.ValidateDenom(ip.UToken); err != nil {
		return err
	}
	if !leveragetypes.HasUTokenPrefix(ip.UToken) {
		// only allow uToken denoms as bonded denoms
		return errors.Wrap(leveragetypes.ErrNotUToken, ip.UToken)
	}

	if err := ip.TotalRewards.Validate(); err != nil {
		return err
	}
	if leveragetypes.HasUTokenPrefix(ip.TotalRewards.Denom) {
		// only allow base token denoms as rewards
		return errors.Wrap(leveragetypes.ErrUToken, ip.TotalRewards.Denom)
	}
	if ip.TotalRewards.IsZero() {
		return ErrProgramWithoutRewards
	}

	if err := ip.RemainingRewards.Validate(); err != nil {
		return err
	}
	if ip.RemainingRewards.Denom != ip.TotalRewards.Denom {
		return ErrProgramRewardMismatch
	}
	if !ip.Funded && ip.RemainingRewards.IsPositive() {
		return ErrNonfundedProgramRewards
	}

	if ip.Duration <= 0 {
		return errors.Wrapf(ErrInvalidProgramDuration, "%d", ip.Duration)
	}
	if ip.StartTime <= 0 {
		return errors.Wrapf(ErrInvalidProgramStart, "%d", ip.Duration)
	}

	return nil
}

// ValidateProposed runs IncentiveProgram.Validate and also checks additional requirements applying
// to incentive programs which have not yet been funded or passed by governance
func (ip IncentiveProgram) ValidateProposed() error {
	if ip.ID != 0 {
		return ErrInvalidProgramID.Wrapf("%d", ip.ID)
	}
	if !ip.RemainingRewards.IsZero() {
		return ErrNonzeroRemainingRewards.Wrap(ip.RemainingRewards.String())
	}
	if ip.Funded {
		return ErrProposedFundedProgram
	}
	return ip.Validate()
}

// ValidatePassed runs IncentiveProgram.Validate and also checks additional requirements applying
// to incentive programs which have already been passed by governance
func (ip IncentiveProgram) ValidatePassed() error {
	if ip.ID == 0 {
		return ErrInvalidProgramID.Wrapf("%d", ip.ID)
	}
	return ip.Validate()
}

// NewBond creates the Bond struct used in GenesisState
func NewBond(addr string, coin sdk.Coin) Bond {
	return Bond{
		Account: addr,
		UToken:  coin,
	}
}

func (b Bond) Validate() error {
	if _, err := sdk.AccAddressFromBech32(b.Account); err != nil {
		return err
	}
	if err := b.UToken.Validate(); err != nil {
		return err
	}
	if !leveragetypes.HasUTokenPrefix(b.UToken.Denom) {
		return leveragetypes.ErrNotUToken.Wrap(b.UToken.Denom)
	}
	return nil
}

// NewRewardTracker creates the RewardTracker struct used in GenesisState
func NewRewardTracker(addr, uDenom string, coins sdk.DecCoins) RewardTracker {
	return RewardTracker{
		Account: addr,
		UToken:  uDenom,
		Rewards: coins,
	}
}

func (rt RewardTracker) Validate() error {
	if _, err := sdk.AccAddressFromBech32(rt.Account); err != nil {
		return err
	}
	if !leveragetypes.HasUTokenPrefix(rt.UToken) {
		return leveragetypes.ErrNotUToken.Wrap(rt.UToken)
	}
	if err := rt.Rewards.Validate(); err != nil {
		return err
	}
	for _, r := range rt.Rewards {
		if leveragetypes.HasUTokenPrefix(r.Denom) {
			return leveragetypes.ErrUToken.Wrap(r.Denom)
		}
	}
	return nil
}

// NewRewardAccumulator creates the RewardAccumulator struct used in GenesisState
func NewRewardAccumulator(uDenom string, exponent uint32, coins sdk.DecCoins) RewardAccumulator {
	return RewardAccumulator{
		UToken:   uDenom,
		Exponent: exponent,
		Rewards:  coins,
	}
}

func (ra RewardAccumulator) Validate() error {
	if !leveragetypes.HasUTokenPrefix(ra.UToken) {
		return leveragetypes.ErrNotUToken.Wrap(ra.UToken)
	}
	if err := ra.Rewards.Validate(); err != nil {
		return err
	}
	for _, r := range ra.Rewards {
		if leveragetypes.HasUTokenPrefix(r.Denom) {
			return leveragetypes.ErrUToken.Wrap(r.Denom)
		}
	}
	return nil
}

// NewUnbonding creates the Unbonding struct used in GenesisState
func NewUnbonding(startTime, endTime int64, coin sdk.Coin) Unbonding {
	return Unbonding{
		Start:  startTime,
		End:    endTime,
		UToken: coin,
	}
}

func (u Unbonding) Validate() error {
	if u.End < u.Start {
		return ErrInvalidUnbonding.Wrap("start time > end time")
	}
	return u.UToken.Validate()
}

// NewAccountUnbondings creates the AccountUnbondings struct used in GenesisState
func NewAccountUnbondings(addr, uDenom string, unbondings []Unbonding) AccountUnbondings {
	return AccountUnbondings{
		Account:    addr,
		UToken:     uDenom,
		Unbondings: unbondings,
	}
}

func (au AccountUnbondings) Validate() error {
	if _, err := sdk.AccAddressFromBech32(au.Account); err != nil {
		return err
	}
	if !leveragetypes.HasUTokenPrefix(au.UToken) {
		return leveragetypes.ErrNotUToken.Wrap(au.UToken)
	}
	for _, u := range au.Unbondings {
		if u.UToken.Denom != au.UToken {
			return ErrInvalidUnbonding.Wrapf("unbonding denom %s does not match accountUnbondings denom %s",
				u.UToken.Denom, au.UToken)
		}
		if err := u.Validate(); err != nil {
			return err
		}
	}
	return nil
}
