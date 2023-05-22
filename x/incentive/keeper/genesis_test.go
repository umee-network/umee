package keeper

import (
	"testing"

	sdk "github.com/cosmos/cosmos-sdk/types"
	"github.com/stretchr/testify/require"

	"github.com/umee-network/umee/v4/util/coin"
)

func TestGenesis(t *testing.T) {
	k := newTestKeeper(t)

	// init a community fund with 1000 UMEE and 10 ATOM available for funding
	k.initCommunityFund(
		coin.New(umee, 1000_000000),
		coin.New(atom, 10_000000),
	)

	// init a third party sponsor account
	sponsor := k.newAccount(
		coin.New(umee, 1000_000000),
	)

	// init a supplier with bonded uTokens
	alice := k.newBondedAccount(
		coin.New(u_umee, 100_000000),
		coin.New(u_atom, 50_000000),
	)
	// create some in-progress unbondings
	k.advanceTimeTo(90)
	k.mustBeginUnbond(alice, coin.New(u_umee, 5_000000))
	k.mustBeginUnbond(alice, coin.New(u_umee, 5_000000))
	k.mustBeginUnbond(alice, coin.New(u_atom, 5_000000))

	// create three separate programs, designed to be upcoming, ongoing, and completed at t=100
	k.addIncentiveProgram(u_umee, 10, 20, sdk.NewInt64Coin(umee, 10_000000), true)
	k.addIncentiveProgram(u_umee, 90, 20, sdk.NewInt64Coin(umee, 10_000000), false)
	k.addIncentiveProgram(u_umee, 140, 20, sdk.NewInt64Coin(umee, 10_000000), false)
	k.sponsor(sponsor, 3)

	// start programs and claim some rewards to set nonzero reward trackers
	k.advanceTimeTo(99)
	k.advanceTimeTo(100)
	k.mustClaim(alice)

	// get genesis state after this scenario
	gs1 := k.ExportGenesis(k.ctx)

	// require import-export idempotency on a fresh keeper
	k2 := newTestKeeper(t)
	k2.InitGenesis(k2.ctx, *gs1)
	gs2 := k2.ExportGenesis(k2.ctx)

	require.Equal(t, gs1, gs2, "genesis states equal")
}
