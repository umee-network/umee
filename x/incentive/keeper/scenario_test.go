package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/umee-network/umee/v4/util/coin"
	"github.com/umee-network/umee/v4/x/incentive"
	"github.com/umee-network/umee/v4/x/leverage/fixtures"
	leveragetypes "github.com/umee-network/umee/v4/x/leverage/types"
)

const (
	umee   = fixtures.UmeeDenom
	atom   = fixtures.AtomDenom
	u_umee = leveragetypes.UTokenPrefix + fixtures.UmeeDenom
	u_atom = leveragetypes.UTokenPrefix + fixtures.AtomDenom
)

func TestBasicIncentivePrograms(t *testing.T) {
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
		coin.New("u/"+fixtures.UmeeDenom, 100_000000),
	)

	// create three separate programs for 10UMEE, which will run for 100 seconds
	// one is funded by the community fund, and two are not. The non-community ones are start later than the first.
	// The first non-community-funded program will not be sponsored, and should thus be cancelled and create no rewards.
	k.addIncentiveProgram(u_umee, 100, 100, sdk.NewInt64Coin(umee, 10_000000), true)
	k.addIncentiveProgram(u_umee, 120, 120, sdk.NewInt64Coin(umee, 10_000000), false)
	k.addIncentiveProgram(u_umee, 120, 120, sdk.NewInt64Coin(umee, 10_000000), false)

	// verify all 3 programs added
	programs, err := k.getAllIncentivePrograms(k.ctx, incentive.ProgramStatusUpcoming)
	require.NoError(k.t, err)
	require.Equal(k.t, 3, len(programs))

	// fund the third program manually
	k.sponsor(sponsor, 3)

	// Verify funding states
	require.True(k.t, k.programFunded(1))
	require.False(k.t, k.programFunded(2))
	require.True(k.t, k.programFunded(3))

	// Verify program status
	require.Equal(k.t, incentive.ProgramStatusUpcoming, k.programStatus(1), "program 1 status (time 1)")
	require.Equal(k.t, incentive.ProgramStatusUpcoming, k.programStatus(2), "program 2 status (time 1)")
	require.Equal(k.t, incentive.ProgramStatusUpcoming, k.programStatus(3), "program 3 status (time 1)")

	// Advance last rewards time to 100, thus starting the first program
	k.advanceTimeTo(100)
	require.Equal(k.t, incentive.ProgramStatusOngoing, k.programStatus(1), "program 1 status (time 100)")
	require.Equal(k.t, incentive.ProgramStatusUpcoming, k.programStatus(2), "program 2 status (time 100)")
	require.Equal(k.t, incentive.ProgramStatusUpcoming, k.programStatus(3), "program 3 status (time 100)")
	// Because rewards are distributed before programs status is updated, no rewards
	// should have been distributed this block
	program1 := k.getProgram(1)
	require.Equal(k.t, program1.TotalRewards, program1.RemainingRewards, "no rewards on program's start block")

	// Advance last rewards time to 101, thus distributing 1 block (1%) of the first program's rewards.
	// No additional programs have started yet.
	k.advanceTimeTo(101)
	require.Equal(k.t, incentive.ProgramStatusOngoing, k.programStatus(1), "program 1 status (time 101)")
	require.Equal(k.t, incentive.ProgramStatusUpcoming, k.programStatus(2), "program 2 status (time 101)")
	require.Equal(k.t, incentive.ProgramStatusUpcoming, k.programStatus(3), "program 3 status (time 101)")
	// 9.9UMEE of the original 10 UMEE remain
	program1 = k.getProgram(1)
	require.Equal(k.t, sdk.NewInt(9_900000), program1.RemainingRewards.Amount, "99 percent of program 1 rewards remain")

	// init a second supplier with bonded uTokens - but he was not present during updateRewards
	bob := k.newBondedAccount(
		coin.New(u_umee, 25_000000),
		coin.New(u_atom, 8_000000),
	)

	// From 100000 rewards distributed, 100% went to alice and 0% went to bob.
	// Pending rewards round down.
	rewards, err := k.calculateRewards(k.ctx, alice)
	require.NoError(k.t, err)
	require.Equal(
		k.t,
		sdk.NewCoins(sdk.NewInt64Coin(umee, 100000)),
		rewards,
		"alice pending rewards at time 101",
	)
	rewards, err = k.calculateRewards(k.ctx, bob)
	require.NoError(k.t, err)
	require.Equal(
		k.t,
		sdk.NewCoins(),
		rewards,
		"bob pending rewards at time 101",
	)

	// Advance last rewards time to 102, thus distributing 1 block (1%) of the first program's rewards.
	// No additional programs have started yet.
	k.advanceTimeTo(102)
	require.Equal(k.t, incentive.ProgramStatusOngoing, k.programStatus(1), "program 1 status (time 102)")
	require.Equal(k.t, incentive.ProgramStatusUpcoming, k.programStatus(2), "program 2 status (time 102)")
	require.Equal(k.t, incentive.ProgramStatusUpcoming, k.programStatus(3), "program 3 status (time 102)")
	// 9.8UMEE of the original 10 UMEE remain.
	// rewards actually distributed rounded down a bit, so remaining rewards have a little more left over.
	program1 = k.getProgram(1)
	require.Equal(k.t, sdk.NewInt(9_800001), program1.RemainingRewards.Amount, "98 percent of program 1 rewards remain")

	// From 100000 rewards distributed this new block, 80% went to alice and 20% went to bob.
	// since alice hasn't claimed rewards yet, these add to the previous block's rewards.
	// rewards actually distributed rounded down a bit, and due to decimal remainders, their sum falls short
	// of the amount that was removed from remainingRewards.
	rewards, err = k.calculateRewards(k.ctx, alice)
	require.NoError(k.t, err)
	require.Equal(
		k.t,
		sdk.NewCoins(sdk.NewInt64Coin(umee, 179999)),
		rewards,
		"alice pending rewards at time 102",
	)
	rewards, err = k.calculateRewards(k.ctx, bob)
	require.NoError(k.t, err)
	require.Equal(
		k.t,
		sdk.NewCoins(sdk.NewInt64Coin(umee, 19999)),
		rewards,
		"bob pending rewards at time 102",
	)

	// Advance last rewards time to 120, starting two additional programs.
	// The one that was not funded is considered completed (a no-op for rewards) instead.
	k.advanceTimeTo(120)
	require.Equal(k.t, incentive.ProgramStatusOngoing, k.programStatus(1), "program 1 status (time 120)")
	require.Equal(k.t, incentive.ProgramStatusCompleted, k.programStatus(2), "program 2 status (time 120)")
	require.Equal(k.t, incentive.ProgramStatusOngoing, k.programStatus(3), "program 3 status (time 120)")

	// Advance last rewards time to 300, ending all programs.
	k.advanceTimeTo(300)
	require.Equal(k.t, incentive.ProgramStatusCompleted, k.programStatus(1), "program 1 status (time 300)")
	require.Equal(k.t, incentive.ProgramStatusCompleted, k.programStatus(2), "program 2 status (time 300)")
	require.Equal(k.t, incentive.ProgramStatusCompleted, k.programStatus(3), "program 3 status (time 300)")
	// Remaining rewards should be exactly zero.
	program1 = k.getProgram(1)
	program2 := k.getProgram(2)
	program3 := k.getProgram(3)
	require.Equal(k.t, sdk.ZeroInt(), program1.RemainingRewards.Amount, "0 percent of program 1 rewards remain")
	require.Equal(k.t, sdk.ZeroInt(), program2.RemainingRewards.Amount, "0 percent of program 2 rewards remain")
	require.Equal(k.t, sdk.ZeroInt(), program3.RemainingRewards.Amount, "0 percent of program 3 rewards remain")

	// These are the final pending rewards observed.
	rewards, err = k.calculateRewards(k.ctx, alice)
	require.NoError(k.t, err)
	require.Equal(
		k.t,
		// a small amount from before bob joined, then 80% of the rest of program 1, and 80% of program 3
		sdk.NewCoins(sdk.NewInt64Coin(umee, 100000+7_920000+8_000000)),
		rewards,
		"alice pending rewards at time 300",
	)
	rewards, err = k.calculateRewards(k.ctx, bob)
	require.NoError(k.t, err)
	require.Equal(
		k.t,
		// 20% of the rest of program 1 (missing the first block), and 20% of program 3
		sdk.NewCoins(sdk.NewInt64Coin(umee, 1_980000+2_000000)),
		rewards,
		"bob pending rewards at time 300",
	)
}

func TestZeroBonded(t *testing.T) {
	k := newTestKeeper(t)
	k.initCommunityFund(
		coin.New(umee, 1000_000000),
	)

	// create incentive program and fund from community
	k.addIncentiveProgram(u_umee, 100, 100, sdk.NewInt64Coin(umee, 10_000000), true)

	// Advance last rewards time to 100, thus starting the program
	k.advanceTimeTo(100)
	require.Equal(k.t, incentive.ProgramStatusOngoing, k.programStatus(1), "program 1 status (time 100)")

	// Advance last rewards time to 150, which would distribute 50 blocks (50%) of the program's rewards.
	// Since no uTokens are bonded though, the rewards are not distributed.
	k.advanceTimeTo(150)
	require.Equal(k.t, incentive.ProgramStatusOngoing, k.programStatus(1), "program 1 status (time 150)")
	// 10UMEE of the original 10 UMEE remain
	program := k.getProgram(1)
	require.Equal(k.t, program.TotalRewards, program.RemainingRewards, "all of program's rewards remain (no bonds)")

	// init a supplier with bonded uTokens
	k.newBondedAccount(
		coin.New(u_umee, 100_000000),
	)

	// Advance last rewards time to 175, which would originally distribute another 25% of the program's rewards
	// for a total of 75%, but now distributes 50% since the first 50 blocks were skipped due to no bonded uTokens.
	k.advanceTimeTo(175)
	// 5UMEE of the original 5 UMEE remain
	program = k.getProgram(1)
	require.Equal(k.t, sdk.NewInt(5_000000), program.RemainingRewards.Amount, "all of program's rewards remain (no bonds)")
}
