package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/umee-network/umee/v5/util/coin"
	"github.com/umee-network/umee/v5/x/incentive"
	"github.com/umee-network/umee/v5/x/leverage/fixtures"
	leveragetypes "github.com/umee-network/umee/v5/x/leverage/types"
)

const (
	umee  = fixtures.UmeeDenom
	atom  = fixtures.AtomDenom
	uUmee = leveragetypes.UTokenPrefix + fixtures.UmeeDenom
	uAtom = leveragetypes.UTokenPrefix + fixtures.AtomDenom
)

func TestBasicIncentivePrograms(t *testing.T) {
	t.Parallel()
	k := newTestKeeper(t)

	// init a community fund with 1000 UMEE and 10 ATOM available for funding
	k.initCommunityFund(
		coin.New(umee, 1000_000000),
		coin.New(atom, 10_000000),
	)

	// init a third party sponsor account with 1000 UMEE and 10 ATOM available for funding
	sponsor := k.newAccount(
		coin.New(umee, 1000_000000),
		coin.New(atom, 10_000000),
	)

	// init a supplier with bonded uTokens
	alice := k.newBondedAccount(
		coin.New(uUmee, 100_000000),
	)

	// create three separate programs for 10UMEE, which will run for 100 seconds
	// one is funded by the community fund, and two are not. The non-community ones are start later than the first.
	// The first non-community-funded program will not be sponsored, and should thus be cancelled and create no rewards.
	k.addIncentiveProgram(uUmee, 100, 100, sdk.NewInt64Coin(umee, 10_000000), true)
	k.addIncentiveProgram(uUmee, 120, 120, sdk.NewInt64Coin(umee, 10_000000), false)
	k.addIncentiveProgram(uUmee, 120, 120, sdk.NewInt64Coin(umee, 10_000000), false)

	// verify all 3 programs added
	programs, err := k.getAllIncentivePrograms(k.ctx, incentive.ProgramStatusUpcoming)
	require.NoError(t, err)
	require.Equal(t, 3, len(programs))

	// fund the third program manually
	k.sponsor(sponsor, 3)

	// Verify funding states
	require.True(t, k.programFunded(1))
	require.False(t, k.programFunded(2))
	require.True(t, k.programFunded(3))

	// Verify program status
	require.Equal(t, incentive.ProgramStatusUpcoming, k.programStatus(1), "program 1 status (time 1)")
	require.Equal(t, incentive.ProgramStatusUpcoming, k.programStatus(2), "program 2 status (time 1)")
	require.Equal(t, incentive.ProgramStatusUpcoming, k.programStatus(3), "program 3 status (time 1)")

	// Advance last rewards time to 100, thus starting the first program
	k.advanceTimeTo(100)
	require.Equal(t, incentive.ProgramStatusOngoing, k.programStatus(1), "program 1 status (time 100)")
	require.Equal(t, incentive.ProgramStatusUpcoming, k.programStatus(2), "program 2 status (time 100)")
	require.Equal(t, incentive.ProgramStatusUpcoming, k.programStatus(3), "program 3 status (time 100)")
	// Because rewards are distributed before programs status is updated, no rewards
	// should have been distributed this block
	program1 := k.getProgram(1)
	require.Equal(t, program1.TotalRewards, program1.RemainingRewards, "no rewards on program's start block")

	// Advance last rewards time to 101, thus distributing 1 block (1%) of the first program's rewards.
	// No additional programs have started yet.
	k.advanceTimeTo(101)
	require.Equal(t, incentive.ProgramStatusOngoing, k.programStatus(1), "program 1 status (time 101)")
	require.Equal(t, incentive.ProgramStatusUpcoming, k.programStatus(2), "program 2 status (time 101)")
	require.Equal(t, incentive.ProgramStatusUpcoming, k.programStatus(3), "program 3 status (time 101)")
	// 9.9UMEE of the original 10 UMEE remain
	program1 = k.getProgram(1)
	require.Equal(t, sdk.NewInt(9_900000), program1.RemainingRewards.Amount, "99 percent of program 1 rewards remain")

	// init a second supplier with bonded uTokens - but he was not present during updateRewards
	bob := k.newBondedAccount(
		coin.New(uUmee, 25_000000),
		coin.New(uAtom, 8_000000),
	)

	// From 100000 rewards distributed, 100% went to alice and 0% went to bob.
	// Pending rewards round down.
	rewards, err := k.calculateRewards(k.ctx, alice)
	require.NoError(t, err)
	require.Equal(
		t,
		sdk.NewCoins(sdk.NewInt64Coin(umee, 100000)),
		rewards,
		"alice pending rewards at time 101",
	)
	rewards, err = k.calculateRewards(k.ctx, bob)
	require.NoError(t, err)
	require.Equal(
		t,
		sdk.NewCoins(),
		rewards,
		"bob pending rewards at time 101",
	)

	// Advance last rewards time to 102, thus distributing 1 block (1%) of the first program's rewards.
	// No additional programs have started yet.
	k.advanceTimeTo(102)
	require.Equal(t, incentive.ProgramStatusOngoing, k.programStatus(1), "program 1 status (time 102)")
	require.Equal(t, incentive.ProgramStatusUpcoming, k.programStatus(2), "program 2 status (time 102)")
	require.Equal(t, incentive.ProgramStatusUpcoming, k.programStatus(3), "program 3 status (time 102)")
	// 9.8UMEE of the original 10 UMEE remain.
	// rewards actually distributed rounded down a bit, so remaining rewards have a little more left over.
	program1 = k.getProgram(1)
	require.Equal(t, sdk.NewInt(9_800001), program1.RemainingRewards.Amount, "98 percent of program 1 rewards remain")

	// From 100000 rewards distributed this new block, 80% went to alice and 20% went to bob.
	// since alice hasn't claimed rewards yet, these add to the previous block's rewards.
	// rewards actually distributed rounded down a bit, and due to decimal remainders, their sum falls short
	// of the amount that was removed from remainingRewards.
	rewards, err = k.calculateRewards(k.ctx, alice)
	require.NoError(t, err)
	require.Equal(
		t,
		sdk.NewCoins(sdk.NewInt64Coin(umee, 179999)),
		rewards,
		"alice pending rewards at time 102",
	)
	rewards, err = k.calculateRewards(k.ctx, bob)
	require.NoError(t, err)
	require.Equal(
		t,
		sdk.NewCoins(sdk.NewInt64Coin(umee, 19999)),
		rewards,
		"bob pending rewards at time 102",
	)

	// Advance last rewards time to 120, starting two additional programs.
	// The one that was not funded is considered completed (a no-op for rewards) instead.
	k.advanceTimeTo(120)
	require.Equal(t, incentive.ProgramStatusOngoing, k.programStatus(1), "program 1 status (time 120)")
	require.Equal(t, incentive.ProgramStatusCompleted, k.programStatus(2), "program 2 status (time 120)")
	require.Equal(t, incentive.ProgramStatusOngoing, k.programStatus(3), "program 3 status (time 120)")

	// Advance last rewards time to 300, ending all programs.
	k.advanceTimeTo(300)
	require.Equal(t, incentive.ProgramStatusCompleted, k.programStatus(1), "program 1 status (time 300)")
	require.Equal(t, incentive.ProgramStatusCompleted, k.programStatus(2), "program 2 status (time 300)")
	require.Equal(t, incentive.ProgramStatusCompleted, k.programStatus(3), "program 3 status (time 300)")
	// Remaining rewards should be exactly zero.
	program1 = k.getProgram(1)
	program2 := k.getProgram(2)
	program3 := k.getProgram(3)
	require.Equal(t, sdk.ZeroInt(), program1.RemainingRewards.Amount, "0 percent of program 1 rewards remain")
	require.Equal(t, sdk.ZeroInt(), program2.RemainingRewards.Amount, "0 percent of program 2 rewards remain")
	require.Equal(t, sdk.ZeroInt(), program3.RemainingRewards.Amount, "0 percent of program 3 rewards remain")

	// verify all 3 programs ended
	programs, err = k.getAllIncentivePrograms(k.ctx, incentive.ProgramStatusCompleted)
	require.NoError(k.t, err)
	require.Equal(k.t, 3, len(programs))
	programs, err = k.getAllIncentivePrograms(k.ctx, incentive.ProgramStatusOngoing)
	require.NoError(k.t, err)
	require.Equal(k.t, 0, len(programs))

	// a small amount from before bob joined, then 80% of the rest of program 1, and 80% of program 3
	aliceRewards := int64(100000 + 7_920000 + 8_000000)
	// 20% of the rest of program 1 (missing the first block), and 20% of program 3
	bobRewards := int64(1_980000 + 2_000000)

	// These are the final pending rewards observed.
	rewards, err = k.calculateRewards(k.ctx, alice)
	require.NoError(t, err)
	require.Equal(
		t,
		sdk.NewCoins(sdk.NewInt64Coin(umee, aliceRewards)),
		rewards,
		"alice pending rewards at time 300",
	)
	// actually claim the rewards (same amount)
	rewards, err = k.UpdateAccount(k.ctx, alice)
	require.NoError(k.t, err)
	require.Equal(
		k.t,
		sdk.NewCoins(sdk.NewInt64Coin(umee, aliceRewards)),
		rewards,
		"alice claimed rewards at time 300",
	)
	// no more pending rewards after claiming
	rewards, err = k.calculateRewards(k.ctx, alice)
	require.NoError(k.t, err)
	require.Equal(k.t, sdk.NewCoins(), rewards, "alice pending rewards after claim")

	rewards, err = k.calculateRewards(k.ctx, bob)
	require.NoError(t, err)
	require.Equal(
		t,
		sdk.NewCoins(sdk.NewInt64Coin(umee, bobRewards)),
		rewards,
		"bob pending rewards at time 300",
	)
}

func TestZeroBonded(t *testing.T) {
	t.Parallel()
	k := newTestKeeper(t)
	k.initCommunityFund(
		coin.New(umee, 1000_000000),
	)

	// In this test case, an incentive program is started but no uTokens of the incentivized denom are
	// bonded during its first half of runtime. During this time, it must not distribute rewards.
	// During the remaining half of the program, all rewards must be distributed (spread evenly over
	// the remaining time.)

	programStart := int64(100)
	k.addIncentiveProgram(uUmee, programStart, 100, sdk.NewInt64Coin(umee, 10_000000), true)
	k.advanceTimeTo(programStart) // starts program, but does not attempt rewards. Do not combine with next line.
	k.advanceTimeTo(programStart + 50)
	program := k.getProgram(1)
	require.Equal(t, incentive.ProgramStatusOngoing, k.programStatus(1), "program 1 status (time 150)")
	require.Equal(t, program.TotalRewards, program.RemainingRewards, "all of program's rewards remain (no bonds)")

	// now create a supplier with bonded tokens, halfway through the program
	alice := k.newBondedAccount(
		coin.New(uUmee, 100_000000),
	)
	k.advanceTimeTo(programStart + 75)
	program = k.getProgram(1)
	require.Equal(t, incentive.ProgramStatusOngoing, k.programStatus(1), "program 1 status (time 175)")
	require.Equal(t, sdk.NewInt(5_000000), program.RemainingRewards.Amount, "half of program rewards distributed")

	// finish the program with user still bonded
	k.advanceTimeTo(programStart + 100)
	program = k.getProgram(1)
	require.Equal(t, incentive.ProgramStatusCompleted, k.programStatus(1), "program 1 status (time 200)")
	require.Equal(t, sdk.ZeroInt(), program.RemainingRewards.Amount, "all of program rewards distributed")

	// measure pending rewards (even though program has ended, user has not yet claimed)
	rewards, err := k.calculateRewards(k.ctx, alice)
	require.NoError(t, err)
	require.Equal(
		t,
		sdk.NewCoins(sdk.NewInt64Coin(umee, 10_000000)),
		rewards,
		"alice pending rewards at time 200",
	)

	// advance time further past program end
	k.advanceTimeTo(programStart + 120)

	// measure pending rewards (unchanged, as user has not yet claimed)
	rewards, err = k.calculateRewards(k.ctx, alice)
	require.NoError(t, err)
	require.Equal(
		t,
		sdk.NewCoins(sdk.NewInt64Coin(umee, 10_000000)),
		rewards,
		"alice pending rewards at time 220",
	)
	// actually claim the rewards (same amount)
	rewards, err = k.UpdateAccount(k.ctx, alice)
	require.NoError(k.t, err)
	require.Equal(
		k.t,
		sdk.NewCoins(sdk.NewInt64Coin(umee, 10_000000)),
		rewards,
		"alice claimed rewards at time 220",
	)
}

func TestZeroBondedAtProgramEnd(t *testing.T) {
	t.Parallel()
	k := newTestKeeper(t)
	k.initCommunityFund(
		coin.New(umee, 1000_000000),
	)

	// In this test case, an incentive program is started but no uTokens of the incentivized denom are
	// bonded during its first quarter or last quarter of runtime. During this time, it must not distribute rewards.
	// During the remaining half of the program, 2/3 rewards must be distributed (spread evenly over
	// the remaining time.) It is 2/3 instead of 3/4 because upon reaching 25% duration with no bonds, the
	// program can adapt to award 1/3 rewards every remaining 25% duration. However, once all users unbond
	// after 75% duration and never return, the program is left with some rewards it cannot distribute.

	programStart := int64(100)
	k.addIncentiveProgram(uUmee, programStart, 100, sdk.NewInt64Coin(umee, 10_000000), true)
	k.advanceTimeTo(programStart)      // starts program, but does not attempt rewards. Do not combine with next line.
	k.advanceTimeTo(programStart + 25) // 25% duration
	program := k.getProgram(1)
	require.Equal(t, incentive.ProgramStatusOngoing, k.programStatus(1), "program 1 status ongoing (time 125)")
	require.Equal(t, program.TotalRewards, program.RemainingRewards, "all of program's rewards remain (no bonds)")

	// now create a supplier with bonded tokens, halfway through the program
	alice := k.newBondedAccount(
		coin.New(uUmee, 100_000000),
	)
	k.advanceTimeTo(programStart + 75) // 75% duration
	program = k.getProgram(1)
	require.Equal(t, incentive.ProgramStatusOngoing, k.programStatus(1), "program 1 status ongoing (time 175)")
	require.Equal(t, sdk.NewInt(3_333334), program.RemainingRewards.Amount, "two thirds of program rewards distributed")

	// measure pending rewards
	aliceReward := sdk.NewCoins(sdk.NewInt64Coin(umee, 6_666666))
	rewards, err := k.calculateRewards(k.ctx, alice)
	require.NoError(t, err)
	require.Equal(t, aliceReward, rewards, "alice pending rewards at time 175")
	// actually claim the rewards (same amount)
	rewards, err = k.UpdateAccount(k.ctx, alice)
	require.NoError(k.t, err)
	require.Equal(k.t, aliceReward, rewards, "alice claimed rewards at time 175")
	// begin unbonding user at 75%, making her ineligible future rewards unless she bonds again
	k.mustBeginUnbond(alice, coin.New(uUmee, 100_000000))

	// complete the program
	k.advanceTimeTo(programStart + 100) // 100% duration
	program = k.getProgram(1)
	require.Equal(t, incentive.ProgramStatusCompleted, k.programStatus(1), "program 1 status completed (time 200)")
	require.Equal(t, sdk.NewInt(3_333334), program.RemainingRewards.Amount, "two thirds of program rewards distributed")

	// measure pending rewards (zero)
	noRewards := sdk.NewCoins()
	rewards, err = k.calculateRewards(k.ctx, alice)
	require.NoError(t, err)
	require.Equal(t, noRewards, rewards, "alice pending rewards at time 200")
	// actually claim the rewards (same amount)
	rewards, err = k.UpdateAccount(k.ctx, alice)
	require.NoError(k.t, err)
	require.Equal(k.t, noRewards, rewards, "alice claimed rewards at time 200")
}

func TestUserSupplyBeforeAndDuring(t *testing.T) {
	t.Parallel()
	k := newTestKeeper(t)
	k.initCommunityFund(
		coin.New(umee, 1000_000000),
	)

	// In this test case, A user supplies and bonds uUmee before the incentive program starts
	// and another user supplies half way through the incentive program

	programStart := int64(100)
	k.addIncentiveProgram(uUmee, programStart, 100, sdk.NewInt64Coin(umee, 10_000000), true)

	// now create a supplier with bonded tokens before the time starts
	k.advanceTimeTo(80)
	alice := k.newBondedAccount(
		coin.New(uUmee, 10_000000),
	)

	k.advanceTimeTo(programStart) // starts program,
	program := k.getProgram(1)
	require.Equal(t, incentive.ProgramStatusOngoing, k.programStatus(1), "program 1 status (time 150)")
	require.Equal(t, program.TotalRewards, program.RemainingRewards, "all of program's rewards remain (no bonds)")

	k.advanceTimeTo(programStart + 50) // time passed half

	program = k.getProgram(1)
	require.Equal(t, incentive.ProgramStatusOngoing, k.programStatus(1), "program 1 status (time 175)")
	require.Equal(t, sdk.NewInt(5_000000), program.RemainingRewards.Amount, "half of program rewards distributed")

	// now creates another supplier with bonded tokens, half way through the program.
	bob := k.newBondedAccount(
		coin.New(uUmee, 30_000000),
	)

	// finish the program with user still bonded
	k.advanceTimeTo(programStart + 100)
	program = k.getProgram(1)
	require.Equal(t, incentive.ProgramStatusCompleted, k.programStatus(1), "program 1 status (time 200)")
	require.Equal(t, sdk.ZeroInt(), program.RemainingRewards.Amount, "all of program rewards distributed")

	// measure pending rewards (even though program has ended, user has not yet claimed)
	rewards, err := k.calculateRewards(k.ctx, alice)
	require.NoError(t, err)

	require.Equal(
		t,
		sdk.NewCoins(sdk.NewInt64Coin(umee, 62_50000)),
		rewards,
		"alice pending rewards at time 200",
	)

	// measure pending rewards (even though program has ended, user has not yet claimed)
	rewards, err = k.calculateRewards(k.ctx, bob)
	require.NoError(t, err)

	require.Equal(
		t,
		sdk.NewCoins(sdk.NewInt64Coin(umee, 3_750000)),
		rewards,
		"bobs pending rewards at time 200",
	)

	// advance time further past program end
	k.advanceTimeTo(programStart + 120)

	// measure pending rewards (unchanged, as user has not yet claimed)
	rewards, err = k.calculateRewards(k.ctx, alice)
	require.NoError(t, err)

	require.Equal(
		t,
		sdk.NewCoins(sdk.NewInt64Coin(umee, 6_250000)),
		rewards,
		"alice pending rewards at time 220",
	)
	// actually claim the rewards (same amount)
	rewards, err = k.UpdateAccount(k.ctx, alice)
	require.NoError(k.t, err)
	require.Equal(
		k.t,
		sdk.NewCoins(sdk.NewInt64Coin(umee, 6_250000)),
		rewards,
		"alice claimed rewards at time 220",
	)

	// measure pending rewards (unchanged, as user has not yet claimed)
	rewards, err = k.calculateRewards(k.ctx, bob)
	require.NoError(t, err)

	require.Equal(
		t,
		sdk.NewCoins(sdk.NewInt64Coin(umee, 3_750000)),
		rewards,
		"bob pending rewards at time 220",
	)
	// actually claim the rewards (second account)
	rewards, err = k.UpdateAccount(k.ctx, bob)
	require.NoError(k.t, err)
	require.Equal(
		k.t,
		sdk.NewCoins(sdk.NewInt64Coin(umee, 3_750000)),
		rewards,
		"bob claimed rewards at time 220",
	)
}

func TestPartialWithdraw(t *testing.T) {
	t.Parallel()
	k := newTestKeeper(t)
	k.initCommunityFund(
		coin.New(umee, 1000_000000),
	)

	// In this test case, A user supplies and bonds uUmee before the incentive program starts
	// and another user supplies half way through the incentive program. The second user then withdraws ~3/4 into the incentive program.

	programStart := int64(100)
	k.addIncentiveProgram(uUmee, programStart, 100, sdk.NewInt64Coin(umee, 10_000000), true)

	// now create a supplier with bonded tokens before the time starts
	k.advanceTimeTo(80)
	alice := k.newBondedAccount(
		coin.New(uUmee, 10_000000),
	)

	k.advanceTimeTo(programStart) // starts program,
	program := k.getProgram(1)
	require.Equal(t, incentive.ProgramStatusOngoing, k.programStatus(1), "program 1 status (time 150)")
	require.Equal(t, program.TotalRewards, program.RemainingRewards, "all of program's rewards remain (no bonds)")

	k.advanceTimeTo(programStart + 50) // time passed half

	program = k.getProgram(1)
	require.Equal(t, incentive.ProgramStatusOngoing, k.programStatus(1), "program 1 status (time 175)")
	require.Equal(t, sdk.NewInt(5_000000), program.RemainingRewards.Amount, "half of program rewards distributed")

	// now creates another supplier with bonded tokens, half way through the program.
	bob := k.newBondedAccount(
		coin.New(uUmee, 30_000000),
	)

	k.advanceTimeTo(programStart + 70) // more time has passed

	// measure pending rewards (unchanged, as user has not yet claimed)
	rewards, err := k.calculateRewards(k.ctx, bob)
	require.NoError(t, err)

	require.Equal(
		t,
		sdk.NewCoins(sdk.NewInt64Coin(umee, 1_500000)),
		rewards,
		"bob pending rewards at time 220",
	)
	// unbonds 20 tokens - still has 10 bonded
	k.mustBeginUnbond(bob, coin.New(uUmee, 20_000000))

	require.NoError(k.t, err)
	require.Equal(
		k.t,
		sdk.NewCoins(sdk.NewInt64Coin(umee, 1_500000)),
		rewards,
		"bob claimed rewards at time 220",
	)

	// finish the program with user still bonded
	k.advanceTimeTo(programStart + 100)
	program = k.getProgram(1)
	require.Equal(t, incentive.ProgramStatusCompleted, k.programStatus(1), "program 1 status (time 200)")
	require.Equal(t, sdk.ZeroInt(), program.RemainingRewards.Amount, "all of program rewards distributed")

	// measure pending rewards (even though program has ended, user has not yet claimed)
	rewards, err = k.calculateRewards(k.ctx, alice)
	require.NoError(t, err)

	require.Equal(
		t,
		sdk.NewCoins(sdk.NewInt64Coin(umee, 7_000000)),
		rewards,
		"alice pending rewards at time 200",
	)

	// measure pending rewards (even though program has ended, user has not yet claimed)
	rewards, err = k.calculateRewards(k.ctx, bob)
	require.NoError(t, err)

	require.Equal(
		t,
		sdk.NewCoins(sdk.NewInt64Coin(umee, 1_500000)),
		rewards,
		"bobs pending rewards at time 200",
	)

	// advance time further past program end
	k.advanceTimeTo(programStart + 120)

	// measure pending rewards (unchanged, as user has not yet claimed)
	rewards, err = k.calculateRewards(k.ctx, alice)
	require.NoError(t, err)

	require.Equal(
		t,
		sdk.NewCoins(sdk.NewInt64Coin(umee, 7_000000)),
		rewards,
		"alice pending rewards at time 220",
	)
	// actually claim the rewards (same amount)
	rewards, err = k.UpdateAccount(k.ctx, alice)
	require.NoError(k.t, err)
	require.Equal(
		k.t,
		sdk.NewCoins(sdk.NewInt64Coin(umee, 7_000000)),
		rewards,
		"alice claimed rewards at time 220",
	)

	// measure pending rewards (unchanged, as user has not yet claimed)
	rewards, err = k.calculateRewards(k.ctx, bob)
	require.NoError(t, err)

	require.Equal(
		t,
		sdk.NewCoins(sdk.NewInt64Coin(umee, 1_500000)),
		rewards,
		"bob pending rewards at time 220",
	)
	// actually claim the rewards (second account)
	rewards, err = k.UpdateAccount(k.ctx, bob)
	require.NoError(k.t, err)
	require.Equal(
		k.t,
		sdk.NewCoins(sdk.NewInt64Coin(umee, 1_500000)),
		rewards,
		"bob claimed rewards at time 220",
	)
}
