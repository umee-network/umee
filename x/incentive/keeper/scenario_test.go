package keeper_test

import (
	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/umee-network/umee/v4/x/incentive"
	//"github.com/umee-network/umee/v4/util/coin"
	//"github.com/umee-network/umee/v4/x/incentive"
	//leveragetypes "github.com/umee-network/umee/v4/x/leverage/types"
)

func (s *IntegrationTestSuite) TestBasicIncentivePrograms() {
	ctx, k, require := s.ctx, s.k, s.Require()
	// lk := s.mockLeverage

	// init a community fund with 1000 UMEE and 10 ATOM available for funding
	_ = s.initCommunityFund(
		sdk.NewInt64Coin(umeeDenom, 1000_000000),
		sdk.NewInt64Coin(atomDenom, 10_000000),
	)

	// init a third party sponsor account with 1000 UMEE and 10 ATOM available for funding
	sponsor := s.newAccount(
		sdk.NewInt64Coin(umeeDenom, 1000_000000),
		sdk.NewInt64Coin(atomDenom, 10_000000),
	)

	// init and bond two suppliers with varying bonded uTokens
	alice := s.newBondedAccount(
		sdk.NewInt64Coin("u/"+umeeDenom, 100_000000),
	)
	bob := s.newBondedAccount(
		sdk.NewInt64Coin("u/"+umeeDenom, 25_000000),
		sdk.NewInt64Coin("u/"+atomDenom, 8_000000),
	)

	// create three separate programs for 10UMEE, which will run for 100 seconds
	// one is funded by the community fund, and two are not. The non-community ones are start later than the first.
	// The first non-community-funded program will not be sponsored, and should thus be cancelled and create no rewards.
	id1 := s.addIncentiveProgram("u/"+umeeDenom, 100, 100, sdk.NewInt64Coin(umeeDenom, 10_000000), true)
	id2 := s.addIncentiveProgram("u/"+umeeDenom, 100, 120, sdk.NewInt64Coin(umeeDenom, 10_000000), false)
	id3 := s.addIncentiveProgram("u/"+umeeDenom, 100, 120, sdk.NewInt64Coin(umeeDenom, 10_000000), false)

	// verify all 3 programs added
	programs, err := k.GetIncentivePrograms(ctx, incentive.ProgramStatusUpcoming)
	require.NoError(err)
	require.Equal(3, len(programs))

	// fund the third program manually
	s.sponsor(sponsor, id3)

	// Verify funding states
	require.True(s.programFunded(id1))
	require.False(s.programFunded(id2))
	require.True(s.programFunded(id3))

	// Verify program status
	require.Equal(incentive.ProgramStatusUpcoming, s.programStatus(id1), "program 1 status (time 1)")
	require.Equal(incentive.ProgramStatusUpcoming, s.programStatus(id2), "program 2 status (time 1)")
	require.Equal(incentive.ProgramStatusUpcoming, s.programStatus(id3), "program 3 status (time 1)")

	// Advance last rewards time to 100, thus starting the first program
	s.advanceTimeTo(100)
	require.Equal(incentive.ProgramStatusOngoing, s.programStatus(id1), "program 1 status (time 100)")
	require.Equal(incentive.ProgramStatusUpcoming, s.programStatus(id2), "program 2 status (time 100)")
	require.Equal(incentive.ProgramStatusUpcoming, s.programStatus(id3), "program 3 status (time 100)")
	// Because rewards are distributed before programs status is updated, no rewards
	// should have been distributed this block
	program1 := s.getProgram(id1)
	require.Equal(program1.TotalRewards, program1.RemainingRewards, "no rewards on program's start block")

	// Advance last rewards time to 101, thus distributing 1 block (1%) of the first program's rewards.
	// No additional programs have started yet.
	s.advanceTimeTo(101)
	require.Equal(incentive.ProgramStatusOngoing, s.programStatus(id1), "program 1 status (time 101)")
	require.Equal(incentive.ProgramStatusUpcoming, s.programStatus(id2), "program 2 status (time 101)")
	require.Equal(incentive.ProgramStatusUpcoming, s.programStatus(id3), "program 3 status (time 101)")
	// 9.9UMEE of the original 10 UMEE remain
	program1 = s.getProgram(id1)
	require.Equal(sdk.NewInt(9_900000), program1.RemainingRewards.Amount, "99 percent of program 1 rewards remain")

	// From 100000 rewards distributed, 66% went to alice and 33% went to bob.
	// Pending rewards round down.
	require.Equal(
		sdk.NewCoins(sdk.NewInt64Coin(umeeDenom, 66666)),
		k.PendingRewards(ctx, alice),
		"alice pending rewards at time 101",
	)
	require.Equal(
		sdk.NewCoins(sdk.NewInt64Coin(umeeDenom, 33333)),
		k.PendingRewards(ctx, bob),
		"bob pending rewards at time 101",
	)
}
