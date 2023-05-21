package incentive

import (
	"testing"

	"gotest.tools/v3/assert"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v4/util/coin"
	leveragetypes "github.com/umee-network/umee/v4/x/leverage/types"
)

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
		UToken:  "u/uumee",
		Rewards: sdk.NewDecCoins(),
	}
	duplicateRewardTracker := DefaultGenesis()
	duplicateRewardTracker.RewardTrackers = []RewardTracker{rt, rt}
	assert.ErrorContains(t, duplicateRewardTracker.Validate(), "duplicate reward trackers")

	invalidRewardAccumulator := DefaultGenesis()
	invalidRewardAccumulator.RewardAccumulators = []RewardAccumulator{{}}
	assert.ErrorIs(t, invalidRewardAccumulator.Validate(), leveragetypes.ErrNotUToken)

	ra := RewardAccumulator{
		UToken:  "u/uumee",
		Rewards: sdk.NewDecCoins(),
	}
	duplicateRewardAccumulator := DefaultGenesis()
	duplicateRewardAccumulator.RewardAccumulators = []RewardAccumulator{ra, ra}
	assert.ErrorContains(t, duplicateRewardAccumulator.Validate(), "duplicate reward accumulators")

	invalidProgram := IncentiveProgram{}
	validProgram := NewIncentiveProgram(1, 1, 1, "u/uumee", coin.New("uumee", 1), coin.Zero("uumee"), false)

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
		UToken:  sdk.NewInt64Coin("u/uumee", 1),
	}

	duplicateBond := DefaultGenesis()
	duplicateBond.Bonds = []Bond{b, b}
	assert.ErrorContains(t, duplicateBond.Validate(), "duplicate bonds")

	invalidAccountUnbondings := DefaultGenesis()
	invalidAccountUnbondings.AccountUnbondings = []AccountUnbondings{{}}
	assert.ErrorContains(t, invalidAccountUnbondings.Validate(), "empty address string is not allowed")

	au := AccountUnbondings{
		Account:    validAddr,
		UToken:     "u/uumee",
		Unbondings: []Unbonding{},
	}
	duplicateAccountUnbonding := DefaultGenesis()
	duplicateAccountUnbonding.AccountUnbondings = []AccountUnbondings{au, au}
	assert.ErrorContains(t, duplicateAccountUnbonding.Validate(), "duplicate account unbondings")
}

func TestValidateIncentiveProgram(t *testing.T) {
	validProgram := NewIncentiveProgram(1, 1, 1, "u/uumee", coin.New("uumee", 1), coin.Zero("uumee"), false)
	assert.NilError(t, validProgram.Validate())

	invalidUToken := validProgram
	invalidUToken.UToken = ""
	assert.ErrorContains(t, invalidUToken.Validate(), "invalid denom")

	invalidUToken.UToken = "uumee"
	assert.ErrorIs(t, invalidUToken.Validate(), leveragetypes.ErrNotUToken)

	invalidTotalRewards := validProgram
	invalidTotalRewards.TotalRewards = sdk.Coin{}
	assert.ErrorContains(t, invalidTotalRewards.Validate(), "invalid denom")

	invalidTotalRewards.TotalRewards = coin.New("u/uumee", 100)
	assert.ErrorIs(t, invalidTotalRewards.Validate(), leveragetypes.ErrUToken)

	invalidTotalRewards.TotalRewards = coin.Zero("uumee")
	assert.ErrorIs(t, invalidTotalRewards.Validate(), ErrProgramWithoutRewards)

	invalidRemainingRewards := validProgram
	invalidRemainingRewards.RemainingRewards = sdk.Coin{}
	assert.ErrorContains(t, invalidRemainingRewards.Validate(), "invalid denom")

	invalidRemainingRewards.RemainingRewards = coin.Zero("abcd")
	assert.ErrorIs(t, invalidRemainingRewards.Validate(), ErrProgramRewardMismatch)

	invalidRemainingRewards.RemainingRewards = coin.New("uumee", 1)
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
	proposedRemainingRewards.RemainingRewards = coin.New("uumee", 1)
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
