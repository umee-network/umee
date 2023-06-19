package incentive

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"gotest.tools/v3/assert"

	"github.com/umee-network/umee/v5/util/coin"
	leveragetypes "github.com/umee-network/umee/v5/x/leverage/types"
)

const uumee = "uumee"

func TestValidateGenesis(t *testing.T) {
	validAddr := "umee1s84d29zk3k20xk9f0hvczkax90l9t94g72n6wm"

	genesis := DefaultGenesis()
	assert.NilError(t, genesis.Validate())

	invalidParams := DefaultGenesis()
	invalidParams.Params.EmergencyUnbondFee = sdk.MustNewDecFromStr("-0.01")
	assert.ErrorContains(t, invalidParams.Validate(), "invalid emergency unbonding fee")

	zeroID := DefaultGenesis()
	zeroID.NextProgramId = 0
	assert.ErrorIs(t, zeroID.Validate(), ErrInvalidProgramID)

	negativeTime := DefaultGenesis()
	negativeTime.LastRewardsTime = -1
	assert.ErrorIs(t, negativeTime.Validate(), ErrDecreaseLastRewardTime)

	invalidRewardTracker := DefaultGenesis()
	invalidRewardTracker.RewardTrackers = []RewardTracker{{}}
	assert.ErrorContains(t, invalidRewardTracker.Validate(), "empty address string is not allowed")

	rt := RewardTracker{
		Account: validAddr,
		UToken:  coin.UumeeDenom,
		Rewards: sdk.NewDecCoins(),
	}
	duplicateRewardTracker := DefaultGenesis()
	duplicateRewardTracker.RewardTrackers = []RewardTracker{rt, rt}
	assert.ErrorContains(t, duplicateRewardTracker.Validate(), "duplicate reward trackers")

	invalidRewardAccumulator := DefaultGenesis()
	invalidRewardAccumulator.RewardAccumulators = []RewardAccumulator{{}}
	assert.ErrorContains(t, invalidRewardAccumulator.Validate(), "invalid denom")

	ra := RewardAccumulator{
		UToken:  coin.UumeeDenom,
		Rewards: sdk.NewDecCoins(),
	}
	duplicateRewardAccumulator := DefaultGenesis()
	duplicateRewardAccumulator.RewardAccumulators = []RewardAccumulator{ra, ra}
	assert.ErrorContains(t, duplicateRewardAccumulator.Validate(), "duplicate reward accumulators")

	invalidProgram := IncentiveProgram{}
	validProgram := NewIncentiveProgram(1, 1, 1, coin.UumeeDenom, coin.Umee1, coin.Zero(uumee), false)

	invalidUpcomingProgram := DefaultGenesis()
	invalidUpcomingProgram.UpcomingPrograms = []IncentiveProgram{invalidProgram}
	assert.ErrorIs(t, invalidUpcomingProgram.Validate(), ErrInvalidProgramID)

	duplicateUpcomingProgram := DefaultGenesis()
	duplicateUpcomingProgram.UpcomingPrograms = []IncentiveProgram{validProgram, validProgram}
	assert.ErrorContains(t, duplicateUpcomingProgram.Validate(), "duplicate upcoming program ID")

	invalidOngoingProgram := DefaultGenesis()
	invalidOngoingProgram.OngoingPrograms = []IncentiveProgram{invalidProgram}
	assert.ErrorIs(t, invalidOngoingProgram.Validate(), ErrInvalidProgramID)

	duplicateOngoingProgram := DefaultGenesis()
	duplicateOngoingProgram.UpcomingPrograms = []IncentiveProgram{validProgram}
	duplicateOngoingProgram.OngoingPrograms = []IncentiveProgram{validProgram}
	assert.ErrorContains(t, duplicateOngoingProgram.Validate(), "duplicate ongoing program ID")

	invalidCompletedProgram := DefaultGenesis()
	invalidCompletedProgram.CompletedPrograms = []IncentiveProgram{invalidProgram}
	assert.ErrorIs(t, invalidCompletedProgram.Validate(), ErrInvalidProgramID)

	duplicateCompletedProgram := DefaultGenesis()
	duplicateCompletedProgram.UpcomingPrograms = []IncentiveProgram{validProgram}
	duplicateCompletedProgram.CompletedPrograms = []IncentiveProgram{validProgram}
	assert.ErrorContains(t, duplicateCompletedProgram.Validate(), "duplicate completed program ID")

	invalidBond := DefaultGenesis()
	invalidBond.Bonds = []Bond{{}}
	assert.ErrorContains(t, invalidBond.Validate(), "empty address string is not allowed")

	b := Bond{
		Account: validAddr,
		UToken:  sdk.NewInt64Coin(coin.UumeeDenom, 1),
	}

	duplicateBond := DefaultGenesis()
	duplicateBond.Bonds = []Bond{b, b}
	assert.ErrorContains(t, duplicateBond.Validate(), "duplicate bonds")

	invalidAccountUnbondings := DefaultGenesis()
	invalidAccountUnbondings.AccountUnbondings = []AccountUnbondings{{}}
	assert.ErrorContains(t, invalidAccountUnbondings.Validate(), "empty address string is not allowed")

	au := AccountUnbondings{
		Account:    validAddr,
		UToken:     coin.UumeeDenom,
		Unbondings: []Unbonding{},
	}
	duplicateAccountUnbonding := DefaultGenesis()
	duplicateAccountUnbonding.AccountUnbondings = []AccountUnbondings{au, au}
	assert.ErrorContains(t, duplicateAccountUnbonding.Validate(), "duplicate account unbondings")
}

func TestValidateIncentiveProgram(t *testing.T) {
	validProgram := NewIncentiveProgram(1, 1, 1, coin.UumeeDenom, coin.Umee1, coin.Zero(uumee), false)
	assert.NilError(t, validProgram.Validate())

	invalidUToken := validProgram
	invalidUToken.UToken = ""
	assert.ErrorContains(t, invalidUToken.Validate(), "invalid denom")

	invalidUToken.UToken = uumee
	assert.ErrorIs(t, invalidUToken.Validate(), leveragetypes.ErrNotUToken)

	invalidTotalRewards := validProgram
	invalidTotalRewards.TotalRewards = sdk.Coin{}
	assert.ErrorContains(t, invalidTotalRewards.Validate(), "invalid denom")

	invalidTotalRewards.TotalRewards = coin.New(coin.UumeeDenom, 100)
	assert.ErrorIs(t, invalidTotalRewards.Validate(), leveragetypes.ErrUToken)

	invalidTotalRewards.TotalRewards = coin.Zero(uumee)
	assert.ErrorIs(t, invalidTotalRewards.Validate(), ErrProgramWithoutRewards)

	invalidRemainingRewards := validProgram
	invalidRemainingRewards.RemainingRewards = sdk.Coin{}
	assert.ErrorContains(t, invalidRemainingRewards.Validate(), "invalid denom")

	invalidRemainingRewards.RemainingRewards = coin.Zero("abcd")
	assert.ErrorIs(t, invalidRemainingRewards.Validate(), ErrProgramRewardMismatch)

	invalidRemainingRewards.RemainingRewards = coin.Umee1
	assert.ErrorIs(t, invalidRemainingRewards.Validate(), ErrNonfundedProgramRewards)

	invalidDuration := validProgram
	invalidDuration.Duration = 0
	assert.ErrorIs(t, invalidDuration.Validate(), ErrInvalidProgramDuration)

	invalidStartTime := validProgram
	invalidStartTime.StartTime = 0
	assert.ErrorIs(t, invalidStartTime.Validate(), ErrInvalidProgramStart)

	// also test validateProposed, which is used for incentive programs in MsgGovCreatePrograms
	assert.ErrorIs(t, validProgram.ValidateProposed(), ErrInvalidProgramID, "proposed with nonzero ID")

	validProposed := validProgram
	validProposed.ID = 0
	assert.NilError(t, validProposed.ValidateProposed())

	proposedRemainingRewards := validProposed
	proposedRemainingRewards.RemainingRewards = coin.Umee1
	assert.ErrorIs(t, proposedRemainingRewards.ValidateProposed(), ErrNonzeroRemainingRewards, "proposed remaining rewards")

	invalidProposed := validProposed
	invalidProposed.StartTime = 0
	assert.ErrorIs(t, invalidProposed.ValidateProposed(), ErrInvalidProgramStart, "proposed invalid program")

	proposedFunded := validProposed
	proposedFunded.Funded = true
	assert.ErrorIs(t, proposedFunded.ValidateProposed(), ErrProposedFundedProgram, "proposed funded program")

	// also test validatePassed, which is used for incentive programs in genesis state
	assert.NilError(t, validProgram.ValidatePassed())
	assert.ErrorIs(t, validProposed.ValidatePassed(), ErrInvalidProgramID, "passed program with zero ID")
	assert.ErrorIs(t, invalidStartTime.ValidatePassed(), ErrInvalidProgramStart, "passed invalid program")
}

func TestValidateStructs(t *testing.T) {
	validAddr := "umee1s84d29zk3k20xk9f0hvczkax90l9t94g72n6wm"
	validBond := NewBond(validAddr, coin.New(coin.UumeeDenom, 1))
	assert.NilError(t, validBond.Validate())

	invalidBond := validBond
	invalidBond.Account = ""
	assert.ErrorContains(t, invalidBond.Validate(), "empty address string is not allowed")

	invalidBond = validBond
	invalidBond.UToken = sdk.Coin{}
	assert.ErrorContains(t, invalidBond.Validate(), "invalid denom")

	invalidBond = validBond
	invalidBond.UToken.Denom = uumee
	assert.ErrorIs(t, invalidBond.Validate(), leveragetypes.ErrNotUToken)

	validTracker := NewRewardTracker(validAddr, coin.UumeeDenom, sdk.NewDecCoins(
		sdk.NewDecCoin(uumee, sdk.OneInt()),
	))
	assert.NilError(t, validTracker.Validate())

	invalidTracker := validTracker
	invalidTracker.UToken = ""
	assert.ErrorContains(t, invalidTracker.Validate(), "invalid denom")

	invalidTracker.UToken = uumee
	assert.ErrorIs(t, invalidTracker.Validate(), leveragetypes.ErrNotUToken)

	invalidTracker = validTracker
	invalidTracker.Account = ""
	assert.ErrorContains(t, invalidTracker.Validate(), "empty address string is not allowed")

	invalidTracker = validTracker
	invalidTracker.Rewards[0].Denom = ""
	assert.ErrorContains(t, invalidTracker.Validate(), "invalid denom")

	invalidTracker = validTracker
	invalidTracker.Rewards[0].Denom = coin.UumeeDenom
	assert.ErrorIs(t, invalidTracker.Validate(), leveragetypes.ErrUToken)

	validAccumulator := NewRewardAccumulator(coin.UumeeDenom, 6, sdk.NewDecCoins(
		sdk.NewDecCoin(uumee, sdk.OneInt()),
	))
	assert.NilError(t, validAccumulator.Validate())

	invalidAccumulator := validAccumulator
	invalidAccumulator.UToken = ""
	assert.ErrorContains(t, invalidAccumulator.Validate(), "invalid denom")

	invalidAccumulator.UToken = uumee
	assert.ErrorIs(t, invalidAccumulator.Validate(), leveragetypes.ErrNotUToken)

	invalidAccumulator = validAccumulator
	invalidAccumulator.Rewards[0].Denom = ""
	assert.ErrorContains(t, invalidAccumulator.Validate(), "invalid denom")

	invalidAccumulator = validAccumulator
	invalidAccumulator.Rewards[0].Denom = coin.UumeeDenom
	assert.ErrorIs(t, invalidAccumulator.Validate(), leveragetypes.ErrUToken)

	validUnbonding := NewUnbonding(1, 1, coin.New(coin.UumeeDenom, 1))
	assert.NilError(t, validUnbonding.Validate())

	invalidUnbonding := validUnbonding
	invalidUnbonding.End = 0
	assert.ErrorIs(t, invalidUnbonding.Validate(), ErrInvalidUnbonding)

	invalidUnbonding = validUnbonding
	invalidUnbonding.UToken.Denom = uumee
	assert.ErrorIs(t, invalidUnbonding.Validate(), leveragetypes.ErrNotUToken)

	invalidUnbonding = validUnbonding
	invalidUnbonding.UToken = sdk.Coin{Denom: coin.UumeeDenom, Amount: sdk.NewInt(-1)}
	assert.ErrorContains(t, invalidUnbonding.Validate(), "negative coin amount")

	validAccountUnbondings := NewAccountUnbondings(validAddr, coin.UumeeDenom, []Unbonding{validUnbonding})
	assert.NilError(t, validAccountUnbondings.Validate())

	invalidAccountUnbondings := validAccountUnbondings
	invalidAccountUnbondings.Account = ""
	assert.ErrorContains(t, invalidAccountUnbondings.Validate(), "empty address")

	invalidAccountUnbondings = validAccountUnbondings
	invalidAccountUnbondings.UToken = ""
	assert.ErrorContains(t, invalidAccountUnbondings.Validate(), "invalid denom")

	invalidAccountUnbondings = validAccountUnbondings
	invalidAccountUnbondings.UToken = uumee
	assert.ErrorIs(t, invalidAccountUnbondings.Validate(), leveragetypes.ErrNotUToken)

	invalidAccountUnbondings = validAccountUnbondings
	invalidAccountUnbondings.Unbondings[0].UToken.Denom = uumee
	assert.ErrorContains(t, invalidAccountUnbondings.Validate(), "does not match")

	invalidAccountUnbondings = validAccountUnbondings
	invalidAccountUnbondings.Unbondings[0].End = 0
	assert.ErrorIs(t, invalidAccountUnbondings.Validate(), ErrInvalidUnbonding)
	invalidAccountUnbondings.Unbondings[0] = validUnbonding // the value in validAccountUnbondings was modified

	invalidAccountUnbondings = validAccountUnbondings
	invalidAccountUnbondings.Unbondings[0].UToken = sdk.Coin{Denom: coin.UumeeDenom, Amount: sdk.NewInt(-1)}
	assert.ErrorContains(t, invalidAccountUnbondings.Validate(), "negative coin amount")
}
