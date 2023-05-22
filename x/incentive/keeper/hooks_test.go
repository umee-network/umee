package keeper

import (
	"testing"

	"github.com/stretchr/testify/require"

	sdk "github.com/cosmos/cosmos-sdk/types"

	"github.com/umee-network/umee/v4/util/coin"
	"github.com/umee-network/umee/v4/x/incentive"
)

func TestHooks(t *testing.T) {
	k := newTestKeeper(t)
	require := require.New(t)

	// create a complex genesis state by running transactions
	alice := k.initScenario1()

	h := k.BondHooks()

	require.Equal(sdk.NewInt(100_000000), h.GetBonded(k.ctx, alice, u_umee), "initial restricted collateral")
	require.NoError(h.ForceUnbondTo(k.ctx, alice, coin.New(u_umee, 200_000000)), "liquidation unbond with no effect")
	require.Equal(sdk.NewInt(100_000000), h.GetBonded(k.ctx, alice, u_umee), "unchanged restricted collateral")

	// verify scenario 1 state is still unchanged by liquidation
	bonded, unbonding, unbondings := k.BondSummary(k.ctx, alice, u_umee)
	require.Equal(coin.New(u_umee, 90_000000), bonded)
	require.Equal(coin.New(u_umee, 10_000000), unbonding)
	require.Equal([]incentive.Unbonding{
		incentive.NewUnbonding(90, 86490, coin.New(u_umee, 5_000000)),
		incentive.NewUnbonding(90, 86490, coin.New(u_umee, 5_000000)),
	}, unbondings)

	// reduce a single in-progress unbonding by liquidation
	require.NoError(h.ForceUnbondTo(k.ctx, alice, coin.New(u_umee, 96_000000)), "liquidation unbond 1")
	require.Equal(sdk.NewInt(96_000000), h.GetBonded(k.ctx, alice, u_umee))
	bonded, unbonding, unbondings = k.BondSummary(k.ctx, alice, u_umee)
	require.Equal(coin.New(u_umee, 90_000000), bonded)
	require.Equal(coin.New(u_umee, 6_000000), unbonding)
	require.Equal([]incentive.Unbonding{
		incentive.NewUnbonding(90, 86490, coin.New(u_umee, 1_000000)),
		incentive.NewUnbonding(90, 86490, coin.New(u_umee, 5_000000)),
	}, unbondings)

	// reduce two in-progress unbondings by liquidation (one is ended altogether)
	require.NoError(h.ForceUnbondTo(k.ctx, alice, coin.New(u_umee, 92_000000)), "liquidation unbond 2")
	require.Equal(sdk.NewInt(92_000000), h.GetBonded(k.ctx, alice, u_umee))
	bonded, unbonding, unbondings = k.BondSummary(k.ctx, alice, u_umee)
	require.Equal(coin.New(u_umee, 90_000000), bonded)
	require.Equal(coin.New(u_umee, 2_000000), unbonding)
	require.Equal([]incentive.Unbonding{
		incentive.NewUnbonding(90, 86490, coin.New(u_umee, 2_000000)),
	}, unbondings)

	// end all unbondings and reduce bonded amount by liquidation
	require.NoError(h.ForceUnbondTo(k.ctx, alice, coin.New(u_umee, 46_000000)), "liquidation unbond 3")
	require.Equal(sdk.NewInt(46_000000), h.GetBonded(k.ctx, alice, u_umee))
	bonded, unbonding, unbondings = k.BondSummary(k.ctx, alice, u_umee)
	require.Equal(coin.New(u_umee, 46_000000), bonded)
	require.Equal(coin.Zero(u_umee), unbonding)
	require.Equal([]incentive.Unbonding{}, unbondings)

	// clear bonds by liquidation
	require.NoError(h.ForceUnbondTo(k.ctx, alice, coin.Zero(u_umee)), "liquidation unbond to zero")
	require.Equal(sdk.ZeroInt(), h.GetBonded(k.ctx, alice, u_umee))
	bonded, unbonding, unbondings = k.BondSummary(k.ctx, alice, u_umee)
	require.Equal(coin.Zero(u_umee), bonded)
	require.Equal(coin.Zero(u_umee), unbonding)
	require.Equal([]incentive.Unbonding{}, unbondings)
}
