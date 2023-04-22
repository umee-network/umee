package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"
	"github.com/umee-network/umee/v4/x/incentive"

	"github.com/umee-network/umee/v4/util/coin"
	//"github.com/umee-network/umee/v4/x/incentive"
	"github.com/umee-network/umee/v4/x/leverage/fixtures"
	leveragetypes "github.com/umee-network/umee/v4/x/leverage/types"
)

func TestBasicIncentivePrograms(t *testing.T) {
	const (
		umee  = fixtures.UmeeDenom
		atom  = fixtures.AtomDenom
		uumee = leveragetypes.UTokenPrefix + fixtures.UmeeDenom
		uatom = leveragetypes.UTokenPrefix + fixtures.AtomDenom
	)

	k := newTestKeeper(t)

	// init a community fund with 1000 UMEE and 10 ATOM available for funding
	_ = k.initCommunityFund(
		coin.New(umee, 1000_000000),
		coin.New(atom, 10_000000),
	)

	// init a third party sponsor account with 1000 UMEE and 10 ATOM available for funding
	sponsor := k.newAccount(
		coin.New(umee, 1000_000000),
		coin.New(atom, 10_000000),
	)

	// init and bond two suppliers with varying bonded uTokens
	_ = k.newBondedAccount(
		coin.New("u/"+fixtures.UmeeDenom, 100_000000),
	)
	_ = k.newBondedAccount(
		coin.New("u/"+fixtures.UmeeDenom, 25_000000),
		coin.New("u/"+fixtures.AtomDenom, 8_000000),
	)

	// create three separate programs for 10UMEE, which will run for 100 seconds
	// one is funded by the community fund, and two are not. The non-community ones are start later than the first.
	// The first non-community-funded program will not be sponsored, and should thus be cancelled and create no rewards.
	k.addIncentiveProgram(uumee, 100, 100, sdk.NewInt64Coin(umee, 10_000000), true)
	k.addIncentiveProgram(uumee, 100, 120, sdk.NewInt64Coin(umee, 10_000000), false)
	k.addIncentiveProgram(uumee, 100, 120, sdk.NewInt64Coin(umee, 10_000000), false)

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

	// TODO: fix bug

	/*
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

	   // From 100000 rewards distributed, 66% went to alice and 33% went to bob.
	   // Pending rewards round down.
	   require.Equal(

	   	k.t,
	   	sdk.NewCoins(sdk.NewInt64Coin(umee, 66666)),
	   	k.calculateRewards(k.ctx, alice),
	   	"alice pending rewards at time 101",

	   )
	   require.Equal(

	   	k.t,
	   	sdk.NewCoins(sdk.NewInt64Coin(umee, 33333)),
	   	k.calculateRewards(k.ctx, bob),
	   	"bob pending rewards at time 101",

	   )
	*/
}
